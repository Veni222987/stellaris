package svc

import (
	"context"
	"fmt"
	"time"

	mqttc "github.com/eclipse/paho.mqtt.golang"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/stellaris/stellaris/server/internal/config"
	"github.com/stellaris/stellaris/server/internal/model"
	cmqtt "github.com/stellaris/stellaris/server/internal/mqtt"
	"github.com/stellaris/stellaris/server/internal/scheduler"
	"github.com/stellaris/stellaris/server/internal/ticket"
	"github.com/stellaris/stellaris/server/internal/ws"
	"golang.org/x/crypto/bcrypt"
)

// ServiceContext 聚合所有依赖：基础设施（DB / Redis / MQTT broker）+ Model 仓库 + 业务组件。
type ServiceContext struct {
	Config config.Config
	DB     *pgxpool.Pool
	Redis  *redis.Client
	MQTT   mqttc.Client

	Users       *model.UserModel
	Galaxies    *model.GalaxyModel
	Planets     *model.PlanetModel
	Agents      *model.AgentModel
	Sessions    *model.SessionModel
	Messages    *model.MessageModel
	Tasks       *model.TaskModel
	TaskChunks  *model.TaskChunkModel

	Publisher    *cmqtt.Publisher
	Subscriber   *cmqtt.Subscriber
	Scheduler    *scheduler.Scheduler
	Orchestrator *scheduler.Orchestrator
	Relay        *scheduler.Relay
	WSHub        *ws.Hub
	WSTickets    *ticket.Store
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := mustNewDB(c)
	mqttClient := mustNewMQTT(c)

	users := model.NewUserModel(db)
	seedAdmin(users, c)
	galaxies := model.NewGalaxyModel(db)
	planets := model.NewPlanetModel(db)
	agents := model.NewAgentModel(db)
	sessions := model.NewSessionModel(db)
	messages := model.NewMessageModel(db)
	tasks := model.NewTaskModel(db)
	chunks := model.NewTaskChunkModel(db)

	publisher := cmqtt.NewPublisher(mqttClient)
	hub := ws.New()
	sch := scheduler.New(agents, planets, tasks, publisher)
	orch := scheduler.NewOrchestrator(sessions, messages, tasks, agents, sch)
	relay := scheduler.NewRelay(sessions, messages, tasks, sch)
	subscriber := cmqtt.NewSubscriber(mqttClient, tasks, chunks, hub, relay, orch)

	return &ServiceContext{
		Config:       c,
		DB:           db,
		Redis:        newRedis(c),
		MQTT:         mqttClient,
		Users:        users,
		Galaxies:     galaxies,
		Planets:      planets,
		Agents:       agents,
		Sessions:     sessions,
		Messages:     messages,
		Tasks:        tasks,
		TaskChunks:   chunks,
		Publisher:    publisher,
		Subscriber:   subscriber,
		Scheduler:    sch,
		Orchestrator: orch,
		Relay:        relay,
		WSHub:        hub,
		WSTickets:    ticket.NewStore(5 * time.Minute),
	}
}

// seedAdmin 把 .env / 配置里的单一管理员账号写入 users 表（按 email upsert）。
// 全局只此一个账号，注册接口已移除。
func seedAdmin(users *model.UserModel, c config.Config) {
	if c.Admin.Email == "" || c.Admin.Password == "" {
		panic("Admin.Email / Admin.Password 未配置（ADMIN_EMAIL / ADMIN_PASSWORD）")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(c.Admin.Password), bcrypt.DefaultCost)
	if err != nil {
		panic(fmt.Errorf("hash admin password: %w", err))
	}
	if _, err := users.Upsert(context.Background(), c.Admin.Email, string(hash)); err != nil {
		panic(fmt.Errorf("seed admin user: %w", err))
	}
}

func mustNewDB(c config.Config) *pgxpool.Pool {
	pool, err := pgxpool.New(context.Background(), c.Postgres.DataSource)
	if err != nil {
		panic(fmt.Errorf("connect postgres: %w", err))
	}
	return pool
}

func newRedis(c config.Config) *redis.Client {
	return redis.NewClient(&redis.Options{Addr: c.Redis.Host})
}

func mustNewMQTT(c config.Config) mqttc.Client {
	opts := mqttc.NewClientOptions().
		AddBroker(c.MQTT.Broker).
		SetClientID(c.MQTT.ClientID).
		SetAutoReconnect(true).
		SetConnectRetry(true)
	client := mqttc.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(fmt.Errorf("connect mqtt: %w", token.Error()))
	}
	return client
}

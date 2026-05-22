package model

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// OpenTestDB 返回连接到本地 docker-compose 起的 postgres 的连接池。
// 调用前会 TRUNCATE 所有业务表，保证测试用例之间隔离。
// 通过环境变量 STELLARIS_TEST_DSN 覆盖连接串。
func OpenTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("STELLARIS_TEST_DSN")
	if dsn == "" {
		dsn = "postgres://stellaris:stellaris@localhost:5432/stellaris?sslmode=disable"
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("pgx connect: %v", err)
	}
	if _, err := pool.Exec(context.Background(),
		`TRUNCATE users, galaxies, planets, agents, sessions, messages, tasks, task_chunks RESTART IDENTITY CASCADE`); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

package config

import "github.com/zeromicro/go-zero/rest"

// Config 是调度中心的运行时配置。
// UserJwt / PlanetJwt 字段名必须与 .api 中 jwt: 引用的标识一致，
// 这样 go-zero 才能自动加载 JWT 中间件的 AccessSecret。
type Config struct {
	rest.RestConf

	UserJwt struct {
		AccessSecret string
		AccessExpire int64
	}

	PlanetJwt struct {
		AccessSecret string
		AccessExpire int64
	}

	// Admin 是全局共用的单一账号，账号密码来自 .env / 环境变量。
	// 启动时按 Email upsert 进 users 表，保证 galaxies.owner_user_id 外键有真实用户行。
	Admin struct {
		Email    string
		Password string
	}

	Postgres struct {
		DataSource string
	}

	Redis struct {
		Host string
	}

	MQTT struct {
		Broker   string
		ClientID string
	}
}

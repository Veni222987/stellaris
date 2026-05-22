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

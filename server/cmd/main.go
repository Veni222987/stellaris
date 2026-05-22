package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"regexp"

	"github.com/stellaris/stellaris/server/internal/config"
	"github.com/stellaris/stellaris/server/internal/handler"
	"github.com/stellaris/stellaris/server/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "server/etc/core.yaml", "config file path (default works when running from repo root)")

// envDefaultRe 匹配 ${VAR:-default} 形式，go-zero 自带 os.ExpandEnv 不支持 default 语法，这里自己展开。
var envDefaultRe = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*):-([^}]*)\}`)

func expandEnvWithDefaults(s string) string {
	return envDefaultRe.ReplaceAllStringFunc(s, func(m string) string {
		sub := envDefaultRe.FindStringSubmatch(m)
		if v, ok := os.LookupEnv(sub[1]); ok && v != "" {
			return v
		}
		return sub[2]
	})
}

func main() {
	flag.Parse()

	raw, err := os.ReadFile(*configFile)
	if err != nil {
		panic(fmt.Errorf("read config %s: %w", *configFile, err))
	}
	expanded := os.ExpandEnv(expandEnvWithDefaults(string(raw)))

	var c config.Config
	if err := conf.LoadFromYamlBytes([]byte(expanded), &c); err != nil {
		panic(fmt.Errorf("load config: %w", err))
	}

	ctx := svc.NewServiceContext(c)

	// 启动 MQTT subscriber 接收 Planet 回传的 result chunk
	if err := ctx.Subscriber.Start(context.Background()); err != nil {
		panic(fmt.Errorf("start mqtt subscriber: %w", err))
	}

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Stellaris Core 启动于 %s:%d\n", c.Host, c.Port)
	server.Start()
}

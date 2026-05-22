// Package config 管理本机 stellaris-cli 的持久化配置（加入星系后存的 token / planet uuid 等）。
package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	CoreAddr   string `yaml:"core_addr"`    // http://ip:port
	GID        string `yaml:"gid"`
	PlanetUUID string `yaml:"planet_uuid"`
	PlanetJwt  string `yaml:"planet_jwt"`
	MQTTBroker string `yaml:"mqtt_broker"`  // tcp://host:1883
	AgentsDir  string `yaml:"agents_dir,omitempty"` // 缺省时 daemon 落到 /etc/stellaris/agents.d
}

func dir() string {
	if d := os.Getenv("STELLARIS_CONFIG_DIR"); d != "" {
		return d
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".stellaris")
}

func path() string { return filepath.Join(dir(), "config.yaml") }

func Save(c *Config) error {
	if err := os.MkdirAll(dir(), 0o700); err != nil {
		return err
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path(), data, 0o600)
}

func Load() (*Config, error) {
	data, err := os.ReadFile(path())
	if errors.Is(err, os.ErrNotExist) {
		return nil, errors.New("尚未加入任何星系，请先执行 `stellaris-cli orbit ...`")
	}
	if err != nil {
		return nil, err
	}
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

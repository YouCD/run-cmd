package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type MQTTConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Broker   string `yaml:"broker"`
	Topic    string `yaml:"topic"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type DingTalkConfig struct {
	Enabled bool   `yaml:"enabled"`
	Webhook string `yaml:"webhook"`
}

type NtfyConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Server   string `yaml:"server"`
	Topic    string `yaml:"topic"`
	Priority int    `yaml:"priority"`
}

type Config struct {
	MQTT     MQTTConfig     `yaml:"mqtt"`
	DingTalk DingTalkConfig `yaml:"dingtalk"`
	Ntfy     NtfyConfig     `yaml:"ntfy"`
}

func DefaultConfig() *Config {
	return &Config{
		MQTT: MQTTConfig{
			Enabled:  false,
			Broker:   "wss://mq-client.youcd.online/mqtt",
			Topic:    "sms",
			Username: "",
			Password: "",
		},
		DingTalk: DingTalkConfig{
			Enabled: false,
			Webhook: "",
		},
		Ntfy: NtfyConfig{
			Enabled:  false,
			Server:   "https://ntfy.sh",
			Topic:    "",
			Priority: 3,
		},
	}
}

func configPath() string {
	if p := os.Getenv("RUN_CMD_CONFIG"); p != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".run-cmd.yaml"
	}
	return filepath.Join(home, ".config", "run_cmd", "config.yaml")
}

func LoadConfig(path string) (*Config, error) {
	if path == "" {
		path = configPath()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}
	return &cfg, nil
}

func InitConfig(path string) error {
	if path == "" {
		path = configPath()
	}

	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("配置文件已存在: %s", path)
	}

	cfg := DefaultConfig()
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("生成配置文件失败: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	fmt.Printf("配置文件已创建: %s\n", path)
	return nil
}

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Token   string            `mapstructure:"token"`
	Queries map[string]string `mapstructure:"queries"`
	DataDir string            `mapstructure:"data_dir"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.ibkr")

	viper.SetDefault("data_dir", "./data")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	if cfg.Token == "" {
		return nil, fmt.Errorf("配置缺少 token")
	}
	if len(cfg.Queries) == 0 {
		return nil, fmt.Errorf("配置缺少 queries")
	}

	// 确保数据目录存在
	absDir, _ := filepath.Abs(cfg.DataDir)
	cfg.DataDir = absDir
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %w", err)
	}

	return &cfg, nil
}

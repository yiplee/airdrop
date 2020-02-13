package config

import (
	"github.com/fox-one/pkg/config"
	"github.com/fox-one/pkg/store/db"
)

type (
	Config struct {
		DB     db.Config `json:"db,omitempty"`
		Wallet Wallet    `json:"wallet,omitempty"`
		Task   Task      `json:"task,omitempty"`
	}

	Wallet struct {
		Endpoint     string `json:"endpoint,omitempty"`
		BrokerID     string `json:"broker_id,omitempty"`
		BrokerSecret string `json:"broker_secret,omitempty"`
		PinSecret    string `json:"pin_secret,omitempty"`
		UserID       string `json:"user_id,omitempty"`
		Pin          string `json:"pin,omitempty"`
	}

	Task struct {
		MaxTargets int `json:"max_targets,omitempty"`
	}
)

func Load(configPath string, cfg *Config) error {
	if err := config.LoadYaml(configPath, &cfg); err != nil {
		return err
	}

	defaultConfig(cfg)
	return nil
}

func defaultConfig(cfg *Config) {
	if cfg.Wallet.Endpoint == "" {
		cfg.Wallet.Endpoint = "https://wallet.fox.one"
	}

	if cfg.Task.MaxTargets == 0 {
		cfg.Task.MaxTargets = 1000
	}
}

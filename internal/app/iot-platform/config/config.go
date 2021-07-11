package config

import (
	"github.com/saman2000hoseini/mossgow/pkg/config"
	"github.com/sirupsen/logrus"
	"gopkg.in/go-playground/validator.v9"
	"time"
)

const (
	app       = "iot-platform"
	cfgFile   = "config.yaml"
	cfgPrefix = "iot-platform"
)

type (
	Config struct {
		CentralServer Server   `mapstructure:"central-server"`
		LocalServer   Server   `mapstructure:"local-server"`
		Cooler        Actuator `mapstructure:"cooler"`
		LightBulb     Actuator `mapstructure:"light-bulb"`
		Temperature   Sensor   `mapstructure:"temperature"`
		Light         Sensor   `mapstructure:"light"`
		JWT           JWT      `mapstructure:"jwt"`
	}

	Server struct {
		Address         string        `mapstructure:"address"`
		GracefulTimeout time.Duration `mapstructure:"graceful-timeout"`
		ReadTimeout     time.Duration `mapstructure:"read-timeout"`
		WriteTimeout    time.Duration `mapstructure:"write-timeout"`
	}

	Actuator struct {
		Type            int           `mapstructure:"type"`
		Address         string        `mapstructure:"address"`
		GracefulTimeout time.Duration `mapstructure:"graceful-timeout"`
		ReadTimeout     time.Duration `mapstructure:"read-timeout"`
		WriteTimeout    time.Duration `mapstructure:"write-timeout"`
	}

	Sensor struct {
		Type               int           `mapstructure:"type"`
		LocalServerAddress string        `mapstructure:"local-server-address"`
		ReadTimeout        time.Duration `mapstructure:"read-timeout"`
	}

	JWT struct {
		Expiration time.Duration `mapstructure:"expiration"`
		Secret     string        `mapstructure:"secret"`
	}
)

func (c Config) Validate() error {
	return validator.New().Struct(c)
}

// Init initializes application configuration.
func Init() Config {
	var cfg Config

	config.Init(app, cfgFile, &cfg, defaultConfig, cfgPrefix)

	if err := cfg.Validate(); err != nil {
		logrus.Fatalf("failed to validate configurations: %s", err.Error())
	}

	return cfg
}

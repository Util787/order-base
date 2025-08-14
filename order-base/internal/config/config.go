package config

import (
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

const (
	EnvLocal = "local"
	EnvDev   = "dev"
	EnvProd  = "prod"
)

type Config struct {
	Env              string        `yaml:"env"`
	ShutdownTimeout  time.Duration `yaml:"shutdown-timeout"`
	PostgresConfig   `yaml:"postgres"`
	HTTPServerConfig `yaml:"http-server"`
	KafkaConfig      `yaml:"kafka"`
}

type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	DbName   string `yaml:"db-name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`

	MaxConns        int           `yaml:"max-conns"`
	ConnMaxLifetime time.Duration `yaml:"conn-max-lifetime"`
	ConnMaxIdleTime time.Duration `yaml:"conn-max-idle-time"`
}

type HTTPServerConfig struct {
	Host              string        `yaml:"host"`
	Port              int           `yaml:"port"`
	ReadHeaderTimeout time.Duration `yaml:"read-header-timeout"`
	WriteTimeout      time.Duration `yaml:"write-timeout"`
	ReadTimeout       time.Duration `yaml:"read-timeout"`
}

type KafkaConfig struct {
	Brokers []string      `yaml:"brokers"`
	Topic   string        `yaml:"topic"`
	GroupID string        `yaml:"group-id"`
	MaxWait time.Duration `yaml:"max-wait"`
}

// MustLoadConfig expects CONFIG_PATH environment variable to be already set
func MustLoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		panic("Failed to load .env file")
	}

	path := os.Getenv("CONFIG_PATH")

	if _, err = os.Stat(path); os.IsNotExist(err) {
		panic("Config path does not exist")
	}

	var cfg Config

	err = cleanenv.ReadConfig(path, &cfg)
	if err != nil {
		panic("Error reading config: " + err.Error())
	}

	if cfg.Env == "" {
		panic("Env is not set")
	}

	if cfg.Env != EnvLocal && cfg.Env != EnvDev && cfg.Env != EnvProd {
		panic("Invalid env")
	}

	return &cfg
}

package config

import (
	"fmt"
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
	Env              string        `yaml:"env" env:"ENV"`
	ShutdownTimeout  time.Duration `yaml:"shutdown-timeout" env:"SHUTDOWN_TIMEOUT"`
	PostgresConfig   `yaml:"postgres"`
	HTTPServerConfig `yaml:"http-server"`
	KafkaConfig      `yaml:"kafka"`
}

type PostgresConfig struct {
	Host     string `yaml:"host" env:"POSTGRES_HOST"`
	Port     int    `yaml:"port" env:"POSTGRES_PORT"`
	DbName   string `yaml:"db-name" env:"POSTGRES_DB_NAME"`
	User     string `yaml:"user" env:"POSTGRES_USER"`
	Password string `yaml:"password" env:"POSTGRES_PASSWORD"`

	MaxConns        int           `yaml:"max-conns" env:"POSTGRES_MAX_CONNS"`
	ConnMaxLifetime time.Duration `yaml:"conn-max-lifetime" env:"POSTGRES_CONN_MAX_LIFETIME"`
	ConnMaxIdleTime time.Duration `yaml:"conn-max-idle-time" env:"POSTGRES_CONN_MAX_IDLE_TIME"`
}

type HTTPServerConfig struct {
	Host              string        `yaml:"host" env:"HTTP_SERVER_HOST"`
	Port              int           `yaml:"port" env:"HTTP_SERVER_PORT"`
	ReadHeaderTimeout time.Duration `yaml:"read-header-timeout" env:"HTTP_SERVER_READ_HEADER_TIMEOUT"`
	WriteTimeout      time.Duration `yaml:"write-timeout" env:"HTTP_SERVER_WRITE_TIMEOUT"`
	ReadTimeout       time.Duration `yaml:"read-timeout" env:"HTTP_SERVER_READ_TIMEOUT"`
}

type KafkaConfig struct {
	Brokers []string      `yaml:"brokers" env:"KAFKA_BROKERS"`
	Topic   string        `yaml:"topic" env:"KAFKA_TOPIC"`
	GroupID string        `yaml:"group-id" env:"KAFKA_GROUP_ID"`
	MaxWait time.Duration `yaml:"max-wait" env:"KAFKA_MAX_WAIT"`
}

// If CONFIG_PATH env variable is set it will load from yaml, if not it will load from env
func MustLoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		panic("Failed to load .env file: " + err.Error())
	}

	path := os.Getenv("CONFIG_PATH")

	var cfg Config

	if path == "" {
		fmt.Println("CONFIG_PATH is not set: load from .env")
		err = cleanenv.ReadEnv(&cfg)
		if err != nil {
			panic("Error reading env: " + err.Error())
		}
	} else {
		fmt.Println("CONFIG_PATH is set: load from .yaml")
		configFile, err := os.Open(path)
		if err != nil {
			panic("Error opening config file: " + err.Error())
		}
		defer configFile.Close()
		err = cleanenv.ParseYAML(configFile, &cfg)
		if err != nil {
			panic("Error reading config: " + err.Error())
		}
	}

	if cfg.Env == "" {
		panic("Env is not set")
	}

	if cfg.Env != EnvLocal && cfg.Env != EnvDev && cfg.Env != EnvProd {
		panic("Invalid env")
	}

	return &cfg
}

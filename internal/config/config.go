package config

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type PostgresConfig struct {
	Host     string `yaml:"host" env-required:"true"`
	Port     string `yaml:"port" env-required:"true"`
	User     string `yaml:"user" env-required:"true"`
	Password string `yaml:"password" env-required:"true"`
	DBName   string `yaml:"dbname" env-required:"true"`
	SSLMode  string `yaml:"sslmode" env-default:"disable"`
}

type Config struct {
	Env       string `yaml:"env" env-default:"local"`
	Host      string `yaml:"bot_host" env-required:"true"`
	BatchSize int    `yaml:"batch_size" env-default:"100"`
	BotToken  string `yaml:"bot_token" env-required:"true"`

	Postgres PostgresConfig `yaml:"postgres" env-required:"true"`
}

func MustLoad() *Config {
	configPath := "./config/local.yaml"

	// проверка существования файла
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", configPath)
	}

	return &cfg
}

package config

import (
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
)

type Config struct {
	Env      string         `yaml:"env" env-default:"local"`
	LogLevel string         `yaml:"log_level" env-default:"INFO"`
	Server   ServerConfig   `yaml:"server"`
	Postgres PostgresConfig `yaml:"postgres"`
	Gmail    GmailConfig    `yaml:"gmail"`
}
type ServerConfig struct {
	Host string `env-default:"localhost"`
	Port int    `env-default:"8080"`
}
type PostgresConfig struct {
	Host     string `env-default:"localhost"`
	Port     int    `env-default:"5432"`
	User     string `env-default:"postgres"`
	Password string
	Database string `env-default:"SocialManagerDB"`
	SSLMode  string `env-default:"disable"`
}

type GmailConfig struct {
	SecretPath string `env-default:"secret.json"`
}

func MustLoad() *Config {
	configPath := fetchConfigPath()
	if configPath == "" {
		panic("config is empty")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("config path if empty " + err.Error())
	}

	return &cfg
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}

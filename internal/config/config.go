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
	Secrets  SecretsConfig  `yaml:"secrets"`
	OAuth2   OAuth2Config   `yaml:"oauth2"`
}

type ServerConfig struct {
	Host string `yaml:"host" env-default:"localhost"`
	Port int    `yaml:"port" env-default:"8080"`
}

type PostgresConfig struct {
	Host     string `yaml:"host" env-default:"localhost"`
	Port     int    `yaml:"port" env-default:"5432"`
	User     string `yaml:"user" env-default:"postgres"`
	Password string `yaml:"password,omitempty"`
	Database string `yaml:"database" env-default:"SocialManagerDB"`
	SSLMode  string `yaml:"SSLMode" env-default:"disable"`
}

type GmailConfig struct {
	RefreshTime string `yaml:"refreshTime" env-default:"5m"`
}

type OAuth2Config struct {
	CredentialPath string `yaml:"credentialPath" env-default:"config/secretGmail.json"`
}

type SecretsConfig struct {
	JWTSecret string `yaml:"jwtSecret" env-default:"SECRET_KEY"`
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
		panic("config path is empty " + err.Error())
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

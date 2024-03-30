package config

import (
	"flag"
	"os"
)

type Config struct {
	Env      string         `yaml:"env" env-default:"local"`
	Postgres PostgresConfig `yaml:"postgres"`
	Redis    RedisConfig    `yaml:"redis"`
}

type PostgresConfig struct {
	Host     string `env-default:"localhost"`
	Port     int    `env-default:"5432"`
	User     string `env-default:"postgres"`
	Password string
	Database string `env-default:"SocialManagerDB"`
	SSLMode  string `env-default:"disable"`
}

type RedisConfig struct {
	Host     string `env-default:"localhost"`
	Port     int    `env-default:"6379"`
	Username string `env-default:"default"`
	Password string `env-default:""`
	Database int    `env-default:"0"`
}

func MustLoad() *Config {
	configpath, err := fetchConfigPath()
	if err != nil {

	}
}

func fetchConfigPath() (string, error) {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res, nil
}

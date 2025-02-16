package config

import (
	"flag"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
)

type Config struct {
	Env      string     `yaml:"env" env-default:"local"`
	AmqpConf AmqpConfig `yaml:"amqp"`
}

type AmqpConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	UserName     string `yaml:"user_name"`
	UserPass     string `yaml:"user_pass"`
	QueueName    string `yaml:"queue"`
	ExchangeName string `yaml:"exchange"`
	RoutingKey   string `yaml:"routing_key"`
}

func (r AmqpConfig) GetAmqpUri() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d/", r.UserName, r.UserPass, r.Host, r.Port)
}

func MustLoad() *Config {
	configPath := fetchConfigPath()
	if configPath == "" {
		panic("config path is empty")
	}

	// check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("config path is empty: " + err.Error())
	}

	return &cfg
}

// fetchConfigPath fetches config path from command line flag or environment variable.
// Priority: flag > env > default.
// Default value is empty string.
func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}

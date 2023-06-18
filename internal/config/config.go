package config

import (
	"os"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
)

const configFilePath = "config/config.yml"

type Config struct {
	AppEnv        string `yaml:"app_env" envconfig:"APP_ENV"`
	TelegramToken string `yaml:"telegram_token" envconfig:"TELEGRAM_TOKEN"`
}

func New() *Config {
	return &Config{}
}

func (c *Config) ParseConfig() error {
	if err := c.parseFile(); err != nil {
		return err
	}
	if err := c.parseEnv(); err != nil {
		return err
	}
	return nil
}

func (c *Config) parseFile() error {
	f, err := os.Open(configFilePath)
	if err != nil {
		return err
	}
	defer f.Close()
	decoder := yaml.NewDecoder(f)
	return decoder.Decode(c)
}

func (c *Config) parseEnv() error {
	return envconfig.Process("", c)
}

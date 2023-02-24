package config

import (
	"github.com/caarlos0/env/v7"
	"github.com/joomcode/errorx"
)

type Config struct {
	Token string `env:"TELEGRAM_API_TOKEN,required"`
}

func ReadConfig() (*Config, error) {
	result := &Config{}
	err := env.Parse(result)
	if err != nil {
		return nil, errorx.EnhanceStackTrace(err, "failed to read config")
	}

	return result, nil
}

package config

import (
	"github.com/caarlos0/env/v7"
	"github.com/goioc/di"
	"github.com/joomcode/errorx"
	"github.com/samber/lo"
)

type Config struct {
	Token          string `env:"TELEGRAM_API_TOKEN,required"`
	OurSpbClientId string `env:"OURSPB_CLIENT_ID,required"`
	OurSpbSecret   string `env:"OURSPB_SECRET,required"`
}

const (
	BeanId = "Config"
)

func RegisterBean() *Config {
	config := lo.Must(readConfig())
	_ = lo.Must(di.RegisterBeanInstance(BeanId, &config))
	return config
}

func readConfig() (*Config, error) {
	result := &Config{}
	err := env.Parse(result)
	if err != nil {
		return nil, errorx.EnhanceStackTrace(err, "failed to read config")
	}

	return result, nil
}

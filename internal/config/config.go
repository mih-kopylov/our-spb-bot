package config

import (
	"github.com/caarlos0/env/v7"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"time"
)

type Config struct {
	TelegramApiToken       string        `env:"TELEGRAM_API_TOKEN,required"`
	TelegramApiEndpoint    string        `env:"TELEGRAM_API_ENDPOINT"`
	OurSpbClientId         string        `env:"OURSPB_CLIENT_ID,required"`
	OurSpbSecret           string        `env:"OURSPB_SECRET,required"`
	FirebaseServiceAccount string        `env:"FIREBASE_SERVICE_ACCOUNT,required"`
	SleepDuration          time.Duration `env:"SLEEP_DURATION,required"`
}

func NewConfig() (*Config, error) {
	result := &Config{}
	err := env.Parse(result)
	if err != nil {
		return nil, errorx.EnhanceStackTrace(err, "failed to read config")
	}

	if result.TelegramApiEndpoint == "" {
		result.TelegramApiEndpoint = tgbotapi.APIEndpoint
	}

	return result, nil
}

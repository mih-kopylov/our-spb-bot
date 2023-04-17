package api

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/config"
	"github.com/sirupsen/logrus"
)

func NewApi(conf *config.Config) (*tgbotapi.BotAPI, error) {
	err := tgbotapi.SetLogger(&LorRusLogger{})
	if err != nil {
		return nil, errorx.EnhanceStackTrace(err, "failed to configure bot api logging")
	}

	api, err := tgbotapi.NewBotAPIWithAPIEndpoint(conf.TelegramApiToken, conf.TelegramApiEndpoint)
	if err != nil {
		return nil, errorx.EnhanceStackTrace(err, "failed to create bot")
	}

	api.Debug = true
	return api, nil
}

type LorRusLogger struct{}

func (l *LorRusLogger) Println(v ...interface{}) {
	logrus.Infoln(v...)
}

func (l *LorRusLogger) Printf(format string, v ...interface{}) {
	logrus.Infof(format, v...)
}

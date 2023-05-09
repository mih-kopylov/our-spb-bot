package api

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/config"
	"go.uber.org/zap"
)

func NewApi(logger *zap.Logger, conf *config.Config) (*tgbotapi.BotAPI, error) {
	err := tgbotapi.SetLogger(&TgLogger{logger: logger})
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

type TgLogger struct {
	logger *zap.Logger
}

func (l *TgLogger) Println(v ...interface{}) {
	l.logger.Sugar().Info(v...)
}

func (l *TgLogger) Printf(format string, v ...interface{}) {
	l.logger.Sugar().Infof(format, v...)
}

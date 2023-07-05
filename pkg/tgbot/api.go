package tgbot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

func NewApi(logger *zap.Logger, params ApiParams) (*tgbotapi.BotAPI, error) {
	err := tgbotapi.SetLogger(&Logger{logger: logger})
	if err != nil {
		return nil, ErrFailedToInitializeBot.Wrap(err, "failed to configure bot api logging")
	}

	api, err := tgbotapi.NewBotAPIWithAPIEndpoint(params.TelegramApiToken, params.TelegramApiEndpoint)
	if err != nil {
		return nil, ErrFailedToInitializeBot.Wrap(err, "failed to create bot api")
	}

	api.Debug = params.Debug
	return api, nil
}

type ApiParams struct {
	TelegramApiToken    string
	TelegramApiEndpoint string
	Debug               bool
}

type Logger struct {
	logger *zap.Logger
}

func (l *Logger) Println(v ...interface{}) {
	l.logger.Sugar().Info(v...)
}

func (l *Logger) Printf(format string, v ...interface{}) {
	l.logger.Sugar().Infof(format, v...)
}

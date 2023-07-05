package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

func NewApi(logger *zap.Logger, params TgApiParams) (*tgbotapi.BotAPI, error) {
	err := tgbotapi.SetLogger(&TgLogger{logger: logger})
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

type TgApiParams struct {
	TelegramApiToken    string
	TelegramApiEndpoint string
	Debug               bool
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

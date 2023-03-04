package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Command interface {
	Name() string
	Description() string
	Handle(message *tgbotapi.Message) error
	Callback(callbackQuery *tgbotapi.CallbackQuery, data string) error
}

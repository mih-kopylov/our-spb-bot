package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Form interface {
	Name() string
	Handle(message *tgbotapi.Message) error
}

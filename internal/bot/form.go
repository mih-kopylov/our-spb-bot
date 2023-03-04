package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Form interface {
	Handle(message *tgbotapi.Message) error
}

package handler

import (
	_ "embed"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
)

//go:embed start.html
var startText string

func StartHandlerApi(bot *tgbotapi.BotAPI, message *tgbotapi.Message, states *state.States) error {
	_, err := states.NewStateIfNotExists(message.From.ID)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to crate user state")
	}

	reply := tgbotapi.NewMessage(message.Chat.ID, startText)
	reply.ParseMode = tgbotapi.ModeHTML
	reply.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)

	_, err = bot.Send(reply)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to send reply")
	}

	return nil
}

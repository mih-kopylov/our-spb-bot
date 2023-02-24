package handler

import (
	_ "embed"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
)

func LoginCommandHandler(bot *tgbotapi.BotAPI, message *tgbotapi.Message, states *state.States) error {
	userState, err := states.GetState(message.Chat.ID)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to get user state")
	}

	userState.MessageHandlerName = LoginFormHandlerName
	userState.Credentials = &state.Credentials{}

	reply := tgbotapi.NewMessage(message.Chat.ID, "Введите логин")
	_, err = bot.Send(reply)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to send reply")
	}

	return nil
}

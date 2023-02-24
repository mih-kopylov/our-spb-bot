package handler

import (
	_ "embed"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
)

func StatusHandlerApi(bot *tgbotapi.BotAPI, message *tgbotapi.Message, states *state.States) error {
	userState, err := states.GetState(message.From.ID)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to get user state")
	}

	reply := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf(`
Пользователь: @%v
Сообщений отправлено: %v
Сообщений в очереди: %v
`, message.From.UserName, userState.SentCount, len(userState.Queue)))
	_, err = bot.Send(reply)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to send reply")
	}

	return nil
}

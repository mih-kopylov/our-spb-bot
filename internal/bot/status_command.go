package bot

import (
	_ "embed"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/queue"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
	"github.com/samber/lo"
	"strings"
	"time"
)

const (
	StatusCommandName = "status"
)

type StatusCommand struct {
	states       state.States       `di.inject:"States"`
	tgbot        *TgBot             `di.inject:"TgBot"`
	messageQueue queue.MessageQueue `di.inject:"Queue"`
}

func (c *StatusCommand) Name() string {
	return StatusCommandName
}

func (c *StatusCommand) Description() string {
	return "Статус обращений"
}

func (c *StatusCommand) Handle(message *tgbotapi.Message) error {
	userState, err := c.states.GetState(message.Chat.ID)
	if err != nil {
		if errorx.IsOfType(err, state.ErrRateLimited) {
			err = c.tgbot.SendMessage(message.Chat, "Превышен лимит подключений к базе данных")
			if err != nil {
				return errorx.EnhanceStackTrace(err, "failed to send reply")
			}

			return nil
		}

		return errorx.EnhanceStackTrace(err, "failed to get user state")
	}

	messagesCount, err := c.messageQueue.UserMessagesCount(userState.UserId)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to count messages in the queue")
	}

	var accounts string
	if len(userState.Accounts) == 0 {
		accounts = "нет"
	} else {
		accounts = strings.Join(lo.Map(userState.Accounts, func(item state.Account, index int) string {
			result := "  " + item.Login
			if item.RateLimitedUntil.After(time.Now()) {
				result += " - заблокирован до " + item.RateLimitedUntil.Format(time.RFC3339)
			} else {
				result += " - готов к отправке обращений"
			}
			return result
		}), "\n")
	}

	reply := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf(`
Пользователь: @%v
Аккаунты:
%v
Сообщений отправлено: %v
Ожидает отправки: %v
Не удалось отправить: %v
Ожидают авторизации: %v

/message - отправить новое обращение 
`,
		message.Chat.UserName,
		accounts,
		userState.SentMessagesCount,
		messagesCount[queue.StatusCreated],
		messagesCount[queue.StatusFailed],
		messagesCount[queue.StatusAwaitingAuthorization],
	))
	_, err = c.tgbot.api.Send(reply)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to send reply")
	}

	return nil
}

func (c *StatusCommand) Callback(_ *tgbotapi.CallbackQuery, _ string) error {
	return nil
}

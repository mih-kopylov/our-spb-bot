package bot

import (
	_ "embed"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/queue"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
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
		return errorx.EnhanceStackTrace(err, "failed to get user state")
	}

	messagesCount, err := c.messageQueue.UserMessagesCount(userState.UserId)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to count messages in the queue")
	}

	var authorizedInPortal string
	if userState.Token == "" {
		authorizedInPortal = "нет"
	} else {
		authorizedInPortal = userState.Login
	}

	reply := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf(`
Пользователь: @%v
Авторизован: %v
Сообщений отправлено: %v
Ожидает отправки: %v
Не удалось отправить: %v
Ожидают авторизации: %v

/message - отправить новое обращение 
`,
		message.Chat.UserName,
		authorizedInPortal,
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

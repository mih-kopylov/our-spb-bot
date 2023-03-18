package bot

import (
	_ "embed"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/queue"
	"time"
)

const (
	ResetStatusCommandName = "reset_status"
)

type ResetStatusCommand struct {
	tgbot        *TgBot             `di.inject:"TgBot"`
	messageQueue queue.MessageQueue `di.inject:"Queue"`
}

func (c *ResetStatusCommand) Name() string {
	return ResetStatusCommandName
}

func (c *ResetStatusCommand) Description() string {
	return "Сбросить статус ошибки у всех обращений"
}

func (c *ResetStatusCommand) Handle(message *tgbotapi.Message) error {
	counter := 0

	err := c.messageQueue.UpdateEachMessage(message.Chat.ID, func(message *queue.Message) {
		if message.Status == queue.StatusFailed {
			message.Tries = 0
			message.RetryAfter = time.Now()
			message.Status = queue.StatusCreated
			message.FailDescription = ""

			counter++
		}
	})
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to update each message")
	}

	reply := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf(`
Пользователь: @%v
Обновлено сообщений: %v

/status - статус обращений 

/message - отправить новое обращение 
`,
		message.Chat.UserName,
		counter,
	))
	_, err = c.tgbot.api.Send(reply)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to send reply")
	}

	return nil
}

func (c *ResetStatusCommand) Callback(_ *tgbotapi.CallbackQuery, _ string) error {
	return nil
}

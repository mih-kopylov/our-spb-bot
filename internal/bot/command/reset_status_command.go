package command

import (
	_ "embed"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/bot/service"
	"github.com/mih-kopylov/our-spb-bot/internal/queue"
	"github.com/mih-kopylov/our-spb-bot/pkg/bot"
	"time"
)

const (
	ResetStatusCommandName = "reset_status"
)

type ResetStatusCommand struct {
	service      *service.Service
	messageQueue queue.MessageQueue
}

func NewResetStatusCommand(service *service.Service, messageQueue queue.MessageQueue) bot.Command {
	return &ResetStatusCommand{
		service:      service,
		messageQueue: messageQueue,
	}
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

	reply := fmt.Sprintf(`
Пользователь: @%v
Обновлено сообщений: %v

/status - статус обращений 

/message - отправить новое обращение 
`,
		message.Chat.UserName,
		counter,
	)
	err = c.service.SendMessage(message.Chat, reply)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to send reply")
	}

	return nil
}

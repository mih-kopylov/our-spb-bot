package command

import (
	_ "embed"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/bot/callback"
	"github.com/mih-kopylov/our-spb-bot/internal/bot/form"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
	"github.com/mih-kopylov/our-spb-bot/pkg/tgbot"
)

const (
	MessageCommandName = "message"
)

type MessageCommand struct {
	stateManager            state.Manager
	service                 *tgbot.Service
	messageCategoryCallback *callback.MessageCategoryCallback
}

func NewMessageCommand(stateManager state.Manager, service *tgbot.Service, messageCategoryCallback *callback.MessageCategoryCallback) tgbot.Command {
	return &MessageCommand{
		stateManager:            stateManager,
		service:                 service,
		messageCategoryCallback: messageCategoryCallback,
	}
}

func (c *MessageCommand) Name() string {
	return MessageCommandName
}

func (c *MessageCommand) Description() string {
	return "Отправить обращение"
}

func (c *MessageCommand) Handle(message *tgbotapi.Message) error {
	userState, err := c.stateManager.GetState(message.Chat.ID)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to get user state")
	}

	if len(userState.Accounts) == 0 {
		return c.service.SendMessage(message.Chat, `Вы не авторизованы на портале.

Для того, чтобы отправить обращение, нужно авторизоваться на портале с логином и паролем. 

Используйте команду /login для этого.`)
	}

	userState.ClearForm()
	userState.MessageHandlerName = form.MessageFormName
	err = c.stateManager.SetState(userState)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to set user state")
	}

	_, err = c.service.SendMessageCustom(message.Chat, "Выберите категорию", func(reply *tgbotapi.MessageConfig) {
		reply.ReplyMarkup = c.messageCategoryCallback.CreateCategoriesReplyMarkup(userState)
	})
	return err
}

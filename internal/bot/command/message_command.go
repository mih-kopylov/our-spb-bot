package command

import (
	_ "embed"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/bot"
	"github.com/mih-kopylov/our-spb-bot/internal/bot/callback"
	"github.com/mih-kopylov/our-spb-bot/internal/bot/form"
	"github.com/mih-kopylov/our-spb-bot/internal/bot/service"
	"github.com/mih-kopylov/our-spb-bot/internal/category"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
)

const (
	MessageCommandName = "message"
)

type MessageCommand struct {
	states                  state.States
	service                 *service.Service
	cateogiresTree          *category.UserCategoryTreeNode
	messageCategoryCallback *callback.MessageCategoryCallback
}

func NewMessageCommand(states state.States, service *service.Service, cateogiresTree *category.UserCategoryTreeNode, messageCategoryCallback *callback.MessageCategoryCallback) bot.Command {
	return &MessageCommand{
		states:                  states,
		service:                 service,
		cateogiresTree:          cateogiresTree,
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
	userState, err := c.states.GetState(message.Chat.ID)
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
	err = c.states.SetState(userState)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to set user state")
	}

	_, err = c.service.SendMessageCustom(message.Chat, "Выберите категорию", func(reply *tgbotapi.MessageConfig) {
		reply.ReplyMarkup = c.messageCategoryCallback.CreateCategoriesReplyMarkup(userState)
	})
	return err
}

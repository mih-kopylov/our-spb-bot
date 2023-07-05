package command

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/bot/form"
	"github.com/mih-kopylov/our-spb-bot/internal/bot/service"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
	"github.com/mih-kopylov/our-spb-bot/pkg/bot"
)

const (
	FileIdCommandName = "file_id"
)

type FileIdCommand struct {
	states  state.States
	service *service.Service
}

func NewFileIdCommand(states state.States, service *service.Service) bot.Command {
	return &FileIdCommand{
		states:  states,
		service: service,
	}
}

func (c *FileIdCommand) Name() string {
	return FileIdCommandName
}

func (c *FileIdCommand) Description() string {
	return "Узнать идентификатор фото"
}

func (c *FileIdCommand) Handle(message *tgbotapi.Message) error {
	userState, err := c.states.GetState(message.Chat.ID)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to get user state")
	}

	userState.MessageHandlerName = form.FileIdFormName
	err = c.states.SetState(userState)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to set user state")
	}

	return c.service.SendMessage(message.Chat, `Отправляйте файлы, в ответ я напишу их идентификатор.

Если написать идентификатор, то я пришлю ссылку на скачивание фото.`)
}

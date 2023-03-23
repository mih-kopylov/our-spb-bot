package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
)

const (
	FileIdCommandName = "file_id"
)

type FileIdCommand struct {
	states state.States `di.inject:"States"`
	tgbot  *TgBot       `di.inject:"TgBot"`
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

	userState.MessageHandlerName = FileIdFormBeanId
	err = c.states.SetState(userState)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to set user state")
	}

	return c.tgbot.SendMessage(message.Chat, `Отправляйте файлы, в ответ я напишу их идентификатор.

Если написать идентификатор, то я пришлю ссылку на скачивание фото.`)
}

func (c *FileIdCommand) Callback(_ *tgbotapi.CallbackQuery, _ string) error {
	return nil
}

package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/goioc/di"
	"github.com/samber/lo"
	"reflect"
)

const (
	FileIdFormBeanId = "FileIdForm"
)

type FileIdForm struct {
	tgbot *TgBot `di.inject:"TgBot"`
}

func RegisterFileIdFormBean() {
	_ = lo.Must(di.RegisterBean(FileIdFormBeanId, reflect.TypeOf((*FileIdForm)(nil))))
}

func (f *FileIdForm) Handle(message *tgbotapi.Message) error {
	if len(message.Photo) > 0 {
		maxPhotoSize := lo.MaxBy(
			message.Photo, func(a tgbotapi.PhotoSize, b tgbotapi.PhotoSize) bool {
				return a.FileSize > b.FileSize
			},
		)
		err := f.tgbot.SendMessageCustom(message.Chat, "FileId = "+maxPhotoSize.FileID, func(reply *tgbotapi.MessageConfig) {
			reply.ReplyToMessageID = message.MessageID
		})
		if err != nil {
			return err
		}
	} else {
		err := f.tgbot.SendMessageCustom(message.Chat, "Фото не найдено", func(reply *tgbotapi.MessageConfig) {
			reply.ReplyToMessageID = message.MessageID
		})
		if err != nil {
			return err
		}
	}

	return nil
}

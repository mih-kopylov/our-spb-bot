package form

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mih-kopylov/our-spb-bot/pkg/tgbot"
	"github.com/samber/lo"
)

const (
	FileIdFormName = "FileIdForm"
)

type FileIdForm struct {
	service *tgbot.Service
}

func NewFileIdForm(service *tgbot.Service) tgbot.Form {
	return &FileIdForm{
		service: service,
	}
}

func (f *FileIdForm) Name() string {
	return FileIdFormName
}

func (f *FileIdForm) Handle(message *tgbotapi.Message) error {
	if len(message.Photo) > 0 {
		maxPhotoSize := lo.MaxBy(
			message.Photo, func(a tgbotapi.PhotoSize, b tgbotapi.PhotoSize) bool {
				return a.Width*a.Height > b.Width*b.Height
			},
		)

		directURL, err := f.service.GetFileDirectUrl(maxPhotoSize.FileID)
		if err != nil {
			return err
		}

		reply := fmt.Sprintf(`Id: %v
Вес: %v байт
Url: %v`,
			maxPhotoSize.FileID,
			maxPhotoSize.FileSize,
			directURL)
		_, err = f.service.SendMessageCustom(message.Chat, reply, func(reply *tgbotapi.MessageConfig) {
			reply.ReplyToMessageID = message.MessageID
		})
		if err != nil {
			return err
		}
	} else if message.Text != "" {
		fileId := message.Text
		directURL, err := f.service.GetFileDirectUrl(fileId)
		if err != nil {
			return err
		}

		reply := fmt.Sprintf(`FileId: %v
Url: %v`, fileId, directURL)
		_, err = f.service.SendMessageCustom(message.Chat, reply, func(reply *tgbotapi.MessageConfig) {
			reply.ReplyToMessageID = message.MessageID
		})
		if err != nil {
			return err
		}
	} else {
		_, err := f.service.SendMessageCustom(message.Chat, "Фото не найдено", func(reply *tgbotapi.MessageConfig) {
			reply.ReplyToMessageID = message.MessageID
		})
		if err != nil {
			return err
		}
	}

	return nil
}

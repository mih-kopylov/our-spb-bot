package handler

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
	"github.com/samber/lo"
	"golang.org/x/exp/slices"
	"time"
)

func SimpleMessageHandler(bot *tgbotapi.BotAPI, message *tgbotapi.Message, states *state.States) error {
	userState, err := states.GetState(message.From.ID)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to get user state")
	}

	if message.Text != "" {
		userState.OverrideText = message.Text
		reply := tgbotapi.NewMessage(message.Chat.ID, "Текст сообщения заменён")
		reply.ReplyToMessageID = message.MessageID
		_, err := bot.Send(reply)
		if err != nil {
			return errorx.EnhanceStackTrace(err, "failed to send reply")
		}
	}
	if len(message.Photo) > 0 {
		maxPhotoSize := lo.MaxBy(message.Photo, func(a tgbotapi.PhotoSize, b tgbotapi.PhotoSize) bool {
			return a.FileSize > b.FileSize
		})
		userState.Files = append(userState.Files, maxPhotoSize.FileID)

		reply := tgbotapi.NewMessage(message.Chat.ID, "Готово к отправке")
		reply.ReplyToMessageID = message.MessageID
		reply.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButtonLocation("Отправить сообщение"),
			),
		)
		_, err := bot.Send(reply)
		if err != nil {
			return errorx.EnhanceStackTrace(err, "failed to send reply")
		}
	}
	if message.Location != nil {
		text := userState.CurrentCategory.Message
		if userState.OverrideText != "" {
			text = userState.OverrideText
		}

		reply := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf(`
Сообщение добавлено в очередь и будет отправлено при первой возможности.

Пользователь: @%v
Категория: %v
Текст: %v
Локация: %v %v
Файлы: %v
`, message.From.UserName, userState.CurrentCategory.Category.Id, text,
			message.Location.Longitude, message.Location.Latitude, userState.Files))
		reply.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)

		userState.Queue = append(userState.Queue, state.QueueMessage{
			CategoryId: userState.CurrentCategory.Category.Id,
			FileUrls:   slices.Clone(userState.Files),
			Text:       text,
			SentAt:     time.Now(),
		})
		userState.Files = []string{}
		userState.OverrideText = ""
		userState.ResetCurrentCategory()

		_, err = bot.Send(reply)
		if err != nil {
			return errorx.EnhanceStackTrace(err, "failed to send reply")
		}
	}

	return nil
}

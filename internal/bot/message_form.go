package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/goioc/di"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/category"
	"github.com/mih-kopylov/our-spb-bot/internal/queue"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
	"github.com/samber/lo"
	"reflect"
	"time"
)

const (
	MessageFormBeanId = "MessageForm"
)

type MessageForm struct {
	states         *state.States                  `di.inject:"States"`
	tgbot          *TgBot                         `di.inject:"TgBot"`
	messageQueue   queue.MessageQueue             `di.inject:"Queue"`
	cateogiresTree *category.UserCategoryTreeNode `di.inject:"Categories"`
}

func RegisterMessageFormBean() {
	_ = lo.Must(di.RegisterBean(MessageFormBeanId, reflect.TypeOf((*MessageForm)(nil))))
}

func (f *MessageForm) Handle(message *tgbotapi.Message) error {
	userState, err := f.states.GetState(message.Chat.ID)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to get user state")
	}

	if message.Text != "" {
		err := userState.SetMessageText(message.Text)
		if err != nil {
			return errorx.EnhanceStackTrace(err, "failed to set message text")
		}

		return f.tgbot.SendMessageCustom(
			message.Chat, "Текст сообщения заменён", func(reply *tgbotapi.MessageConfig) {
				reply.ReplyToMessageID = message.MessageID
			},
		)
	}

	if len(message.Photo) > 0 {
		maxPhotoSize := lo.MaxBy(
			message.Photo, func(a tgbotapi.PhotoSize, b tgbotapi.PhotoSize) bool {
				return a.FileSize > b.FileSize
			},
		)
		err := userState.AddFile(maxPhotoSize.FileID)
		if err != nil {
			return errorx.EnhanceStackTrace(err, "failed to add file")
		}

		return f.tgbot.SendMessageCustom(
			message.Chat, "Фотография добавлена", func(reply *tgbotapi.MessageConfig) {
				reply.ReplyToMessageID = message.MessageID
				reply.ReplyMarkup = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButtonLocation("Отправить сообщение"),
					),
				)
			},
		)
	}

	if message.Location != nil {
		categoryTreeNode := f.cateogiresTree.FindNodeById(userState.GetCurrentCategoryNodeId())
		if categoryTreeNode == nil {
			return errorx.AssertionFailed.New("category is expected to be selected at this phase")
		}

		err := f.messageQueue.Add(
			userState.GetUserId(), &queue.Message{
				CategoryId: categoryTreeNode.Category.Id,
				FileUrls:   userState.GetFiles(),
				Text:       userState.GetMessageText(),
				Location: queue.Location{
					Longitude: message.Location.Longitude,
					Latitude:  message.Location.Latitude,
				},
				SentAt: time.Now(),
			},
		)
		if err != nil {
			return errorx.EnhanceStackTrace(err, "failed to add message to queue")
		}

		replyText := fmt.Sprintf(
			`
Сообщение добавлено в очередь и будет отправлено при первой возможности.

Пользователь: @%v
Категория: %v
Текст: %v
Локация: %v %v
Файлы: %v
`, message.Chat.UserName, categoryTreeNode.Category.Id, userState.GetMessageText(),
			message.Location.Longitude, message.Location.Latitude, userState.GetFiles(),
		)
		err = f.tgbot.SendMessageCustom(
			message.Chat, replyText, func(reply *tgbotapi.MessageConfig) {
				reply.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
			},
		)
		if err != nil {
			return err
		}

		err = userState.ClearFiles()
		if err != nil {
			return errorx.EnhanceStackTrace(err, "failed to clear files")
		}

		err = userState.SetMessageText("")
		if err != nil {
			return errorx.EnhanceStackTrace(err, "failed to clear message text")
		}

		err = userState.SetCurrentCategoryNodeId("")
		if err != nil {
			return errorx.EnhanceStackTrace(err, "failed to reset current category")
		}

		err = userState.SetMessageHandlerName("")
		if err != nil {
			return errorx.EnhanceStackTrace(err, "failed to clear message handler")
		}
	}

	return nil
}

package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/goioc/di"
	"github.com/joomcode/errorx"
	"github.com/lithammer/shortuuid/v4"
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
	states         state.States                   `di.inject:"States"`
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
		userState.MessageText = message.Text
		err := f.states.SetState(userState)
		if err != nil {
			return errorx.EnhanceStackTrace(err, "failed to set user state")
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

		userState.Files = append(userState.Files, maxPhotoSize.FileID)
		err = f.states.SetState(userState)
		if err != nil {
			return errorx.EnhanceStackTrace(err, "failed to set user state")
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
		categoryTreeNode := f.cateogiresTree.FindNodeById(userState.CurrentCategoryNodeId)
		if categoryTreeNode == nil {
			return errorx.AssertionFailed.New("category is expected to be selected at this phase")
		}

		queueMessage := queue.Message{
			Id:         shortuuid.New(),
			UserId:     userState.UserId,
			CategoryId: categoryTreeNode.Category.Id,
			Files:      userState.Files,
			Text:       userState.MessageText,
			Longitude:  message.Location.Longitude,
			Latitude:   message.Location.Latitude,
			CreatedAt:  time.Now(),
			Status:     queue.StatusCreated,
		}
		err := f.messageQueue.Add(&queueMessage)
		if err != nil {
			return errorx.EnhanceStackTrace(err, "failed to add message to queue")
		}

		replyText := fmt.Sprintf(`
Сообщение добавлено в очередь и будет отправлено при первой возможности.

Пользователь: @%v
Сообщение: %v
Категория: %v
Текст: %v
Локация: %v %v
Файлы: %v шт.: %v

/message - отправить новое обращение 
`, message.Chat.UserName,
			queueMessage.Id,
			queueMessage.CategoryId,
			queueMessage.Text,
			queueMessage.Longitude,
			queueMessage.Latitude,
			len(queueMessage.Files),
			queueMessage.Files,
		)
		err = f.tgbot.SendMessageCustom(
			message.Chat, replyText, func(reply *tgbotapi.MessageConfig) {
				reply.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
			},
		)
		if err != nil {
			return err
		}

		userState.Files = nil
		userState.MessageText = ""
		userState.CurrentCategoryNodeId = ""
		userState.MessageHandlerName = ""

		err = f.states.SetState(userState)
		if err != nil {
			return errorx.EnhanceStackTrace(err, "failed to set user state")
		}
	}

	return nil
}

package form

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/lithammer/shortuuid/v4"
	"github.com/mih-kopylov/our-spb-bot/internal/bot"
	"github.com/mih-kopylov/our-spb-bot/internal/bot/service"
	"github.com/mih-kopylov/our-spb-bot/internal/category"
	"github.com/mih-kopylov/our-spb-bot/internal/queue"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
	"github.com/samber/lo"
	"strings"
	"time"
)

const (
	MessageFormName = "MessageForm"
)

type MessageForm struct {
	states         state.States
	service        *service.Service
	messageQueue   queue.MessageQueue
	cateogiresTree *category.UserCategoryTreeNode
}

func (f *MessageForm) Name() string {
	return MessageFormName
}

func NewMessageForm(states state.States, service *service.Service,
	messageQueue queue.MessageQueue, cateogiresTree *category.UserCategoryTreeNode) bot.Form {
	return &MessageForm{
		states:         states,
		service:        service,
		messageQueue:   messageQueue,
		cateogiresTree: cateogiresTree,
	}
}

func (f *MessageForm) Handle(message *tgbotapi.Message) error {
	userState, err := f.states.GetState(message.Chat.ID)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to get user state")
	}

	if message.Text != "" {
		userState.SetFormField(state.FormFieldMessageText, message.Text)

		err = f.states.SetState(userState)
		if err != nil {
			return errorx.EnhanceStackTrace(err, "failed to set user state")
		}

		replyText := "Текст сообщения заменён."
		return f.service.SendMessageCustom(
			message.Chat, replyText, func(reply *tgbotapi.MessageConfig) {
				reply.ReplyToMessageID = message.MessageID
			},
		)
	}

	if len(message.Photo) > 0 {
		maxPhotoSize := lo.MaxBy(
			message.Photo, func(a tgbotapi.PhotoSize, b tgbotapi.PhotoSize) bool {
				return a.Width*a.Height > b.Width*b.Height
			},
		)

		userState.AddValueToStringSlice(state.FormFieldFiles, maxPhotoSize.FileID)
		err = f.states.SetState(userState)
		if err != nil {
			return errorx.EnhanceStackTrace(err, "failed to set user state")
		}

		reply := fmt.Sprintf(`Фотография добавлена

Размер: %vx%v`, maxPhotoSize.Width, maxPhotoSize.Height)
		return f.service.SendMessageCustom(
			message.Chat, reply, func(reply *tgbotapi.MessageConfig) {
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
		categoryTreeNode := f.cateogiresTree.FindNodeById(userState.GetStringFormField(state.FormFieldCurrentCategoryNode))
		if categoryTreeNode == nil {
			return errorx.AssertionFailed.New("category is expected to be selected at this phase")
		}

		createdAt := time.Now()
		messageId := createdAt.Format("06-01-02") + "_" + shortuuid.New()
		if strings.Contains(message.Text, "!") {
			messageId = "00_" + messageId
		}

		queueMessage := queue.Message{
			Id:         messageId,
			UserId:     userState.UserId,
			CategoryId: categoryTreeNode.Category.Id,
			Files:      userState.GetStringSlice(state.FormFieldFiles),
			Text:       userState.GetStringFormField(state.FormFieldMessageText),
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

/status - статус обращений
`, message.Chat.UserName,
			queueMessage.Id,
			queueMessage.CategoryId,
			queueMessage.Text,
			queueMessage.Longitude,
			queueMessage.Latitude,
			len(queueMessage.Files),
			queueMessage.Files,
		)
		err = f.service.SendMessageCustom(
			message.Chat, replyText, func(reply *tgbotapi.MessageConfig) {
				reply.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
			},
		)
		if err != nil {
			return err
		}

		userState.ClearForm()
		userState.MessageHandlerName = ""

		err = f.states.SetState(userState)
		if err != nil {
			return errorx.EnhanceStackTrace(err, "failed to set user state")
		}
	}

	return nil
}
package form

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/lithammer/shortuuid/v4"
	"github.com/mih-kopylov/our-spb-bot/internal/bot/callback"
	"github.com/mih-kopylov/our-spb-bot/internal/bot/service"
	"github.com/mih-kopylov/our-spb-bot/internal/category"
	"github.com/mih-kopylov/our-spb-bot/internal/queue"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
	"github.com/mih-kopylov/our-spb-bot/pkg/bot"
	"github.com/samber/lo"
	"strconv"
	"strings"
	"time"
)

const (
	MessageFormName = "MessageForm"
)

type MessageForm struct {
	states                state.States
	service               *service.Service
	messageQueue          queue.MessageQueue
	categoryService       *category.Service
	deleteMessageCallback *callback.DeleteMessageCallback
	deletePhotoCallback   *callback.DeletePhotoCallback
}

func (f *MessageForm) Name() string {
	return MessageFormName
}

func NewMessageForm(states state.States, service *service.Service,
	messageQueue queue.MessageQueue, categoryService *category.Service,
	deleteMessageCallback *callback.DeleteMessageCallback, deletePhotoCallback *callback.DeletePhotoCallback) bot.Form {
	return &MessageForm{
		states:                states,
		service:               service,
		messageQueue:          messageQueue,
		categoryService:       categoryService,
		deleteMessageCallback: deleteMessageCallback,
		deletePhotoCallback:   deletePhotoCallback,
	}
}

func (f *MessageForm) Handle(message *tgbotapi.Message) error {
	userState, err := f.states.GetState(message.Chat.ID)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to get user state")
	}

	if message.Text != "" {
		return f.handleText(message, userState)
	}

	if len(message.Photo) > 0 {
		return f.handlePhoto(message, userState)
	}

	if message.Location != nil {
		return f.handleLocation(message, userState)
	}

	return nil
}

func (f *MessageForm) handleLocation(message *tgbotapi.Message, userState *state.UserState) error {
	categoriesTree, err := f.categoryService.ParseCategoriesTree(userState.Categories)
	if err != nil {
		return err
	}

	categoryTreeNode := categoriesTree.FindNodeById(userState.GetStringFormField(state.FormFieldCurrentCategoryNode))
	if categoryTreeNode == nil {
		return errorx.AssertionFailed.New("category is expected to be selected at this phase")
	}

	text := userState.GetStringFormField(state.FormFieldMessageText)
	createdAt := time.Now()
	messageId := createdAt.Format("06-01-02") + "_" + shortuuid.New()
	if strings.Contains(text, "!") {
		messageId = "00_" + messageId
	}

	queueMessage := queue.Message{
		Id:         messageId,
		UserId:     userState.UserId,
		CategoryId: categoryTreeNode.Category.Id,
		Files:      userState.GetStringSlice(state.FormFieldFiles),
		Text:       text,
		Longitude:  message.Location.Longitude,
		Latitude:   message.Location.Latitude,
		CreatedAt:  createdAt,
		Status:     queue.StatusCreated,
	}
	err = f.messageQueue.Add(&queueMessage)
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
`, message.Chat.UserName,
		queueMessage.Id,
		queueMessage.CategoryId,
		queueMessage.Text,
		queueMessage.Longitude,
		queueMessage.Latitude,
		len(queueMessage.Files),
		queueMessage.Files,
	)
	_, err = f.service.SendMessageCustom(
		message.Chat, replyText, func(reply *tgbotapi.MessageConfig) {
			reply.ReplyMarkup = f.deleteMessageCallback.CreateReplyMarkup(queueMessage.Id)
		},
	)
	if err != nil {
		return err
	}

	nextCommandsMessageText := `/message - отправить новое обращение 

/status - статус обращений
`

	_, err = f.service.SendMessageCustom(message.Chat, nextCommandsMessageText, func(reply *tgbotapi.MessageConfig) {
		reply.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
	})
	if err != nil {
		return err
	}

	userState.ClearForm()
	userState.MessageHandlerName = ""

	err = f.states.SetState(userState)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to set user state")
	}

	return nil
}

func (f *MessageForm) handlePhoto(message *tgbotapi.Message, userState *state.UserState) error {
	if len(userState.GetStringSlice(state.FormFieldFiles)) == 5 {
		replyText := `Портал допускает максимум 5 файлов в обращении.
Это фото не будет приложено. 
Для того, чтобы использовать именно это фото, можно удалить одно из предыдущих.`
		_, err := f.service.SendMessageCustom(message.Chat, replyText, func(reply *tgbotapi.MessageConfig) {
			reply.ReplyToMessageID = message.MessageID
		})
		return err
	}

	maxPhotoSize := lo.MaxBy(
		message.Photo, func(a tgbotapi.PhotoSize, b tgbotapi.PhotoSize) bool {
			return a.Width*a.Height > b.Width*b.Height
		},
	)

	userState.AddValueToStringSlice(state.FormFieldFiles, maxPhotoSize.FileID)
	userState.PutValueToMap(state.FormFieldMessageIdFile, strconv.Itoa(message.MessageID), maxPhotoSize.FileID)
	err := f.states.SetState(userState)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to set user state")
	}

	replyText := fmt.Sprintf(`Фотография добавлена.

Id: %v
Размер: %vx%v
Вес: %v байт`,
		maxPhotoSize.FileID,
		maxPhotoSize.Width,
		maxPhotoSize.Height,
		maxPhotoSize.FileSize)
	_, err = f.service.SendMessageCustom(message.Chat, replyText, func(reply *tgbotapi.MessageConfig) {
		reply.ReplyToMessageID = message.MessageID
		reply.ReplyMarkup = f.deletePhotoCallback.CreateMarkup(message.MessageID)
	})
	if err != nil {
		return err
	}

	if len(userState.GetStringSlice(state.FormFieldFiles)) == 1 {
		//add Send button only once, when the first photo is added
		_, err = f.service.SendMessageCustom(
			message.Chat, `Теперь можно отправить обращение`, func(reply *tgbotapi.MessageConfig) {
				replyMarkup := tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButtonLocation("Отправить обращение"),
					),
				)
				replyMarkup.OneTimeKeyboard = true
				reply.ReplyMarkup = replyMarkup
			},
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *MessageForm) handleText(message *tgbotapi.Message, userState *state.UserState) error {
	userState.SetFormField(state.FormFieldMessageText, message.Text)

	err := f.states.SetState(userState)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to set user state")
	}

	replyText := "Текст сообщения заменён."
	if strings.Contains(message.Text, "!") {
		replyText += "\nСообщение будет отправлено в первую очередь."
	}
	_, err = f.service.SendMessageCustom(
		message.Chat, replyText, func(reply *tgbotapi.MessageConfig) {
			reply.ReplyToMessageID = message.MessageID
		},
	)
	return err
}

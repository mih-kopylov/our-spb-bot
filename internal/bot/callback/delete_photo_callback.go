package callback

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/bot/service"
	"github.com/mih-kopylov/our-spb-bot/internal/queue"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
	"github.com/mih-kopylov/our-spb-bot/pkg/bot"
	"strconv"
)

const (
	DeletePhotoCallbackName = "DeletePhoto"
)

type DeletePhotoCallback struct {
	states       state.States
	service      *service.Service
	messageQueue queue.MessageQueue
}

func NewDeletePhotoCallback(states state.States, service *service.Service, messageQueue queue.MessageQueue) *DeletePhotoCallback {
	return &DeletePhotoCallback{
		states:       states,
		service:      service,
		messageQueue: messageQueue,
	}
}

func (h *DeletePhotoCallback) Name() string {
	return DeletePhotoCallbackName
}

func (h *DeletePhotoCallback) Handle(callbackQuery *tgbotapi.CallbackQuery, data string) error {
	userState, err := h.states.GetState(callbackQuery.Message.Chat.ID)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to get user state")
	}

	fileId, exists := userState.GetStringMap(state.FormFieldMessageIdFile)[data]
	if !exists {
		return errorx.IllegalArgument.New("failed to find fileId by messageid: %v", data)
	}

	messageIdInt, err := strconv.Atoi(data)
	if err != nil {
		return errorx.IllegalArgument.New("failed to parse messageId from callback data: %v", data)
	}

	userState.RemoveValueFromStringSlice(state.FormFieldFiles, fileId)
	err = h.states.SetState(userState)
	if err != nil {
		return err
	}

	err = h.service.DeleteMessage(callbackQuery.Message)
	if err != nil {
		return err
	}

	err = h.service.DeleteMessageById(callbackQuery.Message.Chat.ID, messageIdInt)
	if err != nil {
		return err
	}

	if len(userState.GetStringSlice(state.FormFieldFiles)) == 0 {
		//when the last photo is removed, remove the Send button as well
		_, err = h.service.SendMessageCustom(callbackQuery.Message.Chat, `Добавьте хотя бы одно фото`, func(reply *tgbotapi.MessageConfig) {
			reply.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *DeletePhotoCallback) CreateMarkup(messageId int) tgbotapi.InlineKeyboardMarkup {
	result := tgbotapi.NewInlineKeyboardMarkup()
	deleteButton := tgbotapi.NewInlineKeyboardButtonData("🗑 Удалить", DeletePhotoCallbackName+bot.CallbackSectionSeparator+strconv.Itoa(messageId))
	result.InlineKeyboard = append(result.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(deleteButton))
	return result
}

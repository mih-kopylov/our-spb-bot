package callback

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/queue"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
	"github.com/mih-kopylov/our-spb-bot/pkg/tgbot"
)

const (
	DeleteMessageCallbackName = "DeleteMessageCallback"
)

type DeleteMessageCallback struct {
	states       state.States
	service      *tgbot.Service
	messageQueue queue.MessageQueue
}

func NewDeleteMessageCallback(states state.States, service *tgbot.Service, messageQueue queue.MessageQueue) *DeleteMessageCallback {
	return &DeleteMessageCallback{
		states:       states,
		service:      service,
		messageQueue: messageQueue,
	}
}

func (h *DeleteMessageCallback) Name() string {
	return DeleteMessageCallbackName
}

func (h *DeleteMessageCallback) Handle(callbackQuery *tgbotapi.CallbackQuery, data string) error {
	userState, err := h.states.GetState(callbackQuery.Message.Chat.ID)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to get user state")
	}

	message, err := h.messageQueue.GetMessage(data)
	if err != nil {
		return h.service.SendMessage(callbackQuery.Message.Chat, fmt.Sprintf(`Не удалось удалить сообщение %v.
Возможно, уже было отправлено.`, data))
	}

	if message.UserId != userState.UserId {
		return errorx.EnhanceStackTrace(err, "can't delete a message of another user")
	}

	err = h.messageQueue.DeleteMessage(message)
	if err != nil {
		return err
	}

	replyText := fmt.Sprintf(`Сообщение %v удалено`, data)
	reply := tgbotapi.NewEditMessageText(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID, replyText)
	err = h.service.Send(reply)
	if err != nil {
		return err
	}

	return nil
}

func (h *DeleteMessageCallback) CreateReplyMarkup(messageId string) tgbotapi.InlineKeyboardMarkup {
	result := tgbotapi.NewInlineKeyboardMarkup()
	deleteButton := tgbotapi.NewInlineKeyboardButtonData("🗑 Удалить", DeleteMessageCallbackName+tgbot.CallbackSectionSeparator+messageId)
	result.InlineKeyboard = append(result.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(deleteButton))
	return result
}

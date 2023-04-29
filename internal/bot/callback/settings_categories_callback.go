package callback

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/bot"
	"github.com/mih-kopylov/our-spb-bot/internal/bot/service"
	"github.com/mih-kopylov/our-spb-bot/internal/category"
	"github.com/mih-kopylov/our-spb-bot/internal/queue"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
)

const (
	SettingsCategoriesCallbackName = "SettingsCategoriesCallback"
	DownloadButtonId               = "Download"
	UploadButtonId                 = "Upload"
	ResetButtonId                  = "Reset"
)

type SettingsCategoriesCallback struct {
	states       state.States
	service      *service.Service
	messageQueue queue.MessageQueue
}

func NewSettingsCategoriesCallback(states state.States, service *service.Service, messageQueue queue.MessageQueue) *SettingsCategoriesCallback {
	return &SettingsCategoriesCallback{
		states:       states,
		service:      service,
		messageQueue: messageQueue,
	}
}

func (h *SettingsCategoriesCallback) Name() string {
	return SettingsCategoriesCallbackName
}

func (h *SettingsCategoriesCallback) Handle(callbackQuery *tgbotapi.CallbackQuery, data string) error {
	userState, err := h.states.GetState(callbackQuery.Message.Chat.ID)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to get user state")
	}

	switch data {
	case DownloadButtonId:
		bytes := []byte(userState.Categories)
		fileName := "categories.yaml"
		err = h.service.SendDocument(callbackQuery.Message.Chat, bytes, fileName)
		if err != nil {
			return err
		}

		return h.service.SendMessage(callbackQuery.Message.Chat, `В документе выше структура категорий.
Его нужно скачать, отредактировать и загрузить обновлённые категории.`)
	case UploadButtonId:
		userState.MessageHandlerName = "UploadCategoriesForm"
		err = h.states.SetState(userState)
		if err != nil {
			return err
		}

		return h.service.SendMessage(callbackQuery.Message.Chat, "Загрузите документ с категориями")
	case ResetButtonId:
		userState.Categories = string(category.DefaultCategoriesText)
		err = h.states.SetState(userState)
		if err != nil {
			return err
		}

		return h.service.SendMessage(callbackQuery.Message.Chat, "Установлены категории по умолчанию")
	default:
		return errorx.IllegalArgument.New("unsupported data: %v", data)
	}
}

func (h *SettingsCategoriesCallback) CreateReplyMarkup() tgbotapi.InlineKeyboardMarkup {
	result := tgbotapi.NewInlineKeyboardMarkup()
	downloadButton := tgbotapi.NewInlineKeyboardButtonData("Скачать категории", SettingsCategoriesCallbackName+bot.CallbackSectionSeparator+DownloadButtonId)
	uploadButton := tgbotapi.NewInlineKeyboardButtonData("Загрузить новые категории", SettingsCategoriesCallbackName+bot.CallbackSectionSeparator+UploadButtonId)
	resetButton := tgbotapi.NewInlineKeyboardButtonData("Сбросить на значения по умолчанию", SettingsCategoriesCallbackName+bot.CallbackSectionSeparator+ResetButtonId)
	result.InlineKeyboard = append(result.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(downloadButton, uploadButton))
	result.InlineKeyboard = append(result.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(resetButton))
	return result
}

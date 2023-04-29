package callback

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/bot"
	"github.com/mih-kopylov/our-spb-bot/internal/bot/service"
)

const (
	SettingsCallbackName = "SettingsCallback"
	CategoriesButtonId   = "Categories"
)

type SettingsCallback struct {
	service                    *service.Service
	settingsCategoriesCallback *SettingsCategoriesCallback
}

func NewSettingsCallback(service *service.Service, settingsCategoriesCallback *SettingsCategoriesCallback) *SettingsCallback {
	return &SettingsCallback{
		service:                    service,
		settingsCategoriesCallback: settingsCategoriesCallback,
	}
}

func (h *SettingsCallback) Name() string {
	return SettingsCallbackName
}

func (h *SettingsCallback) Handle(callbackQuery *tgbotapi.CallbackQuery, data string) error {
	switch data {
	case CategoriesButtonId:
		reply := tgbotapi.NewEditMessageTextAndMarkup(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID,
			`Настройка категорий`, h.settingsCategoriesCallback.CreateReplyMarkup())
		err := h.service.Send(reply)
		if err != nil {
			return err
		}
		return nil
	default:
		return errorx.IllegalArgument.New("unsupported data: %v", data)
	}
}

func (h *SettingsCallback) CreateReplyMarkup() tgbotapi.InlineKeyboardMarkup {
	result := tgbotapi.NewInlineKeyboardMarkup()
	categoriesButton := tgbotapi.NewInlineKeyboardButtonData("Категории", SettingsCallbackName+bot.CallbackSectionSeparator+CategoriesButtonId)
	result.InlineKeyboard = append(result.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(categoriesButton))
	return result
}

package callback

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/bot"
	"github.com/mih-kopylov/our-spb-bot/internal/bot/service"
)

const (
	SettingsCallbackName = "SettingsCallback"
	categoriesButtonId   = "Categories"
	accountsButtonId     = "Accounts"
)

type SettingsCallback struct {
	service                    *service.Service
	settingsCategoriesCallback *SettingsCategoriesCallback
	settingsAccountsCallback   *SettingsAccountsCallback
}

func NewSettingsCallback(service *service.Service, settingsCategoriesCallback *SettingsCategoriesCallback, settingsAccountsCallback *SettingsAccountsCallback) *SettingsCallback {
	return &SettingsCallback{
		service:                    service,
		settingsCategoriesCallback: settingsCategoriesCallback,
		settingsAccountsCallback:   settingsAccountsCallback,
	}
}

func (h *SettingsCallback) Name() string {
	return SettingsCallbackName
}

func (h *SettingsCallback) Handle(callbackQuery *tgbotapi.CallbackQuery, data string) error {
	switch data {
	case categoriesButtonId:
		return h.settingsCategoriesCallback.HandleCategorySettingsButtonClick(callbackQuery)
	case accountsButtonId:
		return h.settingsAccountsCallback.HandleCategoryAccountsButtonClick(callbackQuery)
	default:
		return errorx.IllegalArgument.New("unsupported data: %v", data)
	}
}

func (h *SettingsCallback) CreateReplyMarkup() tgbotapi.InlineKeyboardMarkup {
	result := tgbotapi.NewInlineKeyboardMarkup()
	categoriesButton := tgbotapi.NewInlineKeyboardButtonData("Категории", SettingsCallbackName+bot.CallbackSectionSeparator+categoriesButtonId)
	accountsButton := tgbotapi.NewInlineKeyboardButtonData("Аккаунты", SettingsCallbackName+bot.CallbackSectionSeparator+accountsButtonId)
	result.InlineKeyboard = append(result.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(categoriesButton, accountsButton))
	return result
}

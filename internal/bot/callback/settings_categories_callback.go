package callback

import (
	"encoding/json"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/bot"
	"github.com/mih-kopylov/our-spb-bot/internal/bot/service"
	"github.com/mih-kopylov/our-spb-bot/internal/category"
	"github.com/mih-kopylov/our-spb-bot/internal/spb"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
)

const (
	SettingsCategoriesCallbackName = "SettingsCategoriesCallback"
	downloadButtonId               = "Download"
	uploadButtonId                 = "Upload"
	resetButtonId                  = "Reset"
	downloadPortalButtonId         = "DownloadPortal"
)

type SettingsCategoriesCallback struct {
	states    state.States
	service   *service.Service
	spbClient spb.Client
}

func NewSettingsCategoriesCallback(states state.States, service *service.Service, spbClient spb.Client) *SettingsCategoriesCallback {
	return &SettingsCategoriesCallback{
		states:    states,
		service:   service,
		spbClient: spbClient,
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
	case downloadButtonId:
		bytes := []byte(userState.Categories)
		fileName := "categories.yaml"
		err = h.service.SendDocument(callbackQuery.Message.Chat, bytes, fileName)
		if err != nil {
			return err
		}

		return h.service.SendMessage(callbackQuery.Message.Chat, `В выложенном документе структура категорий.
Его нужно скачать, отредактировать и загрузить обновлённые категории.`)
	case uploadButtonId:
		userState.MessageHandlerName = "UploadCategoriesForm"
		err = h.states.SetState(userState)
		if err != nil {
			return err
		}

		return h.service.SendMessage(callbackQuery.Message.Chat, "Загрузите документ с категориями")
	case resetButtonId:
		userState.Categories = string(category.DefaultCategoriesText)
		err = h.states.SetState(userState)
		if err != nil {
			return err
		}

		return h.service.SendMessage(callbackQuery.Message.Chat, "Установлены категории по умолчанию")
	case downloadPortalButtonId:
		reasons, err := h.spbClient.GetReasons()
		if err != nil {
			return err
		}

		bytes, err := json.MarshalIndent(reasons, "", "  ")
		if err != nil {
			return errorx.EnhanceStackTrace(err, "failed to marshall portal categories")
		}

		return h.service.SendDocument(callbackQuery.Message.Chat, bytes, "portalCategories.json")
	default:
		return errorx.IllegalArgument.New("unsupported data: %v", data)
	}
}

func (h *SettingsCategoriesCallback) HandleCategorySettingsButtonClick(callbackQuery *tgbotapi.CallbackQuery) error {
	reply := tgbotapi.NewEditMessageTextAndMarkup(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID,
		`Настройка категорий.

Для того, чтобы настроить удобные для себя категории, нужно скачать категории портала и свои категории.
В файле со своими категориями упорядочить их так, как удобно.
После этого загрузить файл со своими категориями обратно.`, h.CreateReplyMarkup())
	err := h.service.Send(reply)
	if err != nil {
		return err
	}
	return nil
}

func (h *SettingsCategoriesCallback) CreateReplyMarkup() tgbotapi.InlineKeyboardMarkup {
	result := tgbotapi.NewInlineKeyboardMarkup()
	downloadButton := tgbotapi.NewInlineKeyboardButtonData("Скачать свои категории", SettingsCategoriesCallbackName+bot.CallbackSectionSeparator+downloadButtonId)
	uploadButton := tgbotapi.NewInlineKeyboardButtonData("Загрузить новые категории", SettingsCategoriesCallbackName+bot.CallbackSectionSeparator+uploadButtonId)
	resetButton := tgbotapi.NewInlineKeyboardButtonData("Сбросить на значения по умолчанию", SettingsCategoriesCallbackName+bot.CallbackSectionSeparator+resetButtonId)
	downloadPortalButton := tgbotapi.NewInlineKeyboardButtonData("Скачать категории портала", SettingsCategoriesCallbackName+bot.CallbackSectionSeparator+downloadPortalButtonId)
	result.InlineKeyboard = append(result.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(downloadButton, uploadButton))
	result.InlineKeyboard = append(result.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(resetButton))
	result.InlineKeyboard = append(result.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(downloadPortalButton))
	return result
}

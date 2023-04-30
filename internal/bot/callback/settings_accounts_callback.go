package callback

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/bot"
	"github.com/mih-kopylov/our-spb-bot/internal/bot/service"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
	"github.com/samber/lo"
	"strings"
)

const (
	SettingsAccountsCallbackName = "SettingsAccounts"
	actionsAccountButtonId       = "actions"
	disableAccountButtonId       = "disable"
	enableAccountButtonId        = "enable"
	deleteAccountButtonId        = "delete"
	listAccountsButtonId         = "list"
)

type SettingsAccountsCallback struct {
	states  state.States
	service *service.Service
}

func NewSettingsAccountsCallback(states state.States, service *service.Service) *SettingsAccountsCallback {
	return &SettingsAccountsCallback{
		states:  states,
		service: service,
	}
}

func (h *SettingsAccountsCallback) Name() string {
	return SettingsAccountsCallbackName
}

func (h *SettingsAccountsCallback) Handle(callbackQuery *tgbotapi.CallbackQuery, data string) error {
	userState, err := h.states.GetState(callbackQuery.Message.Chat.ID)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to get user state")
	}

	if data == listAccountsButtonId {
		return h.HandleCategoryAccountsButtonClick(callbackQuery)
	}

	action, value, found := strings.Cut(data, bot.CallbackSectionSeparator)
	if !found {
		return errorx.IllegalArgument.New("failed to parse callback data: %v", data)
	}

	switch action {
	case actionsAccountButtonId:
		return h.handleActionsAccountButton(callbackQuery, value, userState)

	case disableAccountButtonId:
		return h.setAccountStateButton(callbackQuery, value, userState, state.AccountStateDisabled)

	case enableAccountButtonId:
		return h.setAccountStateButton(callbackQuery, value, userState, state.AccountStateEnabled)

	case deleteAccountButtonId:
		return h.handleDeleteAccountButton(callbackQuery, value, userState)

	default:
		return errorx.IllegalArgument.New("unsupported data: %v", data)
	}

}

func (h *SettingsAccountsCallback) handleDeleteAccountButton(callbackQuery *tgbotapi.CallbackQuery, value string, userState *state.UserState) error {
	accountLogin := value
	_, index, found := lo.FindIndexOf(userState.Accounts, func(item state.Account) bool {
		return item.Login == accountLogin
	})
	if !found {
		return errorx.IllegalArgument.New("failed to find account: %v", accountLogin)
	}

	userState.Accounts = append(userState.Accounts[0:index], userState.Accounts[index+1:]...)
	err := h.states.SetState(userState)
	if err != nil {
		return err
	}

	return h.HandleCategoryAccountsButtonClick(callbackQuery)
}

func (h *SettingsAccountsCallback) setAccountStateButton(callbackQuery *tgbotapi.CallbackQuery, value string, userState *state.UserState, accountState state.AccountState) error {
	accountLogin := value
	_, index, found := lo.FindIndexOf(userState.Accounts, func(item state.Account) bool {
		return item.Login == accountLogin
	})
	if !found {
		return errorx.IllegalArgument.New("failed to find account: %v", accountLogin)
	}

	userState.Accounts[index].State = accountState
	err := h.states.SetState(userState)
	if err != nil {
		return err
	}

	return h.Handle(callbackQuery, actionsAccountButtonId+bot.CallbackSectionSeparator+accountLogin)
}

func (h *SettingsAccountsCallback) handleActionsAccountButton(callbackQuery *tgbotapi.CallbackQuery, value string, userState *state.UserState) error {
	accountLogin := value
	account, found := lo.Find(userState.Accounts, func(item state.Account) bool {
		return item.Login == accountLogin
	})
	if !found {
		return errorx.IllegalArgument.New("failed to find account: %v", accountLogin)
	}

	accountStateName, err := account.GetStateName()
	if err != nil {
		return err
	}

	replyText := fmt.Sprintf(`Аккаунт %v
Состояние: %v
Выберите действие`,
		accountLogin,
		accountStateName,
	)

	reply := tgbotapi.NewEditMessageTextAndMarkup(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID,
		replyText, h.createActionMarkup(account))
	err = h.service.Send(reply)
	if err != nil {
		return err
	}

	return nil
}

func (h *SettingsAccountsCallback) HandleCategoryAccountsButtonClick(callbackQuery *tgbotapi.CallbackQuery) error {
	replyMarkup, err := h.createListAccountsReplyMarkup(callbackQuery)
	if err != nil {
		return err
	}

	reply := tgbotapi.NewEditMessageTextAndMarkup(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID,
		`Настройка аккаунтов.
Выберите аккаунт`, replyMarkup)
	err = h.service.Send(reply)
	if err != nil {
		return err
	}

	return nil
}

func (h *SettingsAccountsCallback) createListAccountsReplyMarkup(callbackQuery *tgbotapi.CallbackQuery) (tgbotapi.InlineKeyboardMarkup, error) {
	result := tgbotapi.NewInlineKeyboardMarkup()
	result.InlineKeyboard = [][]tgbotapi.InlineKeyboardButton{}
	userState, err := h.states.GetState(callbackQuery.Message.Chat.ID)
	if err != nil {
		return result, err
	}

	for _, account := range userState.Accounts {
		buttonText := account.Login
		if account.State == state.AccountStateEnabled {
			buttonText += " ✔️"
		}
		if account.State == state.AccountStateDisabled {
			buttonText += " ❌"
		}
		accountButton := tgbotapi.NewInlineKeyboardButtonData(buttonText, SettingsAccountsCallbackName+bot.CallbackSectionSeparator+actionsAccountButtonId+bot.CallbackSectionSeparator+account.Login)
		result.InlineKeyboard = append(result.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(accountButton))
	}

	return result, nil
}

func (h *SettingsAccountsCallback) createActionMarkup(account state.Account) tgbotapi.InlineKeyboardMarkup {
	result := tgbotapi.NewInlineKeyboardMarkup()
	disableButton := tgbotapi.NewInlineKeyboardButtonData("Выключить", SettingsAccountsCallbackName+bot.CallbackSectionSeparator+disableAccountButtonId+bot.CallbackSectionSeparator+account.Login)
	enableButton := tgbotapi.NewInlineKeyboardButtonData("Включить", SettingsAccountsCallbackName+bot.CallbackSectionSeparator+enableAccountButtonId+bot.CallbackSectionSeparator+account.Login)
	deleteButton := tgbotapi.NewInlineKeyboardButtonData("Удалить", SettingsAccountsCallbackName+bot.CallbackSectionSeparator+deleteAccountButtonId+bot.CallbackSectionSeparator+account.Login)
	listButton := tgbotapi.NewInlineKeyboardButtonData("⬆ К списку", SettingsAccountsCallbackName+bot.CallbackSectionSeparator+listAccountsButtonId)
	row := tgbotapi.NewInlineKeyboardRow()
	if account.State == state.AccountStateEnabled {
		row = append(row, disableButton)
	}
	if account.State == state.AccountStateDisabled {
		row = append(row, enableButton)
	}
	row = append(row, deleteButton)
	result.InlineKeyboard = append(result.InlineKeyboard, row)
	result.InlineKeyboard = append(result.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(listButton))
	return result
}

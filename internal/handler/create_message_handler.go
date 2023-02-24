package handler

import (
	_ "embed"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/category"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
)

const (
	StartCommand   = "start"
	MessageCommand = "message"
	StatusCommand  = "status"
	DataBack       = "back"
)

func MessageCommandHandler(bot *tgbotapi.BotAPI, message *tgbotapi.Message, states *state.States) error {
	userState, err := states.GetState(message.From.ID)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to get user state")
	}

	userState.ResetCurrentCategory()

	reply := tgbotapi.NewMessage(message.Chat.ID, "")
	reply.ReplyToMessageID = message.MessageID

	reply.Text = "Выберите категорию"

	reply.ReplyMarkup = createCateogoriesReplyMarkup(userState)

	_, err = bot.Send(reply)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to send reply")
	}

	return nil
}

func SendCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, data string, states *state.States) error {
	userState, err := states.GetState(callbackQuery.Message.Chat.ID)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to get user state")
	}

	var markup tgbotapi.InlineKeyboardMarkup
	var replyText string
	var childFound *category.UserCategoryTreeNode
	if data == DataBack {
		if userState.CurrentCategory.Parent == nil {
			return errorx.AssertionFailed.New("can't go back more than a root")
		}
		childFound = userState.CurrentCategory.Parent
	} else {
		for _, child := range userState.CurrentCategory.Children {
			if child.Id == data {
				childFound = child
			}
		}
	}

	if childFound == nil {
		replyText = "Не удалось найти выбранную категорию"
		markup = tgbotapi.NewInlineKeyboardMarkup()
	} else {
		userState.CurrentCategory = childFound
		if childFound.Category == nil {
			replyText = "Выберите категорию"
			markup = createCateogoriesReplyMarkup(userState)
		} else {
			replyText = fmt.Sprintf(`Выбранная категория: %v
Отправьте фотографии`, childFound.GetFullName())
			markup = createCateogoriesReplyMarkup(userState)
		}
	}

	reply := tgbotapi.NewEditMessageTextAndMarkup(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID, replyText, markup)
	_, err = bot.Send(reply)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to send reply")
	}

	return nil
}

func createCateogoriesReplyMarkup(userState *state.UserState) tgbotapi.InlineKeyboardMarkup {
	result := tgbotapi.NewInlineKeyboardMarkup()

	if userState.CurrentCategory.Parent != nil {
		backButton := tgbotapi.NewInlineKeyboardButtonData("Вверх", SendSection+SectionSeparator+DataBack)
		result.InlineKeyboard = append(result.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(backButton))
	}

	for _, child := range userState.CurrentCategory.Children {
		result.InlineKeyboard = append(result.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(child.Text, SendSection+SectionSeparator+child.Id)))
	}

	return result
}

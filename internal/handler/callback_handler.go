package handler

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/category"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
	"strings"
)

const (
	SectionSeparator = "."
	SendSection      = "send"
)

func CallbackHandler(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, states *state.States) error {
	data := callbackQuery.Data
	section, value, found := strings.Cut(data, SectionSeparator)
	if !found {
		tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "unsupported callback data")
		return nil
	}

	var err error
	switch section {
	case SendSection:
		err = SendCallback(bot, callbackQuery, value, states)
	default:
		err = errorx.AssertionFailed.New("unsupportd section: %v", section)
	}
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to handle callback")
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

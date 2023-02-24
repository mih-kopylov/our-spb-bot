package handler

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
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

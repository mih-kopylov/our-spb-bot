package handler

import (
	_ "embed"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
)

const (
	DataBack = "back"
)

func MessageCommandHandler(bot *tgbotapi.BotAPI, message *tgbotapi.Message, states *state.States) error {
	userState, err := states.GetState(message.Chat.ID)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to get user state")
	}

	userState.MessageHandlerName = MessageFormHandlerName
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

func createCateogoriesReplyMarkup(userState *state.UserState) tgbotapi.InlineKeyboardMarkup {
	result := tgbotapi.NewInlineKeyboardMarkup()

	if userState.CurrentCategory.Parent != nil {
		backButton := tgbotapi.NewInlineKeyboardButtonData("Вверх", SendSection+SectionSeparator+DataBack)
		result.InlineKeyboard = append(result.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(backButton))
	}

	for _, child := range userState.CurrentCategory.Children {
		result.InlineKeyboard = append(result.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(child.Name, SendSection+SectionSeparator+child.Id)))
	}

	return result
}

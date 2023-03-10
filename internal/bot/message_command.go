package bot

import (
	_ "embed"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/category"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
)

const (
	MessageCommandName = "message"
	DataBack           = "back"
)

type MessageCommand struct {
	states         state.States                   `di.inject:"States"`
	tgbot          *TgBot                         `di.inject:"TgBot"`
	cateogiresTree *category.UserCategoryTreeNode `di.inject:"Categories"`
}

func (c *MessageCommand) Name() string {
	return MessageCommandName
}

func (c *MessageCommand) Description() string {
	return "Отправить обращение"
}

func (c *MessageCommand) Handle(message *tgbotapi.Message) error {
	userState, err := c.states.GetState(message.Chat.ID)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to get user state")
	}

	userState.CurrentCategoryNodeId = ""
	err = c.states.SetState(userState)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to set user state")
	}

	return c.tgbot.SendMessageCustom(message.Chat, "Выберите категорию", func(reply *tgbotapi.MessageConfig) {
		reply.ReplyToMessageID = message.MessageID
		reply.ReplyMarkup = c.createCateogoriesReplyMarkup(userState)
	})
}

func (c *MessageCommand) Callback(callbackQuery *tgbotapi.CallbackQuery, data string) error {
	userState, err := c.states.GetState(callbackQuery.Message.Chat.ID)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to get user state")
	}

	var markup tgbotapi.InlineKeyboardMarkup
	var replyText string
	var childFound *category.UserCategoryTreeNode
	currentCategoryNode := c.cateogiresTree.FindNodeById(userState.CurrentCategoryNodeId)
	if data == DataBack {
		if currentCategoryNode.Parent == nil {
			return errorx.AssertionFailed.New("can't go back more than a root")
		}
		childFound = currentCategoryNode.Parent
	} else {
		for _, child := range currentCategoryNode.Children {
			if child.Id == data {
				childFound = child
			}
		}
	}

	if childFound == nil {
		replyText = "Не удалось найти выбранную категорию"
		markup = tgbotapi.NewInlineKeyboardMarkup()
	} else {
		userState.CurrentCategoryNodeId = childFound.Id

		if childFound.Category == nil {
			replyText = "Выберите категорию"
			markup = c.createCateogoriesReplyMarkup(userState)
		} else {
			replyText = fmt.Sprintf(`Выбранная категория: %v
Отправьте фотографии`, childFound.GetFullName())
			markup = c.createCateogoriesReplyMarkup(userState)
			userState.MessageText = childFound.Category.Message
			userState.CurrentCategoryNodeId = childFound.Id
			userState.MessageHandlerName = MessageFormBeanId
		}

		err = c.states.SetState(userState)
		if err != nil {
			return errorx.EnhanceStackTrace(err, "failed to set user state")
		}
	}

	reply := tgbotapi.NewEditMessageTextAndMarkup(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID, replyText, markup)
	_, err = c.tgbot.api.Send(reply)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to send reply")
	}

	return nil
}

func (c *MessageCommand) createCateogoriesReplyMarkup(userState *state.UserState) tgbotapi.InlineKeyboardMarkup {
	result := tgbotapi.NewInlineKeyboardMarkup()

	if userState.CurrentCategoryNodeId != "" {
		backButton := tgbotapi.NewInlineKeyboardButtonData("⬆ Вверх", MessageCommandName+SectionSeparator+DataBack)
		result.InlineKeyboard = append(result.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(backButton))
	}

	currentCategoryNode := c.cateogiresTree.FindNodeById(userState.CurrentCategoryNodeId)

	for _, child := range currentCategoryNode.Children {
		itemButton := tgbotapi.NewInlineKeyboardButtonData(child.Name, MessageCommandName+SectionSeparator+child.Id)
		result.InlineKeyboard = append(result.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(itemButton))
	}

	return result
}

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

	if len(userState.Accounts) == 0 {
		return c.tgbot.SendMessage(message.Chat, `Вы не авторизованы на портале.

Для того, чтобы отправить обращение, нужно авторизоваться на портале с логином и паролем. 

Используйте команду /login для этого.`)
	}

	userState.ClearForm()
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
	currentCategoryNode := c.cateogiresTree.FindNodeByName(userState.GetStringFormField(state.FormFieldCurrentCategoryNode))
	if data == DataBack {
		if currentCategoryNode.Parent == nil {
			return errorx.AssertionFailed.New("can't go back more than a root")
		}
		childFound = currentCategoryNode.Parent
	} else {
		for _, child := range currentCategoryNode.Children {
			if child.Name == data {
				childFound = child
			}
		}
	}

	if childFound == nil {
		replyText = "Не удалось найти выбранную категорию"
		markup = tgbotapi.NewInlineKeyboardMarkup()
	} else {
		userState.SetFormField(state.FormFieldCurrentCategoryNode, childFound.Name)

		if childFound.Category == nil {
			replyText = "Выберите категорию"
			markup = c.createCateogoriesReplyMarkup(userState)
		} else {
			replyText = fmt.Sprintf(`Выбранная категория: %v
Прикрепите фотографии.

Для того, чтобы заменить текст по умолчанию, так же отправьте его в ответ.
Если текст будет содержать "!", то сообщение будет отправлено с повышенным приоритетом, в первую очередь`, childFound.GetFullName())
			markup = c.createCateogoriesReplyMarkup(userState)
			userState.SetFormField(state.FormFieldMessageText, childFound.Category.Message)
			userState.SetFormField(state.FormFieldCurrentCategoryNode, childFound.Name)
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

	if userState.GetStringFormField(state.FormFieldCurrentCategoryNode) != "" {
		backButton := tgbotapi.NewInlineKeyboardButtonData("⬆ Вверх", MessageCommandName+SectionSeparator+DataBack)
		result.InlineKeyboard = append(result.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(backButton))
	}

	currentCategoryNode := c.cateogiresTree.FindNodeByName(userState.GetStringFormField(state.FormFieldCurrentCategoryNode))

	buttonsPerRow := 2

	for i := 0; i < len(currentCategoryNode.Children); i += buttonsPerRow {
		var row []tgbotapi.InlineKeyboardButton
		for j := 0; j < buttonsPerRow; j++ {
			if i+j < len(currentCategoryNode.Children) {
				child := currentCategoryNode.Children[i+j]
				itemButton := tgbotapi.NewInlineKeyboardButtonData(child.Name, MessageCommandName+SectionSeparator+child.Name)
				row = append(row, itemButton)
			}
		}

		result.InlineKeyboard = append(result.InlineKeyboard, row)
	}

	return result
}

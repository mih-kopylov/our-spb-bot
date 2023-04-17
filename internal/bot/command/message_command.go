package command

import (
	_ "embed"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/bot"
	"github.com/mih-kopylov/our-spb-bot/internal/bot/form"
	"github.com/mih-kopylov/our-spb-bot/internal/bot/service"
	"github.com/mih-kopylov/our-spb-bot/internal/category"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
	"strings"
)

const (
	MessageCommandName = "message"
	DataBack           = "back"
)

type MessageCommand struct {
	states         state.States
	service        *service.Service
	cateogiresTree *category.UserCategoryTreeNode
}

func NewMessageCommand(states state.States, service *service.Service, cateogiresTree *category.UserCategoryTreeNode) bot.Command {
	return &MessageCommand{
		states:         states,
		service:        service,
		cateogiresTree: cateogiresTree,
	}
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
		return c.service.SendMessage(message.Chat, `Вы не авторизованы на портале.

Для того, чтобы отправить обращение, нужно авторизоваться на портале с логином и паролем. 

Используйте команду /login для этого.`)
	}

	userState.ClearForm()
	err = c.states.SetState(userState)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to set user state")
	}

	return c.service.SendMessageCustom(message.Chat, "Выберите категорию", func(reply *tgbotapi.MessageConfig) {
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
	currentCategoryNode := c.cateogiresTree.FindNodeById(userState.GetStringFormField(state.FormFieldCurrentCategoryNode))
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
		markup.InlineKeyboard = [][]tgbotapi.InlineKeyboardButton{}
	} else {
		userState.SetFormField(state.FormFieldCurrentCategoryNode, childFound.Id)

		if childFound.Category == nil {
			replyText = strings.TrimSpace(fmt.Sprintf("Выберите категорию\n%v", childFound.GetFullName()))
			markup = c.createCateogoriesReplyMarkup(userState)
		} else {
			replyText = fmt.Sprintf(`Выбранная категория: %v
Прикрепите фотографии.

Для того, чтобы заменить текст по умолчанию, так же отправьте его в ответ.
Если текст будет содержать "!", то сообщение будет отправлено с повышенным приоритетом, в первую очередь`, childFound.GetFullName())
			markup = c.createCateogoriesReplyMarkup(userState)
			userState.SetFormField(state.FormFieldMessageText, childFound.Category.Message)
			userState.SetFormField(state.FormFieldCurrentCategoryNode, childFound.Id)
			userState.MessageHandlerName = form.MessageFormName
		}

		err = c.states.SetState(userState)
		if err != nil {
			return errorx.EnhanceStackTrace(err, "failed to set user state")
		}
	}

	reply := tgbotapi.NewEditMessageTextAndMarkup(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID, replyText, markup)
	err = c.service.Send(reply)
	if err != nil {
		return err
	}

	return nil
}

func (c *MessageCommand) createCateogoriesReplyMarkup(userState *state.UserState) tgbotapi.InlineKeyboardMarkup {
	result := tgbotapi.NewInlineKeyboardMarkup()

	if userState.GetStringFormField(state.FormFieldCurrentCategoryNode) != "" {
		backButton := tgbotapi.NewInlineKeyboardButtonData("⬆ Вверх", MessageCommandName+bot.SectionSeparator+DataBack)
		result.InlineKeyboard = append(result.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(backButton))
	}

	currentCategoryNode := c.cateogiresTree.FindNodeById(userState.GetStringFormField(state.FormFieldCurrentCategoryNode))

	buttonsPerRow := 2

	for i := 0; i < len(currentCategoryNode.Children); i += buttonsPerRow {
		var row []tgbotapi.InlineKeyboardButton
		for j := 0; j < buttonsPerRow; j++ {
			if i+j < len(currentCategoryNode.Children) {
				child := currentCategoryNode.Children[i+j]
				itemButton := tgbotapi.NewInlineKeyboardButtonData(child.Name, MessageCommandName+bot.SectionSeparator+child.Id)
				row = append(row, itemButton)
			}
		}

		result.InlineKeyboard = append(result.InlineKeyboard, row)
	}

	return result
}

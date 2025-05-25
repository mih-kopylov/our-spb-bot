package callback

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/bot"
	"github.com/mih-kopylov/our-spb-bot/internal/bot/service"
	"github.com/mih-kopylov/our-spb-bot/internal/category"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
	"go.uber.org/zap"
)

const (
	MessageCategoryCallbackName = "MessageCategoryCallback"
	DataBack                    = "back"
)

type MessageCategoryCallback struct {
	logger          *zap.Logger
	states          state.States
	service         *service.Service
	categoryService *category.Service
}

func NewMessageCategoryCallback(logger *zap.Logger, states state.States, service *service.Service, categoryService *category.Service) *MessageCategoryCallback {
	return &MessageCategoryCallback{
		logger:          logger,
		states:          states,
		service:         service,
		categoryService: categoryService,
	}
}

func (h *MessageCategoryCallback) Name() string {
	return MessageCategoryCallbackName
}

func (h *MessageCategoryCallback) Handle(callbackQuery *tgbotapi.CallbackQuery, data string) error {
	userState, err := h.states.GetState(callbackQuery.Message.Chat.ID)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to get user state")
	}

	categoriesTree, err := h.categoryService.ParseCategoriesTree(userState.Categories)
	if err != nil {
		return err
	}

	var markup tgbotapi.InlineKeyboardMarkup
	var replyText string
	var childFound *category.UserCategoryTreeNode
	currentCategoryNode := categoriesTree.FindNodeById(userState.GetStringFormField(state.FormFieldCurrentCategoryNode))
	if data == DataBack {
		if currentCategoryNode.Parent == nil {
			return errorx.AssertionFailed.New("can't go back more than a root")
		}
		childFound = currentCategoryNode.Parent
	} else {
		for _, child := range currentCategoryNode.Children {
			if child.Id() == data {
				childFound = child
				break
			}
		}
	}

	if childFound == nil {
		replyText = "Не удалось найти выбранную категорию"
		markup = tgbotapi.NewInlineKeyboardMarkup()
		markup.InlineKeyboard = [][]tgbotapi.InlineKeyboardButton{}
	} else {
		userState.SetFormField(state.FormFieldCurrentCategoryNode, childFound.Id())

		if childFound.Category == nil {
			replyText = strings.TrimSpace(fmt.Sprintf("Выберите категорию\n%v", childFound.GetFullName()))
			markup = h.CreateCategoriesReplyMarkup(userState)
		} else {
			replyText = fmt.Sprintf(`Выбранная категория: %v
Текст по умолчанию: %v

Прикрепите фотографии.

Для того, чтобы заменить текст по умолчанию, так же отправьте его в ответ.
Если текст будет содержать "!", то сообщение будет отправлено с повышенным приоритетом, в первую очередь`,
				childFound.GetFullName(),
				childFound.Category.Message,
			)
			markup = h.CreateCategoriesReplyMarkup(userState)
			userState.SetFormField(state.FormFieldMessageText, childFound.Category.Message)
			userState.SetFormField(state.FormFieldCurrentCategoryNode, childFound.Id())
		}

		err = h.states.SetState(userState)
		if err != nil {
			return errorx.EnhanceStackTrace(err, "failed to set user state")
		}
	}

	reply := tgbotapi.NewEditMessageTextAndMarkup(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID, replyText, markup)
	err = h.service.Send(reply)
	if err != nil {
		return err
	}

	return nil
}

func (h *MessageCategoryCallback) CreateCategoriesReplyMarkup(userState *state.UserState) tgbotapi.InlineKeyboardMarkup {
	result := tgbotapi.NewInlineKeyboardMarkup()
	result.InlineKeyboard = [][]tgbotapi.InlineKeyboardButton{}

	currentCategoryNodeId := userState.GetStringFormField(state.FormFieldCurrentCategoryNode)
	if currentCategoryNodeId != "" {
		backButton := tgbotapi.NewInlineKeyboardButtonData("⬆ Вверх", MessageCategoryCallbackName+bot.CallbackSectionSeparator+DataBack)
		result.InlineKeyboard = append(result.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(backButton))
	}

	categoriesTree, err := h.categoryService.ParseCategoriesTree(userState.Categories)
	if err != nil {
		h.logger.Error("failed to parse user categories",
			zap.Error(err))
	} else {
		currentCategoryNode := categoriesTree.FindNodeById(currentCategoryNodeId)
		if currentCategoryNode == nil {
			h.logger.Error("can't find current category node by id",
				zap.String("id", currentCategoryNodeId))
		} else {
			buttonsPerRow := 2

			for i := 0; i < len(currentCategoryNode.Children); i += buttonsPerRow {
				var row []tgbotapi.InlineKeyboardButton
				for j := 0; j < buttonsPerRow; j++ {
					if i+j < len(currentCategoryNode.Children) {
						child := currentCategoryNode.Children[i+j]
						itemButton := tgbotapi.NewInlineKeyboardButtonData(child.Name, MessageCategoryCallbackName+bot.CallbackSectionSeparator+child.Id())
						row = append(row, itemButton)
					}
				}

				result.InlineKeyboard = append(result.InlineKeyboard, row)
			}
		}
	}
	return result
}

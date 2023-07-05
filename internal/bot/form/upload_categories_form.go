package form

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/category"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
	"github.com/mih-kopylov/our-spb-bot/pkg/tgbot"
	"go.uber.org/zap"
)

const (
	UploadCategoriesFormName = "UploadCategoriesForm"
)

type UploadCategoriesForm struct {
	logger          *zap.Logger
	stateManager    state.Manager
	service         *tgbot.Service
	categoryService *category.Service
}

func NewUploadCategoriesForm(logger *zap.Logger, stateManager state.Manager, service *tgbot.Service, categoryService *category.Service) tgbot.Form {
	return &UploadCategoriesForm{
		logger:          logger,
		stateManager:    stateManager,
		service:         service,
		categoryService: categoryService,
	}
}

func (f *UploadCategoriesForm) Name() string {
	return UploadCategoriesFormName
}

func (f *UploadCategoriesForm) Handle(message *tgbotapi.Message) error {
	userState, err := f.stateManager.GetState(message.Chat.ID)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to get user state")
	}

	if message.Document == nil {
		_, err := f.service.SendMessageCustom(message.Chat, "В сообщении не найден документ", func(reply *tgbotapi.MessageConfig) {
			reply.ReplyToMessageID = message.MessageID
		})
		return err
	}

	fileContent, err := f.service.DownloadFile(message.Document.FileID)
	if err != nil {
		return err
	}

	_, err = f.categoryService.ParseCategoriesTree(string(fileContent))
	if err != nil {
		f.logger.Error("can't parse document")
		_, err = f.service.SendMessageCustom(message.Chat, "Документ должен быть в yaml формате", func(reply *tgbotapi.MessageConfig) {
			reply.ReplyToMessageID = message.MessageID
		})
		if err != nil {
			return err
		}
	}

	userState.Categories = string(fileContent)
	userState.MessageHandlerName = ""
	err = f.stateManager.SetState(userState)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to set user state")
	}

	return f.service.SendMessage(message.Chat, "Категории обновлены")
}

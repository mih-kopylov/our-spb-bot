package app

import (
	"github.com/mih-kopylov/our-spb-bot/internal/bot"
	"github.com/mih-kopylov/our-spb-bot/internal/bot/api"
	"github.com/mih-kopylov/our-spb-bot/internal/bot/callback"
	"github.com/mih-kopylov/our-spb-bot/internal/bot/command"
	"github.com/mih-kopylov/our-spb-bot/internal/bot/form"
	"github.com/mih-kopylov/our-spb-bot/internal/bot/service"
	"github.com/mih-kopylov/our-spb-bot/internal/category"
	"github.com/mih-kopylov/our-spb-bot/internal/config"
	"github.com/mih-kopylov/our-spb-bot/internal/info"
	"github.com/mih-kopylov/our-spb-bot/internal/queue"
	"github.com/mih-kopylov/our-spb-bot/internal/spb"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
	"github.com/mih-kopylov/our-spb-bot/internal/storage"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

func RunApplication(version string, commit string) error {
	logrus.SetLevel(logrus.DebugLevel)
	fx.New(createApp(version, commit)).Run()

	return nil
}

func createApp(version string, commit string) fx.Option {
	return fx.Options(
		fx.Supply(info.NewInfo(version, commit)),
		fx.Provide(
			config.NewConfig,
			api.NewApi,
			storage.NewFirebaseStorage,
			fx.Annotate(
				state.NewFirebaseState, fx.As(new(state.States)),
			),
			fx.Annotate(
				queue.NewFirebaseQueue, fx.As(new(queue.MessageQueue)),
			),
			category.NewUserCategoryTreeNode,

			service.NewService,
			fx.Annotate(
				bot.NewTgBot, fx.ParamTags(``, ``, `group:"commands"`, `group:"callbacks"`, `group:"forms"`),
			),
			queue.NewMessageSender,
			fx.Annotate(
				spb.NewReqClient, fx.As(new(spb.Client)),
			),
			//commands
			fx.Annotate(
				command.NewStartCommand, fx.ResultTags(`group:"commands"`),
			),
			fx.Annotate(
				command.NewStatusCommand, fx.ResultTags(`group:"commands"`),
			),
			fx.Annotate(
				command.NewMessageCommand, fx.ResultTags(`group:"commands"`),
			),
			fx.Annotate(
				command.NewLoginCommand, fx.ResultTags(`group:"commands"`),
			),
			fx.Annotate(
				command.NewResetStatusCommand, fx.ResultTags(`group:"commands"`),
			),
			fx.Annotate(
				command.NewFileIdCommand, fx.ResultTags(`group:"commands"`),
			),
			//callbacks
			callback.NewMessageCategoryCallback,
			fx.Annotate(
				func(cb *callback.MessageCategoryCallback) bot.Callback {
					return cb
				}, fx.ResultTags(`group:"callbacks"`),
			),
			//forms
			fx.Annotate(
				form.NewMessageForm, fx.ResultTags(`group:"forms"`),
			),
			fx.Annotate(
				form.NewLoginForm, fx.ResultTags(`group:"forms"`),
			),
			fx.Annotate(
				form.NewPasswordForm, fx.ResultTags(`group:"forms"`),
			),
			fx.Annotate(
				form.NewFileIdForm, fx.ResultTags(`group:"forms"`),
			),
		),
		fx.Invoke(func(bot *bot.TgBot) error {
			return bot.Start()
		}),
		fx.Invoke(func(sender *queue.MessageSender) error {
			return sender.Start()
		}),
	)
}

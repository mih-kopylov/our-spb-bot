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
	"github.com/mih-kopylov/our-spb-bot/internal/log"
	"github.com/mih-kopylov/our-spb-bot/internal/migration"
	"github.com/mih-kopylov/our-spb-bot/internal/queue"
	"github.com/mih-kopylov/our-spb-bot/internal/spb"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
	"github.com/mih-kopylov/our-spb-bot/internal/storage"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func RunApplication(version string, commit string) error {
	fx.New(createApp(version, commit)).Run()

	return nil
}

func createApp(version string, commit string) fx.Option {
	return fx.Options(
		fx.WithLogger(func(logger *zap.Logger) fxevent.Logger {
			result := fxevent.ZapLogger{Logger: logger}
			result.UseLogLevel(zap.DebugLevel)
			return &result
		}),
		fx.Supply(info.NewInfo(version, commit)),
		fx.Provide(
			log.NewLogger,
			config.NewConfig,
			api.NewApi,
			storage.NewFirebaseStorage,
			fx.Annotate(
				state.NewFirebaseState, fx.As(new(state.States)),
			),
			fx.Annotate(
				queue.NewFirebaseQueue, fx.As(new(queue.MessageQueue)),
			),
			category.NewService,

			service.NewService,
			fx.Annotate(
				bot.NewTgBot, fx.ParamTags(``, ``, ``, `group:"commands"`, `group:"callbacks"`, `group:"forms"`),
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
			fx.Annotate(
				command.NewSettingsCommand, fx.ResultTags(`group:"commands"`),
			),
			//callbacks
			callback.NewMessageCategoryCallback,
			fx.Annotate(
				func(cb *callback.MessageCategoryCallback) bot.Callback {
					return cb
				}, fx.ResultTags(`group:"callbacks"`),
			),
			callback.NewDeleteMessageCallback,
			fx.Annotate(
				func(cb *callback.DeleteMessageCallback) bot.Callback {
					return cb
				}, fx.ResultTags(`group:"callbacks"`),
			),
			callback.NewSettingsCallback,
			fx.Annotate(
				func(cb *callback.SettingsCallback) bot.Callback {
					return cb
				}, fx.ResultTags(`group:"callbacks"`),
			),
			callback.NewSettingsCategoriesCallback,
			fx.Annotate(
				func(cb *callback.SettingsCategoriesCallback) bot.Callback {
					return cb
				}, fx.ResultTags(`group:"callbacks"`),
			),
			callback.NewSettingsAccountsCallback,
			fx.Annotate(
				func(cb *callback.SettingsAccountsCallback) bot.Callback {
					return cb
				}, fx.ResultTags(`group:"callbacks"`),
			),
			callback.NewDeletePhotoCallback,
			fx.Annotate(
				func(cb *callback.DeletePhotoCallback) bot.Callback {
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
			fx.Annotate(
				form.NewUploadCategoriesForm, fx.ResultTags(`group:"forms"`),
			),
			fx.Annotate(
				form.NewAccountTimeForm, fx.ResultTags(`group:"forms"`),
			),
			//migrations
			fx.Annotate(
				migration.NewMigrations, fx.ParamTags(``, `group:"migrations"`),
			),
			fx.Annotate(
				migration.NewAccountTimeMigration, fx.ResultTags(`group:"migrations"`),
			),
		),

		fx.Invoke(func(migrations *migration.Migrations) error {
			return migrations.RunAll()
		}),

		fx.Invoke(func(bot *bot.TgBot) error {
			return bot.Start()
		}),

		fx.Invoke(func(sender *queue.MessageSender) error {
			return sender.Start()
		}),
	)
}

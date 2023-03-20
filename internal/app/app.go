package app

import (
	"github.com/goioc/di"
	"github.com/mih-kopylov/our-spb-bot/internal/bot"
	"github.com/mih-kopylov/our-spb-bot/internal/category"
	"github.com/mih-kopylov/our-spb-bot/internal/config"
	"github.com/mih-kopylov/our-spb-bot/internal/info"
	"github.com/mih-kopylov/our-spb-bot/internal/queue"
	"github.com/mih-kopylov/our-spb-bot/internal/spb"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
	"github.com/mih-kopylov/our-spb-bot/internal/storage"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

func RunApplication(version string, commit string) error {
	logrus.SetLevel(logrus.DebugLevel)
	info.RegisterBean(version, commit)
	conf := config.RegisterBean()
	storage.RegisterBean(conf)
	state.RegisterBean()
	category.RegisterBean()
	bot.RegisterApiBean(conf)
	bot.RegisterBotBean()
	queue.RegisterQueueBean()
	queue.RegisterSenderBean()
	spb.RegisterBean(conf)
	lo.Must0(di.InitializeContainer())

	tgBot := di.GetInstance(bot.TgBotBeanId).(*bot.TgBot)
	tgBot.RegisterBotCommands()
	tgBot.ProcessUpdates()

	return nil
}

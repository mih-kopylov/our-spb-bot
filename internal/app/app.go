package app

import (
	"github.com/goioc/di"
	"github.com/mih-kopylov/our-spb-bot/internal/bot"
	"github.com/mih-kopylov/our-spb-bot/internal/category"
	"github.com/mih-kopylov/our-spb-bot/internal/config"
	_ "github.com/mih-kopylov/our-spb-bot/internal/config"
	"github.com/mih-kopylov/our-spb-bot/internal/queue"
	"github.com/mih-kopylov/our-spb-bot/internal/spb"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
	_ "github.com/mih-kopylov/our-spb-bot/internal/state"
	"github.com/samber/lo"
)

func RunApplication() error {
	conf := config.RegisterBean()
	state.RegisterBean()
	category.RegisterBean()
	bot.RegisterApiBean(conf)
	bot.RegisterBotBean()
	queue.RegisterBean()
	spb.RegisterBean(conf)
	lo.Must0(di.InitializeContainer())

	tgBot := di.GetInstance(bot.TgBotBeanId).(*bot.TgBot)
	tgBot.RegisterBotCommands()
	tgBot.ProcessUpdates()

	return nil
}

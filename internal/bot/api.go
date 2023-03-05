package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/goioc/di"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/config"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

const (
	TgApiBeanId = "TgApi"
)

func RegisterApiBean(conf *config.Config) {
	api := lo.Must(createApi(conf))
	_ = lo.Must(di.RegisterBeanInstance(TgApiBeanId, api))
}

func createApi(conf *config.Config) (*tgbotapi.BotAPI, error) {
	api, err := tgbotapi.NewBotAPI(conf.Token)
	if err != nil {
		return nil, errorx.EnhanceStackTrace(err, "failed to create bot")
	}

	api.Debug = true
	err = tgbotapi.SetLogger(&LorRusLogger{})
	return api, nil
}

type LorRusLogger struct{}

func (l *LorRusLogger) Println(v ...interface{}) {
	logrus.Infoln(v...)
}

func (l *LorRusLogger) Printf(format string, v ...interface{}) {
	logrus.Infof(format, v...)
}

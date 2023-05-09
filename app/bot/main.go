package main

import (
	"github.com/mih-kopylov/our-spb-bot/internal/app"
	"github.com/mih-kopylov/our-spb-bot/internal/log"
	"go.uber.org/zap"
)

var (
	version string
	commit  string
)

func main() {
	err := app.RunApplication(version, commit)
	if err != nil {
		log.NewLogger().With(zap.Error(err)).Fatal("")
	}
}

package main

import (
	"github.com/mih-kopylov/our-spb-bot/internal/app"
	"github.com/sirupsen/logrus"
)

var (
	version string
	commit  string
)

func main() {
	err := app.RunApplication(version, commit)
	if err != nil {
		logrus.Fatal(err)
	}
}

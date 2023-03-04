package main

import (
	"github.com/mih-kopylov/our-spb-bot/internal/app"
	"github.com/sirupsen/logrus"
)

func main() {
	err := app.RunApplication()
	if err != nil {
		logrus.Fatal(err)
	}
}

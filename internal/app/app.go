package app

import (
	"log"

	"github.com/DarkhanShakhan/telegram-bot-template/internal/config"
	"github.com/DarkhanShakhan/telegram-bot-template/internal/ports/telegram"
)

type App struct {
	config    *config.Config
	botServer *telegram.Server
}

func InitApp() *App {
	a := newApp()
	var err error
	for _, init := range []func() error{
		a.initConfig,
		a.initTelegramBot,
	} {
		if err = init(); err != nil {
			log.Fatal(err)
		}
	}
	return a
}

func newApp() *App {
	return &App{}
}

func (a *App) initConfig() error {
	config := config.New()
	if err := config.ParseConfig(); err != nil {
		return err
	}
	a.config = config
	return nil
}

func (a *App) initTelegramBot() error {
	server, err := telegram.New(a.config.TelegramToken)
	if err != nil {
		return err
	}
	a.botServer = server
	return nil
}

func (a *App) Start() {
	log.Println("starting telegram bot")
	a.botServer.Start()
}

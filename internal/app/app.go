package app

import (
	"context"
	"log"
	"time"

	"github.com/DarkhanShakhan/telegram-bot-template/internal/cache/matches"
	"github.com/DarkhanShakhan/telegram-bot-template/internal/cache/users"
	"github.com/DarkhanShakhan/telegram-bot-template/internal/config"
	"github.com/DarkhanShakhan/telegram-bot-template/internal/ports/telegram"
	matchesR "github.com/DarkhanShakhan/telegram-bot-template/internal/repository/matches"
	"github.com/DarkhanShakhan/telegram-bot-template/internal/service/match"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	config     *config.Config
	botServer  *telegram.Server
	cache      matches.Cache
	usersCache users.Cache
	service    match.Service
	pool       *pgxpool.Pool
}

func InitApp() *App {
	a := newApp()
	var err error
	for _, init := range []func() error{
		a.initConfig,
		a.initCache,
		a.initDB,
		a.initService,
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

func (a *App) initCache() error {
	a.cache = matches.New("matches", 15*time.Minute, 20*time.Minute)
	a.usersCache = users.New("users", 15*time.Minute, 20*time.Minute)
	return nil
}

func (a *App) initDB() error {
	config, err := pgxpool.ParseConfig(a.config.PostgresDSN)
	if err != nil {
		return err
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return err
	}
	a.pool = pool
	return a.pool.Ping(context.Background())
}

func (a *App) initService() error {
	repository := matchesR.New(a.pool)
	a.service = match.New(repository)
	return nil
}

func (a *App) initTelegramBot() error {
	server, err := telegram.New(a.config.TelegramToken, a.cache, a.usersCache, a.service)
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

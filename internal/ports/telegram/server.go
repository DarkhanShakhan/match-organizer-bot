package telegram

import (
	"github.com/DarkhanShakhan/telegram-bot-template/internal/cache/matches"
	"github.com/DarkhanShakhan/telegram-bot-template/internal/cache/users"
	"github.com/DarkhanShakhan/telegram-bot-template/internal/ports/telegram/router"
	"github.com/DarkhanShakhan/telegram-bot-template/internal/service/match"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Server struct {
	bot          *tgbotapi.BotAPI
	matchesCache matches.Cache
	usersCache   users.Cache
	matchService match.Service
}

func New(token string, matchesCache matches.Cache, userCache users.Cache, matchService match.Service) (*Server, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return &Server{bot: bot, matchesCache: matchesCache, matchService: matchService, usersCache: userCache}, nil
}

func (s *Server) Start() {
	u := tgbotapi.UpdateConfig{
		Timeout: 60,
	}
	routerHandler := router.NewRouter(s.bot, s.matchesCache, s.usersCache, s.matchService)

	for update := range s.bot.GetUpdatesChan(u) {
		go routerHandler.HandleUpdate(update)
	}
}

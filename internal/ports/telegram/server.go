package telegram

import (
	"github.com/DarkhanShakhan/telegram-bot-template/internal/ports/telegram/router"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Server struct {
	bot *tgbotapi.BotAPI
}

func New(token string) (*Server, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	//TODO: config bot
	return &Server{bot: bot}, nil
}

func (s *Server) Start() {
	u := tgbotapi.UpdateConfig{
		Timeout: 60,
	}
	//FIXME:
	routerHandler := router.NewRouter(s.bot)

	for update := range s.bot.GetUpdatesChan(u) {
		routerHandler.HandleUpdate(update) //FIXME: is it safe?
	}
}

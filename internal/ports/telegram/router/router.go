package router

import (
	"log"
	"runtime/debug"

	"github.com/DarkhanShakhan/telegram-bot-template/internal/ports/telegram/path"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Commander interface {
	HandleCallback(callback *tgbotapi.CallbackQuery, callbackPath path.CallbackPath)
	HandleCommand(callback *tgbotapi.Message, commandPath path.CommandPath)
}

type Router interface {
	HandleUpdate(update tgbotapi.Update)
}

type router struct {
	bot           *tgbotapi.BotAPI
	demoCommander Commander
}

func NewRouter(bot *tgbotapi.BotAPI) Router {
	return &router{
		bot: bot,
	}
}

func (r *router) HandleUpdate(update tgbotapi.Update) {
	defer func() {
		if panicValue := recover(); panicValue != nil {
			log.Printf("recovered from panic: %v\n%v", panicValue, string(debug.Stack()))
		}
	}()
	switch {
	case update.CallbackQuery != nil:
		r.handleCallback(update.CallbackQuery)
	case update.Message != nil:
		r.handleMessage(update.Message)
	}
}

func (r *router) handleCallback(callback *tgbotapi.CallbackQuery) {
	callbackPath, err := path.ParseCallback(callback.Data)
	if err != nil {
		//TODO: log errors
		log.Println(err)
		return
	}
	msg := tgbotapi.NewMessage(callback.Message.Chat.ID, callbackPath.Domain)
	_, err = r.bot.Send(msg)
	if err != nil {
		log.Println(err)
	}
}

func (r *router) handleMessage(msg *tgbotapi.Message) {
	_, _ = r.bot.Send(tgbotapi.NewMessage(msg.From.ID, "hello"))
}

package router

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/DarkhanShakhan/telegram-bot-template/internal/cache/matches"
	"github.com/DarkhanShakhan/telegram-bot-template/internal/cache/users"
	"github.com/DarkhanShakhan/telegram-bot-template/internal/domain/entity"
	"github.com/DarkhanShakhan/telegram-bot-template/internal/domain/enum"
	"github.com/DarkhanShakhan/telegram-bot-template/internal/service/match"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Router interface {
	HandleUpdate(update tgbotapi.Update)
}

type router struct {
	bot       *tgbotapi.BotAPI
	cache     matches.Cache
	userCache users.Cache
	service   match.Service
}

func NewRouter(bot *tgbotapi.BotAPI, cache matches.Cache, userCache users.Cache, service match.Service) Router {
	return &router{
		bot:       bot,
		cache:     cache,
		service:   service,
		userCache: userCache,
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
	status, err := r.cache.GetStatus(callback.From.UserName)
	if err == nil {
		r.createMatchCallback(callback, status)
		return
	}
	callbacks := strings.Split(callback.Data, "-")
	if len(callbacks) == 0 {
		return
	}
	switch callbacks[0] {
	case "add_members":
		msg := tgbotapi.NewMessage(callback.From.ID, "Добавить в команду:")
		id, _ := strconv.Atoi(callbacks[1])
		match, err := r.service.GetMatchByMatchID(context.Background(), int64(id))
		if err != nil {
			log.Println(err)
			return
		}
		msg.ReplyMarkup = matchTeamsKeyboard(match.Teams)
		r.bot.Send(msg)
	case "cancel_match":
	case "add_team_members":
		teamID, _ := strconv.Atoi(callbacks[1])
		r.userCache.SetTeamID(callback.From.UserName, int64(teamID))
		r.userCache.SetStatus(callback.From.UserName, users.StatusAddTeamMembers)
		msg := tgbotapi.NewMessage(callback.From.ID, `Отправьте юзернеймы тех, кого хотите добавить, 
		через пробел и с "@" в начале`)
		r.bot.Send(msg)
	}
}

func matchTeamsKeyboard(teams []*entity.Team) tgbotapi.InlineKeyboardMarkup {
	rows := [][]tgbotapi.InlineKeyboardButton{}
	row := []tgbotapi.InlineKeyboardButton{}
	for ix, team := range teams {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(color[team.Name], fmt.Sprintf("add_team_members-%d", team.ID)))
		if ix%2 == 1 {
			rows = append(rows, row)
			row = []tgbotapi.InlineKeyboardButton{}
		}
	}
	if len(row) != 0 {
		rows = append(rows, row)
	}
	mrkup := tgbotapi.NewInlineKeyboardMarkup(rows...)
	return mrkup
}

func (r *router) createMatchCallback(callback *tgbotapi.CallbackQuery, status matches.Status) {
	switch status {
	case matches.StatusNew:
		r.cache.SetSportType(callback.From.UserName, enum.SportType(callback.Data))
		r.bot.Send(tgbotapi.NewMessage(callback.From.ID, "Где будет матч?"))
		return
	case matches.StatusLocation:
		r.cache.SetDay(callback.From.UserName, enum.MatchDay(callback.Data))
		msg := tgbotapi.NewMessage(callback.From.ID, "В какое время будет матч?")
		msg.ReplyMarkup = matchTimeKeyboard(enum.MatchDay(callback.Data))
		r.bot.Send(msg)
	case matches.StatusMatchDay:
		hour, _ := strconv.Atoi(callback.Data)
		r.cache.SetTime(callback.From.UserName, time.Duration(hour)*time.Hour)
		msg := tgbotapi.NewMessage(callback.From.ID, "Длительность матча?")
		msg.ReplyMarkup = matchDurationKeyboard
		r.bot.Send(msg)
	case matches.StatusStartTime:
		mins, _ := strconv.Atoi(callback.Data)
		r.cache.SetDuration(callback.From.UserName, time.Duration(mins)*time.Minute)
		msg := tgbotapi.NewMessage(callback.From.ID, "Сколько человек в команде?")
		msg.ReplyMarkup = matchTeamSizeKeyboard
		r.bot.Send(msg)
	case matches.StatusDuration:
		teamSize, _ := strconv.Atoi(callback.Data)
		r.cache.SetTeamSize(callback.From.UserName, int64(teamSize))
		msg := tgbotapi.NewMessage(callback.From.ID, "Сколько команд?")
		msg.ReplyMarkup = matchTeamCountKeyboard
		r.bot.Send(msg)
	case matches.StatusTeamSize:
		teamCount, _ := strconv.Atoi(callback.Data)
		r.cache.SetTeamCount(callback.From.UserName, int64(teamCount))
		msg := tgbotapi.NewMessage(callback.From.ID, "Сколько стоит аренда?")
		r.bot.Send(msg)
	case matches.StatusRent:
		isPrivate := false
		if callback.Data == "закрытый" {
			isPrivate = true
		}
		r.cache.SetPrivate(callback.From.UserName, isPrivate)
		msg := tgbotapi.NewMessage(callback.From.ID, "Организовать матч?")
		msg.ReplyMarkup = matchConfirmKeyboard
		r.bot.Send(msg)
	case matches.StatusPrivate:
		confirmed := true
		if callback.Data == "отменить" {
			confirmed = false
		}
		if confirmed {
			match, _ := r.cache.GetMatch(callback.From.UserName)
			user, err := r.service.GetUserByUsername(context.Background(), callback.From.UserName)
			if err != nil {
				user, err = r.service.CreateUser(context.Background(), &entity.User{
					Name:     callback.From.FirstName,
					Username: callback.From.UserName,
					ChatID:   int(callback.From.ID),
				})
				if err != nil {
					log.Println(err)
					return
				}
			}
			match.OrganizerID = user.ID
			match, err = r.service.CreateMatch(context.Background(), match)
			if err != nil {
				log.Println(err)
				return
			}
			match.OrganizerUsername = user.Username
			msg := tgbotapi.NewMessage(callback.From.ID, fmt.Sprint(match))
			msg.ReplyMarkup = matchOptionsKeyboard(match.ID)
			r.bot.Send(msg)
		}
		r.cache.DeleteMatch(callback.From.UserName)
	default:
	}
}

func matchOptionsKeyboard(matchID int64) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Добавить участников", fmt.Sprintf("add_members-%d", matchID)),
			tgbotapi.NewInlineKeyboardButtonData("Отменить матч", fmt.Sprintf("cancel_match-%d", matchID)),
		),
	)
}

// var matchOptionsKeyboard = tgbotapi.NewInlineKeyboardMarkup(
// 	tgbotapi.NewInlineKeyboardRow(
// 		tgbotapi.NewInlineKeyboardButtonData("", "организовать"),
// 		tgbotapi.NewInlineKeyboardButtonData("Отменить", "отменить"),
// 	),
// )

func (r *router) handleMessage(msg *tgbotapi.Message) {
	if msg.IsCommand() {
		r.handleCommand(msg)
		return
	}
	status, err := r.cache.GetStatus(msg.From.UserName)
	if err == nil {
		switch status {
		case matches.StatusSportType:
			r.cache.SetLocation(msg.From.UserName, msg.Text)
			msgToSend := tgbotapi.NewMessage(msg.From.ID, "В какой день будет матч?")
			msgToSend.ReplyMarkup = matchDayKeyboard
			r.bot.Send(msgToSend)
		case matches.StatusTeamCount:
			rent, _ := strconv.Atoi(msg.Text)
			r.cache.SetRent(msg.From.UserName, int64(rent))
			msgToSend := tgbotapi.NewMessage(msg.From.ID, "Закрытый или открытый матч?")
			msgToSend.ReplyMarkup = matchPrivateKeyboard
			r.bot.Send(msgToSend)
		}
	}
	userStatus := r.userCache.GetStatus(msg.From.UserName)
	switch userStatus {
	case users.StatusAddTeamMembers:
		r.addTeamMembers(msg)
	}
}

func (r *router) addTeamMembers(msg *tgbotapi.Message) {
	members := strings.Fields(msg.Text)
	user, ok := r.userCache.GetUser(msg.From.UserName)
	if !ok {
		log.Println("user not found in cache")
		return
	}
	if err := r.service.AddTeamMembers(context.Background(), user.TeamID, members); err != nil {
		log.Println(err)
	}
	users := r.service.GetUsersByUsernames(context.Background(), members)
	for _, user := range users {
		r.bot.Send(tgbotapi.NewMessage(int64(user.ChatID), "Вас приглашают на матч"))
	}
	r.userCache.SetStatus(user.Username, 0)
	//TODO:respond successfully
}

func (r *router) handleCommand(msg *tgbotapi.Message) {
	cmd := msg.Command()
	switch cmd {
	case "create_match":
		r.cache.SetMatch(msg.From.UserName)
		msgToSend := tgbotapi.NewMessage(msg.From.ID, "Выберите вид спорта")
		msgToSend.ReplyMarkup = sportTypeKeyboard
		r.bot.Send(msgToSend)
	}
}

var sportTypeKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("football", "football"),
		tgbotapi.NewInlineKeyboardButtonData("volleyball", "volleyball"),
		tgbotapi.NewInlineKeyboardButtonData("basketball", "basketball"),
	),
)

var matchDayKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("сегодня", "today"),
		tgbotapi.NewInlineKeyboardButtonData("завтра", "tomorrow"),
	),
)

func matchTimeKeyboard(day enum.MatchDay) tgbotapi.InlineKeyboardMarkup {
	switch day {
	case enum.MatchDayToday:
		h := time.Now().Hour()
		switch {
		case h < 9:
			return tgbotapi.NewInlineKeyboardMarkup(
				matchTimeRows...,
			)
		case h < 14:
			return tgbotapi.NewInlineKeyboardMarkup(
				matchTimeRows[1:]...,
			)

		case h < 19:
			return tgbotapi.NewInlineKeyboardMarkup(
				matchTimeRows[2:]...,
			)
		default:
			return tgbotapi.NewInlineKeyboardMarkup()
		}
	default:
		return tgbotapi.NewInlineKeyboardMarkup(
			matchTimeRows...,
		)
	}
}

var matchTimeRows = [][]tgbotapi.InlineKeyboardButton{
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("10:00", "10"),
		tgbotapi.NewInlineKeyboardButtonData("11:00", "11"),
		tgbotapi.NewInlineKeyboardButtonData("12:00", "12"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("15:00", "15"),
		tgbotapi.NewInlineKeyboardButtonData("16:00", "16"),
		tgbotapi.NewInlineKeyboardButtonData("17:00", "17"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("20:00", "20"),
		tgbotapi.NewInlineKeyboardButtonData("21:00", "21"),
		tgbotapi.NewInlineKeyboardButtonData("22:00", "22"),
	),
}

var matchDurationKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("1 час", "60"),
		tgbotapi.NewInlineKeyboardButtonData("1.5 часа", "90"),
		tgbotapi.NewInlineKeyboardButtonData("2 часа", "120"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("2.5 часа", "150"),
		tgbotapi.NewInlineKeyboardButtonData("3 часа", "180"),
		tgbotapi.NewInlineKeyboardButtonData("3.5 часа", "210"),
	),
)

var matchTeamSizeKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("4", "4"),
		tgbotapi.NewInlineKeyboardButtonData("5", "5"),
		tgbotapi.NewInlineKeyboardButtonData("6", "6"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("7", "7"),
		tgbotapi.NewInlineKeyboardButtonData("8", "8"),
		tgbotapi.NewInlineKeyboardButtonData("9", "9"),
	),
)

var matchTeamCountKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("2", "2"),
		tgbotapi.NewInlineKeyboardButtonData("3", "3"),
		tgbotapi.NewInlineKeyboardButtonData("4", "4"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("5", "5"),
		tgbotapi.NewInlineKeyboardButtonData("6", "6"),
		tgbotapi.NewInlineKeyboardButtonData("7", "7"),
	),
)

var matchPrivateKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Открытый", "открытый"),
		tgbotapi.NewInlineKeyboardButtonData("Закрытый", "закрытый"),
	),
)

var matchConfirmKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Организовать", "организовать"),
		tgbotapi.NewInlineKeyboardButtonData("Отменить", "отменить"),
	),
)

var color = map[string]string{"red": "🟥", "blue": "🟦", "green": "🟩", "yellow": "🟨", "purple": "🟪", "black": "⬛️", "brown": "🟫"}

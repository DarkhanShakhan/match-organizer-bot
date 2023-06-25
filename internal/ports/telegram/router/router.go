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
		matchID, _ := strconv.Atoi(callbacks[1])
		match, err := r.service.GetMatchByMatchID(context.Background(), int64(matchID))
		if err != nil {
			log.Println(err)
			return
		}
		r.service.CancelMatch(context.Background(), int64(matchID))
		r.bot.Send(tgbotapi.NewMessage(callback.From.ID, "Вы отменили матч"))
		for _, team := range match.Teams {
			for _, member := range team.Members {
				//TODO: return fees to paid members
				r.bot.Send(tgbotapi.NewMessage(int64(member.ChatID), fmt.Sprintf("Матч #%d отменен", matchID)))
			}
		}

	case "add_team_members":
		teamID, _ := strconv.Atoi(callbacks[1])
		r.userCache.SetTeamID(callback.From.UserName, int64(teamID))
		r.userCache.SetStatus(callback.From.UserName, users.StatusAddTeamMembers)
		msg := tgbotapi.NewMessage(callback.From.ID, `Отправьте юзернеймы тех, кого хотите добавить, 
		через пробел и с "@" в начале`)
		r.bot.Send(msg)
	case "get_matches_by_sport":
		matches, err := r.service.GetOpenMatchesBySport(context.Background(), enum.SportType(callbacks[1]))
		if err != nil {
			log.Println(err)
			return
		}
		r.bot.Send(tgbotapi.NewMessage(callback.From.ID, fmt.Sprintf(`🔜 Ближайшие матчи по %sу

		`, callbacks[1])))
		if len(matches) == 0 {
			r.bot.Send(tgbotapi.NewMessage(callback.From.ID, `😥 К сожалению, матчей нет`))
			return
		}
		for _, m := range matches {
			msg := tgbotapi.NewMessage(callback.From.ID,
				fmt.Sprintf(`Матч #%d - Начало %d/%d %d:00(%.1f часа) - %d тг/чел - Осталось %d мест`,
					m.ID, m.StartAt.Day(), m.StartAt.Month(), m.StartAt.Hour(), float64(m.FinishAt.Sub(m.StartAt).Minutes())/60.0,
					m.Rent/(m.TeamCount*m.TeamSize), (m.TeamCount*m.TeamSize)-m.MembersCount))
			msg.ReplyMarkup = matchMoreKeyboard(m.ID)
			r.bot.Send(msg)
		}
	case "get_match_by_id":
		matchID, _ := strconv.Atoi(callbacks[1])
		match, err := r.service.GetMatchByMatchID(context.Background(), int64(matchID))
		if err != nil {
			log.Println(err)
			return
		}
		msg := tgbotapi.NewMessage(callback.From.ID,
			fmt.Sprint(match),
		)
		rows := []tgbotapi.InlineKeyboardButton{}
		if match.OrganizerUsername == callback.From.UserName {
			rows = append(rows, tgbotapi.NewInlineKeyboardButtonData(
				"Отменить матч",
				fmt.Sprintf("cancel_match-%d", matchID)))
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardButtonData(
			"Записаться на матч",
			fmt.Sprintf("signup_match-%d", matchID)),
		)
		for _, team := range match.Teams {
			for _, member := range team.Members {
				if callback.From.UserName == member.Username {
					rows = append(rows[:len(rows)-1],
						tgbotapi.NewInlineKeyboardButtonData("Отменить участие", fmt.Sprintf("signout_match-%d", matchID)),
					)
					if !member.Confirmed {
						rows = append(rows, tgbotapi.NewInlineKeyboardButtonData(
							"Подтвердить участие", fmt.Sprintf("confirm_match-%d", matchID),
						))
					}
					if !member.Paid {
						rows = append(rows, tgbotapi.NewInlineKeyboardButtonData(
							"Оплатить взнос", fmt.Sprintf("pay_match-%d", matchID),
						))
					}
					break
				}

			}

		}
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows)
		r.bot.Send(msg)
	case "pay_match":
	case "confirm_match":
		matchID, _ := strconv.Atoi(callbacks[1])
		match, _ := r.service.GetMatchByMatchID(context.TODO(), int64(matchID))
		organizer, _ := r.service.GetUserByUsername(context.Background(), match.OrganizerUsername)
		user, _ := r.service.GetUserByUsername(context.Background(), callback.From.UserName)
		r.service.SetMatchConfirmed(context.Background(), true, user.ID, int64(matchID))
		//TODO: notify organizer
		msg := tgbotapi.NewMessage(callback.From.ID, "Вы подтвердили участие в матче")
		msg.ReplyMarkup = matchMoreKeyboard(int64(matchID))
		r.bot.Send(msg)
		msg = tgbotapi.NewMessage(int64(organizer.ChatID), fmt.Sprintf("@%s подтвердил участие в матче %d", user.Username, matchID))
		msg.ReplyMarkup = matchMoreKeyboard(match.ID)
		r.bot.Send(msg)
	case "signout_match":
		matchID, _ := strconv.Atoi(callbacks[1])
		match, _ := r.service.GetMatchByMatchID(context.TODO(), int64(matchID))
		organizer, _ := r.service.GetUserByUsername(context.Background(), match.OrganizerUsername)
		user, _ := r.service.GetUserByUsername(context.Background(), callback.From.UserName)
		r.service.SignOutMatch(context.Background(), int64(matchID), user.ID)
		msg := tgbotapi.NewMessage(callback.From.ID, "Вы отменили участие в матче")
		msg.ReplyMarkup = matchMoreKeyboard(int64(matchID))
		r.bot.Send(msg)
		msg = tgbotapi.NewMessage(int64(organizer.ChatID), fmt.Sprintf("@%s отменил участие в матче %d", user.Username, matchID))
		msg.ReplyMarkup = matchMoreKeyboard(match.ID)
		r.bot.Send(msg)
	case "signup_match":
		matchID, _ := strconv.Atoi(callbacks[1])
		match, _ := r.service.GetMatchByMatchID(context.Background(), int64(matchID))
		organizer, _ := r.service.GetUserByUsername(context.Background(), match.OrganizerUsername)
		teamID := 0
		min := match.TeamSize
		for _, team := range match.Teams {
			if len(team.Members) < int(min) {
				teamID = int(team.ID)
				min = int64(len(team.Members))
			}
		}
		user, _ := r.service.GetUserByUsername(context.Background(), callback.From.UserName)
		r.service.SignUpToMatch(context.Background(), user.ID, int64(teamID))
		msg := tgbotapi.NewMessage(callback.From.ID, "Вы записались на матч")
		msg.ReplyMarkup = matchMoreKeyboard(match.ID)
		msg = tgbotapi.NewMessage(int64(organizer.ChatID), fmt.Sprintf("@%s отменил участие в матче %d", user.Username, matchID))

		r.bot.Send(msg)
	}
}

// func matchSignUpKeyboard(matchID int64) tgbotapi.InlineKeyboardMarkup {
// 	return tgbotapi.NewInlineKeyboardMarkup(
// 		tgbotapi.NewInlineKeyboardRow(
// 			tgbotapi.NewInlineKeyboardButtonData(
// 				"Записаться на матч",
// 				fmt.Sprintf("signup_match-%d", matchID)),
// 		),
// 	)
// }

func matchMoreKeyboard(matchID int64) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Подробнее о матче", fmt.Sprintf("get_match_by_id-%d", matchID)),
		),
	)
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
		return
	}
	matchID, err := r.service.GetMatchIDByTeamID(context.Background(), user.TeamID)
	if err != nil {
		log.Println(err)
		return
	}
	match, err := r.service.GetMatchByMatchID(context.Background(), matchID)
	if err != nil {
		log.Println(err)
		return
	}
	users := r.service.GetUsersByUsernames(context.Background(), members)
	for _, user := range users {
		r.bot.Send(tgbotapi.NewMessage(int64(user.ChatID), "Вас приглашают на матч"))
		msgToSend := tgbotapi.NewMessage(int64(user.ChatID), fmt.Sprint(match))
		msgToSend.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Отменить участие", fmt.Sprintf("signout_match-%d", matchID)),
				tgbotapi.NewInlineKeyboardButtonData("Подтвердить участие", fmt.Sprintf("confirm_match-%d", matchID)),
			))
		r.bot.Send(msgToSend)
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
	case "get_matches":
		msgToSend := tgbotapi.NewMessage(msg.From.ID, "Выберите вид спорта")
		msgToSend.ReplyMarkup = sportTypeCommandKeyboard
		r.bot.Send(msgToSend)
	case "my_matches":
		user, _ := r.service.GetUserByUsername(context.Background(), msg.From.UserName)
		matches, _ := r.service.GetMatchesByUserID(context.Background(), user.ID)
		r.bot.Send(tgbotapi.NewMessage(msg.From.ID, `🔜 Ближайшие ваши матчи
		`))
		if len(matches) == 0 {
			r.bot.Send(tgbotapi.NewMessage(msg.From.ID, `😥 К сожалению, матчей нет`))
			return
		}
		for _, m := range matches {
			msg := tgbotapi.NewMessage(msg.From.ID,
				fmt.Sprintf(`Матч #%d - Начало %d/%d %d:00(%.1f часа) - %d тг/чел - Осталось %d мест`,
					m.ID, m.StartAt.Day(), m.StartAt.Month(), m.StartAt.Hour(), float64(m.FinishAt.Sub(m.StartAt).Minutes())/60.0,
					m.Rent/(m.TeamCount*m.TeamSize), (m.TeamCount*m.TeamSize)-m.MembersCount))
			msg.ReplyMarkup = matchMoreKeyboard(m.ID)
			r.bot.Send(msg)
		}

	case "organized_matches":
		user, _ := r.service.GetUserByUsername(context.Background(), msg.From.UserName)
		matches, _ := r.service.GetMatchesByUserID(context.Background(), user.ID)
		r.bot.Send(tgbotapi.NewMessage(msg.From.ID, `🔜 Матчи организованные вами
		`))
		if len(matches) == 0 {
			r.bot.Send(tgbotapi.NewMessage(msg.From.ID, `😥 К сожалению, матчей нет`))
			return
		}
		for _, m := range matches {
			msg := tgbotapi.NewMessage(msg.From.ID,
				fmt.Sprintf(`Матч #%d - Начало %d/%d %d:00(%.1f часа) - %d тг/чел - Осталось %d мест`,
					m.ID, m.StartAt.Day(), m.StartAt.Month(), m.StartAt.Hour(), float64(m.FinishAt.Sub(m.StartAt).Minutes())/60.0,
					m.Rent/(m.TeamCount*m.TeamSize), (m.TeamCount*m.TeamSize)-m.MembersCount))
			msg.ReplyMarkup = matchMoreKeyboard(m.ID)
			r.bot.Send(msg)
		}
	}
}

var sportTypeCommandKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("football", "get_matches_by_sport-football"),
		tgbotapi.NewInlineKeyboardButtonData("volleyball", "get_matches_by_sport-volleyball"),
		tgbotapi.NewInlineKeyboardButtonData("basketball", "get_matches_by_sport-basketball"),
	),
)

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

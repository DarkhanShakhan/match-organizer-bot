package entity

import (
	"fmt"
	"time"

	"github.com/DarkhanShakhan/telegram-bot-template/internal/domain/enum"
)

type Match struct {
	ID                int64          `db:"id"`
	Type              enum.SportType `db:"sport"`
	OrganizerID       int64          `db:"organizer_id"`
	OrganizerUsername string
	Location          string    `db:"location"`
	Rent              int64     `db:"rent"`
	StartAt           time.Time `db:"start_at"`
	FinishAt          time.Time `db:"finish_at"`
	TeamSize          int64     `db:"team_size"`
	TeamCount         int64     `db:"team_count"`
	IsPrivate         bool      `db:"private"`
	Teams             []*Team
}

type Team struct {
	ID      int64  `db:"id"`
	Name    string `db:"name"`
	Size    int64  `db:"size"`
	Members []*User
}

type User struct {
	ID       int64  `db:"id"`
	Name     string `db:"name"`
	Username string `db:"username"`
	ChatID   int    `db:"chat_id"`
}

func (u *User) String() string {
	return ""
}

func (m *Match) String() string {
	out := fmt.Sprintf(
		`📢 Игра #%d
	🏆 Спорт: %s
	📍 %s
	👤 Организатор: @%s
	💰 С человека по %dтг
	🗓 Дата матча: %d/%d
	🕖 Начало матча: %d:00 (%.1f часа)
	👥 Формат: %dvs%d (%d команды)

	`,
		m.ID, m.Type, m.Location, m.OrganizerUsername, m.Rent/(m.TeamCount*m.TeamSize), m.StartAt.Day(), m.StartAt.Month(), m.StartAt.Hour(), float64(m.FinishAt.Sub(m.StartAt).Minutes())/60.0,
		m.TeamSize, m.TeamSize, m.TeamCount,
	)
	for _, team := range m.Teams {
		out += team.String()
	}
	out += fmt.Sprintf(`
	🏃‍♂️Осталось %d мест 
	`, m.places())
	return out
}

func (m *Match) places() int64 {
	total := m.TeamCount * m.TeamSize
	for _, t := range m.Teams {
		total -= int64(len(t.Members))
	}
	return total
}

func (t *Team) String() string {
	out := fmt.Sprintf(`%s %d / %d :(`, color[t.Name], len(t.Members), t.Size)
	for _, m := range t.Members {
		out += m.String()
	}
	out += `)

	`
	return out
}

var color = map[string]string{"red": "🟥", "blue": "🟦", "green": "🟩", "yellow": "🟨", "purple": "🟪", "black": "⬛️", "brown": "🟫"}

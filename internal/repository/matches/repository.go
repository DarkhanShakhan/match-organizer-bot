package matches

import (
	"context"

	"github.com/DarkhanShakhan/telegram-bot-template/internal/domain/entity"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	CreateMatch(ctx context.Context, match *entity.Match) (*entity.Match, error)
	GetUserByUsername(ctx context.Context, username string) (*entity.User, error)
	AddTeamMembers(ctx context.Context, teamID int64, userIDs []int64) error
	CreateUser(ctx context.Context, user *entity.User) (*entity.User, error)
	GetMatch(ctx context.Context, matchID int64) (*entity.Match, error)
}

type repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

const (
	createMatchStmt = `INSERT INTO matches(sport, organizer_id, location,team_size, team_count, rent, start_at, finish_at, private)
						VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)
						RETURNING id;`
	createUserStmt        = `INSERT INTO users(name, username, chat_id) VALUES($1, $2, $3) RETURNING id;`
	getUserByUsernameStmt = `SELECT id, name, username, chat_id FROM users WHERE username=$1;`
	createTeamStmt        = `INSERT INTO teams(name,size,match_id) VALUES($1, $2, $3);`
	getTeamsByMatchIDStmt = `SELECT id, name, size FROM teams WHERE match_id=$1`
	getMatchByIDStmt      = `SELECT id, sport,organizer_id, location,team_size,team_count,rent,start_at, finish_at, private
								FROM matches WHERE id = $1`
	createTeamMemberStmt = `INSERT INTO team_members(team_id, member_id, confirmed) VALUES($1, $2, $3);`
)

func (r *repository) AddTeamMembers(ctx context.Context, teamID int64, userIDs []int64) error {
	for _, id := range userIDs {
		_, err := r.pool.Exec(ctx, createTeamMemberStmt, teamID, id, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *repository) GetMatch(ctx context.Context, matchID int64) (*entity.Match, error) {
	var m entity.Match
	if err := pgxscan.Get(ctx, r.pool, &m, getMatchByIDStmt, matchID); err != nil {
		return nil, err
	}
	var teams []*entity.Team
	if err := pgxscan.Select(ctx, r.pool, &teams, getTeamsByMatchIDStmt, matchID); err != nil {
		return nil, err
	}
	m.Teams = teams
	return &m, nil
}

func (r *repository) CreateMatch(ctx context.Context, match *entity.Match) (*entity.Match, error) {
	var id int64
	if err := r.pool.QueryRow(ctx, createMatchStmt, match.Type,
		match.OrganizerID, match.Location,
		match.TeamSize, match.TeamCount,
		match.Rent, match.StartAt,
		match.FinishAt, match.IsPrivate).Scan(&id); err != nil {
		return nil, err
	}
	for _, team := range match.Teams {
		r.pool.Exec(ctx, createTeamStmt, team.Name, team.Size, id)
	}
	var teams []*entity.Team
	if err := pgxscan.Select(ctx, r.pool, &teams, getTeamsByMatchIDStmt, id); err != nil {
		return nil, err
	}
	match.Teams = teams
	match.ID = id
	return match, nil

}

func (r *repository) GetUserByUsername(ctx context.Context, username string) (*entity.User, error) {
	var user entity.User
	if err := pgxscan.Get(ctx, r.pool, &user, getUserByUsernameStmt, username); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) CreateUser(ctx context.Context, user *entity.User) (*entity.User, error) {
	var id int64
	if err := r.pool.QueryRow(ctx, createUserStmt, user.Name, user.Username, user.ChatID).Scan(&id); err != nil {
		return nil, err
	}
	user.ID = id
	return user, nil
}

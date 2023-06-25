package matches

import (
	"context"

	"github.com/DarkhanShakhan/telegram-bot-template/internal/domain/entity"
	"github.com/DarkhanShakhan/telegram-bot-template/internal/domain/enum"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	CreateMatch(ctx context.Context, match *entity.Match) (*entity.Match, error)
	GetUserByUsername(ctx context.Context, username string) (*entity.User, error)
	GetUserByID(ctx context.Context, id int64) (*entity.User, error)
	AddTeamMembers(ctx context.Context, teamID int64, userIDs []int64) error
	CreateUser(ctx context.Context, user *entity.User) (*entity.User, error)
	GetMatch(ctx context.Context, matchID int64) (*entity.Match, error)
	GetTeamsByMatchID(ctx context.Context, matchID int64) ([]*entity.Team, error)
	GetTeamMembers(ctx context.Context, teamID int64) ([]*entity.User, error)
	GetMatchIDByTeamID(ctx context.Context, id int64) (int64, error)
	GetOpenMatchesBySport(ctx context.Context, sport enum.SportType) ([]*entity.Match, error)
	SetMatchConfirmed(ctx context.Context, confirmed bool, memberID, matchID int64) error
	SetMatchPaid(ctx context.Context, paid bool, memberID, matchID int64) error
	SignUpToMatch(ctx context.Context, userID, matchID int64) error
	DeleteTeamMember(ctx context.Context, memberID, matchID int64) error
	CancelMatch(ctx context.Context, matchID int64) error
	GetMatchesByUserID(ctx context.Context, userID int64) ([]*entity.Match, error)
	GetMatchesByOrganizerID(ctx context.Context, userID int64) ([]*entity.Match, error)
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
	getMatchByIDStmt      = `SELECT id, sport,organizer_id, location,team_size,team_count,rent,start_at, finish_at
								FROM matches WHERE id = $1 AND cancelled=false;`
	createTeamMemberStmt   = `INSERT INTO team_members(team_id, member_id, confirmed) VALUES($1, $2, $3);`
	getMembersByTeamIDStmt = `SELECT u.id, u.name, u.username, u.chat_id, tm.confirmed, tm.paid, tm.cancelled 
								FROM team_members tm 
								LEFT JOIN users u 
								ON tm.member_id = u.id
								WHERE tm.team_id = $1;`
	getUserByIDStmt           = `SELECT id, name, username, chat_id FROM users WHERE id=$1;`
	getOpenMatchesBySportStmt = `SELECT m.id,m.team_size,m.team_count, m.rent,m.start_at, m.finish_at, count(tm.member_id) as members_count
									FROM matches m
									LEFT JOIN teams t ON m.id = t.match_id
									LEFT JOIN team_members tm ON t.id = tm.team_id
									WHERE private=false AND sport=$1 AND start_at - interval '30 minutes' > NOW() AND m.cancelled = false
									GROUP BY m.id,m.team_size,m.team_count, m.rent,m.start_at, m.finish_at
									ORDER BY m.start_at DESC;
									`
	setMatchConfirmedStmt   = `UPDATE team_members SET confirmed=$1 WHERE member_id=$2 AND team_id=$3;`
	setMatchPaidStmt        = `UPDATE team_members SET paid=$1 WHERE member_id=$2 AND team_id=$3;`
	deleteTeamMemberStmt    = `DELETE FROM team_members WHERE member_id = $2 AND team_id=$1;`
	getMatchIDByTeamIDStmt  = `SELECT match_id as id FROM teams WHERE id=$1;`
	getTeamIDByMatchAndUser = `SELECT t.id AS id
								FROM teams t 
								LEFT JOIN team_members tm 
								ON t.id=tm.team_id
								WHERE t.match_id=$1 AND tm.member_id = $2;
								`
	signUpToMatchStmt      = `INSERT INTO team_members (team_id, member_id, confirmed) VALUES ($2, $1, true);`
	cancelMatchStmt        = `UPDATE matches SET cancelled = true WHERE id=$1;`
	getMatchesByUserIDStmt = `SELECT m.id,m.team_size,m.team_count, m.rent,m.start_at, m.finish_at, count(tm.member_id) as members_count
								FROM matches m
								LEFT JOIN teams t ON m.id = t.match_id
								LEFT JOIN team_members tm ON t.id = tm.team_id
								WHERE tm.member_id=$1 AND start_at - interval '30 minutes' > NOW() AND m.cancelled = false
								GROUP BY m.id,m.team_size,m.team_count, m.rent,m.start_at, m.finish_at
								ORDER BY m.start_at DESC;`
	getMatchesByOrganizerIDStmt = `SELECT m.id,m.team_size,m.team_count, m.rent,m.start_at, m.finish_at, count(tm.member_id) as members_count
								FROM matches m
								LEFT JOIN teams t ON m.id = t.match_id
								LEFT JOIN team_members tm ON t.id = tm.team_id
								WHERE m.organizer_id=$1 AND start_at - interval '30 minutes' > NOW() AND m.cancelled = false
								GROUP BY m.id,m.team_size,m.team_count, m.rent,m.start_at, m.finish_at
								ORDER BY m.start_at DESC;`
)

func (r *repository) GetMatchesByOrganizerID(ctx context.Context, userID int64) ([]*entity.Match, error) {
	var matches []*entity.Match
	err := pgxscan.Select(ctx, r.pool, &matches, getMatchesByOrganizerIDStmt, userID)
	if err != nil {
		return nil, err
	}
	return matches, nil
}

func (r *repository) GetMatchesByUserID(ctx context.Context, userID int64) ([]*entity.Match, error) {
	var matches []*entity.Match
	err := pgxscan.Select(ctx, r.pool, &matches, getMatchesByUserIDStmt, userID)
	if err != nil {
		return nil, err
	}
	return matches, nil
}

func (r *repository) CancelMatch(ctx context.Context, matchID int64) error {
	_, err := r.pool.Exec(context.Background(), cancelMatchStmt, matchID)
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) SignUpToMatch(ctx context.Context, userID, teamID int64) error {
	_, err := r.pool.Exec(ctx, signUpToMatchStmt, userID, teamID)
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) GetMatchIDByTeamID(ctx context.Context, id int64) (int64, error) {
	var matchID ID
	err := pgxscan.Get(ctx, r.pool, &matchID, getMatchIDByTeamIDStmt, id)
	if err != nil {
		return 0, err
	}
	return matchID.ID, nil
}

func (r *repository) GetTeamIDByMatchAndUser(ctx context.Context, matchID, userID int64) (int64, error) {
	var teamID ID
	err := pgxscan.Get(ctx, r.pool, &teamID, getTeamIDByMatchAndUser, matchID, userID)
	if err != nil {
		return 0, err
	}
	return teamID.ID, nil
}

type ID struct {
	ID int64 `db:"id"`
}

func (r *repository) SetMatchConfirmed(ctx context.Context, confirmed bool, memberID, matchID int64) error {
	teamID, err := r.GetTeamIDByMatchAndUser(ctx, matchID, memberID)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, setMatchConfirmedStmt, confirmed, memberID, teamID)
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) SetMatchPaid(ctx context.Context, paid bool, memberID, matchID int64) error {
	teamID, err := r.GetTeamIDByMatchAndUser(ctx, matchID, memberID)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, setMatchPaidStmt, paid, memberID, teamID)
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) DeleteTeamMember(ctx context.Context, matchID, memberID int64) error {
	teamID, err := r.GetTeamIDByMatchAndUser(ctx, matchID, memberID)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, deleteTeamMemberStmt, teamID, memberID)
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) GetOpenMatchesBySport(ctx context.Context, sport enum.SportType) ([]*entity.Match, error) {
	var matches []*entity.Match
	err := pgxscan.Select(ctx, r.pool, &matches, getOpenMatchesBySportStmt, sport)
	if err != nil {
		return nil, err
	}
	return matches, nil
}

func (r *repository) GetTeamMembers(ctx context.Context, teamID int64) ([]*entity.User, error) {
	var users []*entity.User
	err := pgxscan.Select(ctx, r.pool, &users, getMembersByTeamIDStmt, teamID)
	if err != nil {
		return nil, err
	}
	return users, nil

}

func (r *repository) AddTeamMembers(ctx context.Context, teamID int64, userIDs []int64) error {
	for _, id := range userIDs {
		_, err := r.pool.Exec(ctx, createTeamMemberStmt, teamID, id, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *repository) GetUserByID(ctx context.Context, id int64) (*entity.User, error) {
	var user entity.User
	err := pgxscan.Get(ctx, r.pool, &user, getUserByIDStmt, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) GetMatch(ctx context.Context, matchID int64) (*entity.Match, error) {
	var m entity.Match
	if err := pgxscan.Get(ctx, r.pool, &m, getMatchByIDStmt, matchID); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *repository) GetTeamsByMatchID(ctx context.Context, matchID int64) ([]*entity.Team, error) {
	var teams []*entity.Team
	if err := pgxscan.Select(ctx, r.pool, &teams, getTeamsByMatchIDStmt, matchID); err != nil {
		return nil, err
	}
	return teams, nil
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

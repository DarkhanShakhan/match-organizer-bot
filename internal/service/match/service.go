package match

import (
	"context"

	"github.com/DarkhanShakhan/telegram-bot-template/internal/domain/entity"
	"github.com/DarkhanShakhan/telegram-bot-template/internal/domain/enum"
	"github.com/DarkhanShakhan/telegram-bot-template/internal/repository/matches"
	"github.com/samber/lo"
)

type Service interface {
	CreateMatch(ctx context.Context, match *entity.Match) (*entity.Match, error)
	AddTeamMembers(ctx context.Context, teamID int64, members []string) error
	GetUserByUsername(ctx context.Context, username string) (*entity.User, error)
	GetMatchByMatchID(ctx context.Context, id int64) (*entity.Match, error)
	GetMatchIDByTeamID(ctx context.Context, id int64) (int64, error)
	CreateUser(ctx context.Context, user *entity.User) (*entity.User, error)
	GetUsersByUsernames(ctx context.Context, members []string) []*entity.User
	GetOpenMatchesBySport(ctx context.Context, sport enum.SportType) ([]*entity.Match, error)
	SetMatchPaid(ctx context.Context, paid bool, memberID, matchID int64) error
	SetMatchConfirmed(ctx context.Context, confirmed bool, memberID, teamID int64) error
	SignUpToMatch(ctx context.Context, userID, matchID int64) error
	SignOutMatch(ctx context.Context, userID, matchID int64) error
	CancelMatch(ctx context.Context, matchID int64) error
	GetMatchesByUserID(ctx context.Context, userID int64) ([]*entity.Match, error)
	GetMatchesByOrganizerID(ctx context.Context, userID int64) ([]*entity.Match, error)
}

type service struct {
	matchesRepository matches.Repository
}

func New(matchesRepository matches.Repository) Service {
	return &service{
		matchesRepository: matchesRepository,
	}
}

func (s *service) GetMatchesByOrganizerID(ctx context.Context, userID int64) ([]*entity.Match, error) {
	return s.matchesRepository.GetMatchesByOrganizerID(ctx, userID)
}

func (s *service) GetMatchesByUserID(ctx context.Context, userID int64) ([]*entity.Match, error) {
	return s.matchesRepository.GetMatchesByUserID(ctx, userID)
}

func (s *service) CancelMatch(ctx context.Context, matchID int64) error {
	return s.matchesRepository.CancelMatch(ctx, matchID)
}

func (s *service) SignOutMatch(ctx context.Context, userID, matchID int64) error {
	return s.matchesRepository.DeleteTeamMember(ctx, userID, matchID)
}

func (s *service) SignUpToMatch(ctx context.Context, userID, teamID int64) error {
	return s.matchesRepository.SignUpToMatch(ctx, userID, teamID)
}

func (s *service) SetMatchConfirmed(ctx context.Context, confirmed bool, memberID, matchID int64) error {
	return s.matchesRepository.SetMatchConfirmed(ctx, confirmed, memberID, matchID)
}

func (s *service) SetMatchPaid(ctx context.Context, confirmed bool, memberID, matchID int64) error {
	return s.matchesRepository.SetMatchPaid(ctx, confirmed, memberID, matchID)
}

func (s *service) CreateMatch(ctx context.Context, match *entity.Match) (*entity.Match, error) {
	match.Teams = lo.Map(teams[:match.TeamCount], func(item string, _ int) *entity.Team {
		return &entity.Team{
			Name: item,
		}
	})
	return s.matchesRepository.CreateMatch(ctx, match)
}

func (s *service) GetMatchIDByTeamID(ctx context.Context, id int64) (int64, error) {
	return s.matchesRepository.GetMatchIDByTeamID(ctx, id)
}

func (s *service) GetOpenMatchesBySport(ctx context.Context, sport enum.SportType) ([]*entity.Match, error) {
	return s.matchesRepository.GetOpenMatchesBySport(ctx, sport)
}

func (s *service) GetUserByUsername(ctx context.Context, username string) (*entity.User, error) {
	return s.matchesRepository.GetUserByUsername(ctx, username)
}
func (s *service) CreateUser(ctx context.Context, user *entity.User) (*entity.User, error) {
	return s.matchesRepository.CreateUser(ctx, user)
}

func (s *service) GetMatchByMatchID(ctx context.Context, id int64) (*entity.Match, error) {
	match, err := s.matchesRepository.GetMatch(ctx, id)
	if err != nil {
		return nil, err
	}
	user, err := s.matchesRepository.GetUserByID(ctx, match.OrganizerID)
	if err != nil {
		return nil, err
	}
	match.OrganizerUsername = user.Username
	teams, err := s.matchesRepository.GetTeamsByMatchID(ctx, id)
	if err != nil {
		return nil, err
	}
	for _, team := range teams {
		members, err := s.matchesRepository.GetTeamMembers(ctx, team.ID)
		if err != nil {
			return nil, err
		}
		team.Members = members
	}
	match.Teams = teams
	return match, nil
}

func (s *service) AddTeamMembers(ctx context.Context, teamID int64, members []string) error {
	var users []*entity.User
	for _, member := range members {
		user, err := s.matchesRepository.GetUserByUsername(ctx, member[1:])
		if err != nil {
			continue
		}
		users = append(users, user)
	}
	userIDs := lo.Map(users, func(item *entity.User, _ int) int64 {
		return item.ID
	})
	return s.matchesRepository.AddTeamMembers(ctx, teamID, userIDs)

}

func (s *service) GetUsersByUsernames(ctx context.Context, members []string) []*entity.User {
	var users []*entity.User
	for _, member := range members {
		user, err := s.matchesRepository.GetUserByUsername(ctx, member[1:])
		if err != nil {
			continue
		}
		users = append(users, user)
	}
	return users
}

var teams = []string{"red", "blue", "green", "yellow", "purple", "black", "brown"}

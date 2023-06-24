package match

import (
	"context"

	"github.com/DarkhanShakhan/telegram-bot-template/internal/domain/entity"
	"github.com/DarkhanShakhan/telegram-bot-template/internal/repository/matches"
	"github.com/samber/lo"
)

type Service interface {
	CreateMatch(ctx context.Context, match *entity.Match) (*entity.Match, error)
	AddTeamMembers(ctx context.Context, teamID int64, members []string) error
	GetUserByUsername(ctx context.Context, username string) (*entity.User, error)
	GetMatchByMatchID(ctx context.Context, id int64) (*entity.Match, error)
	CreateUser(ctx context.Context, user *entity.User) (*entity.User, error)
	GetUsersByUsernames(ctx context.Context, members []string) []*entity.User
}

type service struct {
	matchesRepository matches.Repository
}

func New(matchesRepository matches.Repository) Service {
	return &service{
		matchesRepository: matchesRepository,
	}
}

func (s *service) CreateMatch(ctx context.Context, match *entity.Match) (*entity.Match, error) {
	match.Teams = lo.Map(teams[:match.TeamCount], func(item string, _ int) *entity.Team {
		return &entity.Team{
			Name: item,
		}
	})
	return s.matchesRepository.CreateMatch(ctx, match)
}

func (s *service) GetUserByUsername(ctx context.Context, username string) (*entity.User, error) {
	return s.matchesRepository.GetUserByUsername(ctx, username)
}
func (s *service) CreateUser(ctx context.Context, user *entity.User) (*entity.User, error) {
	return s.matchesRepository.CreateUser(ctx, user)
}

func (s *service) GetMatchByMatchID(ctx context.Context, id int64) (*entity.Match, error) {
	return s.matchesRepository.GetMatch(ctx, id)
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

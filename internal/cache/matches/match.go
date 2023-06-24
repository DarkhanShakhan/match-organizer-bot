package matches

import (
	"errors"
	"time"

	"github.com/DarkhanShakhan/telegram-bot-template/internal/domain/entity"
	"github.com/DarkhanShakhan/telegram-bot-template/internal/domain/enum"
	"github.com/patrickmn/go-cache"
)

type Cache interface {
	SetMatch(username string)
	SetSportType(username string, sportType enum.SportType) error
	SetLocation(username string, location string) error
	SetDay(username string, day enum.MatchDay) error
	SetTime(username string, time time.Duration) error
	SetDuration(username string, duration time.Duration) error
	SetTeamSize(username string, size int64) error
	SetTeamCount(username string, count int64) error
	SetRent(username string, rent int64) error
	SetPrivate(username string, private bool) error
	DeleteMatch(username string)
	GetStatus(username string) (Status, error)
	GetMatch(username string) (*entity.Match, error)
}

type Status int64

const (
	StatusNew Status = iota + 1
	StatusSportType
	StatusLocation
	StatusMatchDay
	StatusStartTime
	StatusDuration
	StatusTeamSize
	StatusTeamCount
	StatusRent
	StatusPrivate
	StatusOrga
)

type matchesCache struct {
	prefix string
	cache  *cache.Cache
}

func New(prefix string, defaultDuration, cleanupInterval time.Duration) Cache {
	return &matchesCache{
		prefix: prefix,
		cache:  cache.New(defaultDuration, cleanupInterval),
	}
}

type match struct {
	entity.Match
	status Status
}

func (c *matchesCache) SetMatch(username string) {
	c.cache.Set(c.prefix+username, &match{status: StatusNew}, 0)

}
func (c *matchesCache) SetSportType(username string, sportType enum.SportType) error {
	m, err := c.getMatch(username)
	if err != nil {
		return err
	}
	m.Match.Type = sportType
	m.status = StatusSportType
	c.cache.Set(c.prefix+username, m, 0)
	return nil
}
func (c *matchesCache) SetLocation(username string, location string) error {
	m, err := c.getMatch(username)
	if err != nil {
		return err
	}
	m.Match.Location = location
	m.status = StatusLocation
	c.cache.Set(c.prefix+username, m, 0)
	return nil
}
func (c *matchesCache) SetDay(username string, day enum.MatchDay) error {
	m, err := c.getMatch(username)
	if err != nil {
		return err
	}
	now := time.Now()
	switch day {
	case enum.MatchDayToday:
		m.Match.StartAt = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case enum.MatchDayTomorrow:
		m.Match.StartAt = time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	}
	m.status = StatusMatchDay
	return nil
}

func (c *matchesCache) SetTime(username string, startTime time.Duration) error {
	m, err := c.getMatch(username)
	if err != nil {
		return err
	}
	m.Match.StartAt = m.Match.StartAt.Add(startTime)
	m.status = StatusStartTime
	c.cache.Set(c.prefix+username, m, 0)
	return nil
}
func (c *matchesCache) SetDuration(username string, duration time.Duration) error {
	m, err := c.getMatch(username)
	if err != nil {
		return err
	}
	m.Match.FinishAt = m.Match.StartAt.Add(duration)
	m.status = StatusDuration
	c.cache.Set(c.prefix+username, m, 0)
	return nil
}

func (c *matchesCache) SetTeamSize(username string, size int64) error {
	m, err := c.getMatch(username)
	if err != nil {
		return err
	}
	m.Match.TeamSize = size
	m.status = StatusTeamSize
	c.cache.Set(c.prefix+username, m, 0)
	return nil
}
func (c *matchesCache) SetTeamCount(username string, count int64) error {
	m, err := c.getMatch(username)
	if err != nil {
		return err
	}
	m.Match.TeamCount = count
	m.status = StatusTeamCount
	c.cache.Set(c.prefix+username, m, 0)
	return nil
}
func (c *matchesCache) SetRent(username string, rent int64) error {
	m, err := c.getMatch(username)
	if err != nil {
		return err
	}
	m.Match.Rent = rent
	m.status = StatusRent
	c.cache.Set(c.prefix+username, m, 0)
	return nil
}
func (c *matchesCache) SetPrivate(username string, private bool) error {
	m, err := c.getMatch(username)
	if err != nil {
		return err
	}
	m.Match.IsPrivate = private
	m.status = StatusPrivate
	c.cache.Set(c.prefix+username, m, 0)
	return nil
}
func (c *matchesCache) DeleteMatch(username string) {
	c.cache.Delete(c.prefix + username)
}
func (c *matchesCache) GetStatus(username string) (Status, error) {
	v, ok := c.cache.Get(c.prefix + username)
	if !ok {
		return 0, errors.New("cannot get value")
	}
	m, ok := v.(*match)
	if !ok {
		return 0, errors.New("invalid value")
	}
	return m.status, nil
}
func (c *matchesCache) GetMatch(username string) (*entity.Match, error) {
	v, ok := c.cache.Get(c.prefix + username)
	if !ok {
		return nil, errors.New("cannot get value")
	}
	m, ok := v.(*match)
	if !ok {
		return nil, errors.New("invalid value")
	}
	return &m.Match, nil
}

func (c *matchesCache) getMatch(username string) (*match, error) {
	v, ok := c.cache.Get(c.prefix + username)
	if !ok {
		return nil, errors.New("cannot get value")
	}
	m, ok := v.(*match)
	if !ok {
		return nil, errors.New("invalid value")
	}
	return m, nil
}

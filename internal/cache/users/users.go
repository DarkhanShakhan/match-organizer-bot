package users

import (
	"time"

	"github.com/patrickmn/go-cache"
)

type Status int64

const (
	StatusAddTeamMembers Status = iota + 1
)

type User struct {
	Username string
	ChatID   int64
	MatchID  int64
	TeamID   int64
	Status   Status
}

type Cache interface {
	SetTeamID(username string, teamID int64)
	SetStatus(username string, status Status)
	GetStatus(username string) Status
	SetUser(user User)
	GetUser(username string) (*User, bool)
}

type userCache struct {
	prefix string
	cache  *cache.Cache
}

func New(prefix string, defaultDuration, cleanupInterval time.Duration) Cache {
	return &userCache{
		prefix: prefix,
		cache:  cache.New(defaultDuration, cleanupInterval),
	}
}

func (c *userCache) SetStatus(username string, status Status) {
	u, ok := c.GetUser(username)
	if !ok {
		c.cache.Set(c.prefix+username, &User{Status: status}, 0)
		return
	}
	u.Status = status
	c.cache.Set(c.prefix+username, u, 0)
}

func (c *userCache) SetTeamID(username string, teamID int64) {
	u, ok := c.GetUser(username)
	if !ok {
		c.cache.Set(c.prefix+username, &User{TeamID: teamID}, 0)
		return
	}
	u.TeamID = teamID
	c.cache.Set(c.prefix+username, u, 0)
}

func (c *userCache) GetStatus(username string) Status {
	u, ok := c.GetUser(username)
	if !ok {
		return 0
	}
	return u.Status
}

func (c *userCache) SetUser(user User) {
	c.cache.Set(c.prefix+user.Username, &user, 0)
}

func (c *userCache) GetUser(username string) (*User, bool) {
	v, ok := c.cache.Get(c.prefix + username)
	if !ok {
		return nil, false
	}
	u, ok := v.(*User)
	return u, ok
}

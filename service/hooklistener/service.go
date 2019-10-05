package hooklistener

import (
	"github.com/dewey/webhook-receiver/feed"
	"github.com/dewey/webhook-receiver/notification"
	"github.com/go-kit/kit/log"
)

// Service is an interface for a incoming hook listener service
type Service interface {
	ValidToken(uuid string) (bool, error)
}

type service struct {
	l             log.Logger
	fr            feed.Repository
	nr            notification.Repository
	feedURL       string
	cacheFilePath string
	hookToken     string
}

// NewService initializes a new hook listener service
func NewService(l log.Logger, fr feed.Repository, nr notification.Repository, feedURL string, cacheFilePath string, hookToken string) *service {
	return &service{
		l:             l,
		fr:            fr,
		nr:            nr,
		feedURL:       feedURL,
		cacheFilePath: cacheFilePath,
		hookToken:     hookToken,
	}
}

func (s *service) ValidToken(uuid string) (bool, error) {
	// TODO(dewey): Ideally we'd have some small DB where we can store tokens for users and then do a lookup here
	if uuid != "" && uuid == s.hookToken {
		return true, nil
	}
	return false, nil
}

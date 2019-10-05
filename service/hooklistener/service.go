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
}

// NewService initializes a new usermanager service
func NewService(l log.Logger, fr feed.Repository, nr notification.Repository, feedURL string, cacheFilePath string) *service {
	return &service{
		l:             l,
		fr:            fr,
		nr:            nr,
		feedURL:       feedURL,
		cacheFilePath: cacheFilePath,
	}
}

func (s *service) ValidToken(uuid string) (bool, error) {
	if uuid != "" && uuid == "a901cd2a-3ca0-416f-b1cd-51d67f440c18" {
		return true, nil
	}
	return false, nil
}

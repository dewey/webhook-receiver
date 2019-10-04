package fetcher

import (
	"github.com/go-kit/kit/log"
	"github.com/mmcdole/gofeed"

	_ "github.com/lib/pq"
)

// Service is an interface for a RSS Feed fetcher
type Service interface {
	Entries(feedURL string) ([]*gofeed.Item, error)
}

type service struct {
	l log.Logger
}

// NewService initializes a new fetcher service
func NewService(l log.Logger) *service {
	return &service{
		l: l,
	}
}

func (s *service) Entries(feedURL string) ([]*gofeed.Item, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(feedURL)
	if err != nil {
		return nil, err
	}
	return feed.Items, nil
}

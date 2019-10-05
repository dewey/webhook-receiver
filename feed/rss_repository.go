package feed

import (
	"github.com/go-kit/kit/log"
	"github.com/mmcdole/gofeed"
)

type repository struct {
	l log.Logger
}

// NewRepository initializes a new fetcher service
func NewRepository(l log.Logger) *repository {
	return &repository{
		l: l,
	}
}

func (s *repository) Entries(feedURL string) ([]*gofeed.Item, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(feedURL)
	if err != nil {
		return nil, err
	}
	return feed.Items, nil
}

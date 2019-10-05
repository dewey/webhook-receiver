package feed

import (
	"github.com/mmcdole/gofeed"
)

// Repository is an interface for a RSS Feed fetcher
type Repository interface {
	Entries(feedURL string) ([]*gofeed.Item, error)
}

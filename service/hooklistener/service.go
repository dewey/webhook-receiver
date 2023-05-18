package hooklistener

import (
	"github.com/dewey/webhook-receiver/cache"
	"github.com/mmcdole/gofeed"
	"regexp"

	"github.com/dewey/webhook-receiver/feed"
	"github.com/dewey/webhook-receiver/notification"
	"github.com/go-kit/log"
)

var (
	reSplitCacheKey  = regexp.MustCompile(`(\d{4}-\d{2}-\d{2})(?::(twitter|mastodon|mock\d+))?:(\w{16})`)
	reLegacyCacheKey = regexp.MustCompile(`[a-z0-9]{16}`)
)

// Service is an interface for a incoming hook listener service
type Service interface {
	ValidToken(uuid string) (bool, error)
}

type service struct {
	l         log.Logger
	fr        feed.Repository
	nr        []notification.Repository
	cr        cache.Repository
	feedURL   string
	hookToken string
}

// NewService initializes a new hook listener service
func NewService(l log.Logger, fr feed.Repository, nr []notification.Repository, cr cache.Repository, feedURL string, hookToken string) *service {
	return &service{
		l:         l,
		fr:        fr,
		nr:        nr,
		cr:        cr,
		feedURL:   feedURL,
		hookToken: hookToken,
	}
}

// ValidToken checks if the given token is a valid token. Only we can trigger logic via the received webhook.
func (s *service) ValidToken(uuid string) (bool, error) {
	if uuid != "" && uuid == s.hookToken {
		return true, nil
	}
	return false, nil
}

// getUncachedFeedItem returns a feed item if there's something new and uncached
func (s *service) getNextUncachedFeedItem(items []*gofeed.Item, notificationService string) (*gofeed.Item, bool, error) {
	// For each iteration we only send one notification even if there are more cache misses (aka. unsent tweets). This acts
	// as a natural rate limit and jittering, and they are more spread out.
	for _, item := range items {
		_, exists, err := s.cr.Get(item.GUID, notificationService)
		if err != nil {
			return nil, false, err
		}
		// Item doesn't exist in cache yet, it still needs to be posted
		if !exists {
			return item, false, nil
		}
	}
	return nil, true, nil
}

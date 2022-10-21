package hooklistener

import (
	"errors"
	"regexp"
	"time"

	"github.com/mmcdole/gofeed"

	"github.com/dewey/webhook-receiver/feed"
	"github.com/dewey/webhook-receiver/notification"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

var (
	reSplitCacheKey = regexp.MustCompile(`(\d{4}-\d{2}-\d{2}):(.+)`)
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

// getCacheKey returns the cache key without the timestamp if it exists
func (s *service) getCacheKey(line string) (time.Time, string, error) {
	tokens := reSplitCacheKey.FindStringSubmatch(line)
	if len(tokens) == 3 {
		t, err := time.Parse("2006-01-02", tokens[1])
		if err != nil {
			level.Error(s.l).Log("err", err)
			return time.Time{}, "", err
		}
		return t, tokens[2], nil
	}
	if len(tokens) == 0 {
		return time.Time{}, line, nil
	}
	return time.Time{}, "", errors.New("couldn't get cache key from cache line")
}

// isCached checks if feed items are already cached
func (s *service) isCached(items []*gofeed.Item, cache map[string]time.Time) bool {
	for _, item := range items {
		if _, ok := cache[item.GUID]; !ok {
			return false
		}
	}
	return true
}

// getUncachedFeedItem returns a feed item if there's something new and uncached
func (s *service) getNextUncachedFeedItem(items []*gofeed.Item, cache map[string]time.Time) (*gofeed.Item, bool, error) {
	// For each iteration we only send one notification even if there are more cache misses (aka. unsent tweets). This acts
	// as a natural rate limit and jittering and they are more spread out.
	for _, item := range items {
		if _, ok := cache[item.GUID]; !ok {
			// Item doesn't exist in cache yet, it still needs to be posted
			return item, false, nil
		}
	}
	return nil, true, nil
}

// hasTweetedToday checks if something was posted today, if there's already a Tweet it returns true
func (s *service) hasTweetedToday(m map[string]time.Time) bool {
	for _, val := range m {
		if time.Since(val).Hours() < 24 {
			return true
		} else {
			// Cache entries with no timestamp are skipped. This is to be backwards compatible with the old
			// format and old posts are assumed to be posted already.
			continue
		}
	}
	return false
}

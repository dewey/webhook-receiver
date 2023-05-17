package hooklistener

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"

	"github.com/dewey/webhook-receiver/feed"
	"github.com/dewey/webhook-receiver/notification"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
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
	l             log.Logger
	fr            feed.Repository
	nr            []notification.Repository
	feedURL       string
	cacheFilePath string
	hookToken     string
}

// NewService initializes a new hook listener service
func NewService(l log.Logger, fr feed.Repository, nr []notification.Repository, feedURL string, cacheFilePath string, hookToken string) *service {
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

// getCacheKey returns the cache key without the timestamp if it exists from a full cache entry
func (s *service) getCacheKey(cacheEntry string) (time.Time, string, error) {
	if cacheEntry == "" {
		return time.Time{}, "", errors.New("cache entry is empty")
	}
	tokens := reSplitCacheKey.FindStringSubmatch(cacheEntry)
	if len(tokens) == 4 && tokens[2] != "" {
		t, err := time.Parse("2006-01-02", tokens[1])
		if err != nil {
			level.Error(s.l).Log("err", err)
			return time.Time{}, "", err
		}
		return t, fmt.Sprintf("%s:%s", tokens[2], tokens[3]), nil
	}
	// Legacy: "2023-05-17:fb12e19ed8f09522" cache format
	if len(tokens) == 4 && tokens[2] == "" {
		t, err := time.Parse("2006-01-02", tokens[1])
		if err != nil {
			level.Error(s.l).Log("err", err)
			return time.Time{}, "", err
		}
		return t, fmt.Sprintf("twitter:%s", tokens[3]), nil
	}
	// Legacy: "fb12e19ed8f09522", no timestamps and the cache key was the post id
	legacyToken := reLegacyCacheKey.FindStringSubmatch(cacheEntry)

	// When we had the legacy cache entries we only sent tweets. We also check for the occurance of ":" to differentiate
	// it from the new cache format
	if len(legacyToken) == 1 && !strings.Contains(cacheEntry, ":") {
		return time.Time{}, fmt.Sprintf("twitter:%s", cacheEntry), nil
	}

	return time.Time{}, "", errors.New("couldn't get cache key from cache line")
}

// getUncachedFeedItem returns a feed item if there's something new and uncached
func (s *service) getNextUncachedFeedItem(items []*gofeed.Item, notificationService string, cache map[string]time.Time) (*gofeed.Item, bool, error) {
	// For each iteration we only send one notification even if there are more cache misses (aka. unsent tweets). This acts
	// as a natural rate limit and jittering and they are more spread out.
	for _, item := range items {
		if _, ok := cache[fmt.Sprintf("%s:%s", notificationService, item.GUID)]; !ok {
			// Item doesn't exist in cache yet, it still needs to be posted
			return item, false, nil
		}
	}
	return nil, true, nil
}

// hasPostedToday checks if something was posted today, if there's already a post it returns true
func (s *service) hasPostedToday(m map[string]time.Time, notificationService string) bool {
	for key, val := range m {
		if time.Since(val).Hours() < 24 && strings.Contains(key, notificationService) {
			return true
		} else {
			// Cache entries with no timestamp are skipped. This is to be backwards compatible with the old
			// format and old posts are assumed to be posted already.
			continue
		}
	}
	return false
}

package hooklistener

import (
	"errors"
	"regexp"
	"time"

	"github.com/dewey/webhook-receiver/feed"
	"github.com/dewey/webhook-receiver/notification"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
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

// hasTweetedToday checks if something was posted today, if there's already a Tweet it returns true
func (s *service) hasTweetedToday(m map[string]time.Time) bool {
	for _, val := range m {
		if time.Now().Sub(val).Hours() < 24 {
			return true
		} else {
			// Cache entries with no timestamp are skipped. This is to be backwards compatible with the old
			// format and old posts are assumed to be posted already.
			continue
		}
	}
	return false
}

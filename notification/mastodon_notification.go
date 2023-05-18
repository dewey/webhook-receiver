package notification

import (
	"bytes"
	"context"
	"strings"

	"github.com/mattn/go-mastodon"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/pkg/errors"
)

type mastodonRepository struct {
	l log.Logger
	c *mastodon.Client
}

// NewMastodonRepository initializes a new Mastodon notifier repository
func NewMastodonRepository(l log.Logger, c *mastodon.Client) *mastodonRepository {
	return &mastodonRepository{
		l: l,
		c: c,
	}
}

func (s *mastodonRepository) String() string {
	return "mastodon"
}

func (s *mastodonRepository) Post(ctx context.Context, text string, author string, url string) error {
	// We split into words, and add as many words while trying to stay roughly under 100 characters for the post
	var length int
	var summaryTokens []string
	for _, t := range strings.Split(text, " ") {
		runes := bytes.Runes([]byte(t))
		// If length of toot is below 100, we add tokens to the list
		if length <= 100 {
			length += len(runes)
			summaryTokens = append(summaryTokens, t)
		} else {
			// Only if the word is shorter than 5 do we also still add it to the list
			if len(runes) < 5 {
				summaryTokens = append(summaryTokens, t)
				length += len(runes)
				break
			} else {
				break
			}
		}
	}
	status, err := s.c.PostStatus(ctx, &mastodon.Toot{
		Status: "”" + strings.Join(summaryTokens, " ") + "...“" + "\n\n" + url,
		//InReplyToID: "",
		//MediaIDs:    nil,
		//Sensitive:   false,
		//SpoilerText: "",
		//Visibility:  "",
		//ScheduledAt: nil,
		//Poll:        nil,
	})
	if err != nil {
		return errors.Wrap(err, "posting status update")
	}

	level.Info(s.l).Log("msg", "toot successfully sent", "id", status.ID, "url", status.URL)
	return nil
}

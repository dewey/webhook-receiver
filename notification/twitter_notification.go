package notification

import (
	"bytes"
	"fmt"
	"html"
	"net/http"
	"strings"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/pkg/errors"
)

type repository struct {
	l  log.Logger
	c  *twitter.Client
	tu *twitter.User
}

// NewRepository initializes a new Twitter notifier repository
func NewRepository(l log.Logger, c *twitter.Client, tu *twitter.User) *repository {
	return &repository{
		l:  l,
		c:  c,
		tu: tu,
	}
}

func (s *repository) Post(text string, author string, url string) error {
	// We split into words, and add as many words so we stay routhly under 100 characters for the Tweet
	var length int
	var summaryTokens []string
	for _, t := range strings.Split(text, " ") {
		runes := bytes.Runes([]byte(t))
		// If length of tweet is below 100, we add tokens to the list
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
	tweetBody := html.UnescapeString("”" + strings.Join(summaryTokens, " ") + "...“" + "\n\n" + url)
	t, resp, err := s.c.Statuses.Update(tweetBody, &twitter.StatusUpdateParams{
		TweetMode: "extended",
	})
	if err != nil {
		return errors.Wrap(err, "posting status update")
	}
	if resp.StatusCode != http.StatusOK {
		level.Error(s.l).Log("err", "unexpected status code from twitter", "status_code", resp.StatusCode)
		return errors.New("unexpected status code from twitter")
	}
	level.Info(s.l).Log("msg", "tweet successfully sent", "id", t.IDStr, "url", fmt.Sprintf("https://twitter.com/%s/status/%s", s.tu.ScreenName, t.IDStr))
	return nil
}

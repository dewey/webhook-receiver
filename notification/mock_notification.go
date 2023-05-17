package notification

import (
	"bytes"
	"context"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

type mockRepository struct {
	l    log.Logger
	name string
}

// NewMockRepository initializes a new mock notifier repository to test notifications locally
func NewMockRepository(l log.Logger, serviceName string) *mockRepository {
	return &mockRepository{
		l:    l,
		name: serviceName,
	}
}

func (s *mockRepository) String() string {
	return s.name
}

func (s *mockRepository) Post(ctx context.Context, text string, author string, url string) error {
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
	level.Info(s.l).Log("msg", "mocked notification successfully sent", "id", 123, "url", "https://example.com/123")
	return nil
}

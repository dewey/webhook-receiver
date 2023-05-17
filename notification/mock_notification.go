package notification

import (
	"context"
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
	level.Info(s.l).Log("msg", "mocked notification successfully sent", "notification_service", s.String(), "url", "https://example.com/123")
	return nil
}

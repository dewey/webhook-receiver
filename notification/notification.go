package notification

import "context"

// Repository is an interface for a notifier repository
type Repository interface {
	Post(ctx context.Context, text string, author string, url string) error
}

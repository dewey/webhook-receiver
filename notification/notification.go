package notification

import (
	"context"
	"strings"
)

// Repository is an interface for a notifier repository
type Repository interface {
	Post(ctx context.Context, text string, author string, url string) error
	String() string
}

type Notifiers []Repository

func (n Notifiers) String() string {
	var repositories []string
	for _, repository := range n {
		repositories = append(repositories, repository.String())
	}
	return strings.Join(repositories, ", ")
}

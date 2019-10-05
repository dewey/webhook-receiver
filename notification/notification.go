package notification

// Repository is an interface for a notifier repository
type Repository interface {
	Post(text string, author string, url string) error
}

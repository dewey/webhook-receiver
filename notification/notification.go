package notification

// Repository is an interface for a notifier repository
type Repository interface {
	Post(gofeed.Item item) error
}

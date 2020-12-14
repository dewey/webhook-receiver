package notification

// Repository is an interface for a notifier repository
type Repository interface {
	Post(item gofeed.Item) error
}

package cache

import "time"

// Repository is an interface for the cache
type Repository interface {
	Get(key string, notificationService string) (*Entry, bool, error)
	Set(entry Entry) error
	EntryExists(date time.Time, notificationService string) (bool, error)
}

// Entry is a struct for a cache entry
type Entry struct {
	Key                 string `db:"key"`
	NotificationService string `db:"notification_service"`
	Date                string `db:"date"`
}

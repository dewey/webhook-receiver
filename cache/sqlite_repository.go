package cache

import (
	"github.com/go-kit/log"
	"github.com/jmoiron/sqlx"
	"time"
)

type repository struct {
	l  log.Logger
	db *sqlx.DB
}

// NewRepository initializes a new cache repository
func NewRepository(l log.Logger, db *sqlx.DB) (*repository, error) {
	return &repository{
		l:  l,
		db: db,
	}, nil
}

// Get returns a cache entry for a given key
func (s *repository) Get(key string, notificationService string) (*Entry, bool, error) {
	var entry Entry
	err := s.db.Get(&entry, "SELECT * FROM cache WHERE key=$1 AND notification_service=$2", key, notificationService)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, false, nil
		}
		return nil, false, err
	}

	return &entry, true, nil
}

// Set sets a cache entry
func (s *repository) Set(entry Entry) error {
	_, err := s.db.NamedExec("INSERT INTO cache (key, notification_service, date) VALUES (:key, :notification_service, :date)",
		map[string]interface{}{
			"key":                  entry.Key,
			"notification_service": entry.NotificationService,
			"date":                 entry.Date,
		})
	return err
}

// EntryExists checks if a cache entry exists for a specific day and notification service
func (s *repository) EntryExists(date time.Time, notificationService string) (bool, error) {
	var count int
	if err := s.db.Get(&count, "SELECT COUNT(*) FROM cache WHERE date=$1 AND notification_service=$2", date.Format("2006-01-02"), notificationService); err != nil {
		return false, err
	}
	return count > 0, nil
}

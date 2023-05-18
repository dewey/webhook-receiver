package hooklistener

import (
	"github.com/dewey/webhook-receiver/cache"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-kit/log/level"
	"github.com/pkg/errors"
)

// NewHandler initializes a new archiver API handler
func NewHandler(s service) *chi.Mux {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Post("/{uuid}", webHookHandler(s))
	})

	return r
}

func webHookHandler(s service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Checking if UUID is in our whitelist
		valid, err := s.ValidToken(chi.URLParam(r, "uuid"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			level.Error(s.l).Log("err", err)
			return
		}
		// TODO(dewey): This should not be in the handler, but for now it's good enough
		if valid {
			level.Info(s.l).Log("msg", "received valid token on webhook endpoint", "uuid", chi.URLParam(r, "uuid"))
			items, err := s.fr.Entries(s.feedURL)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				level.Error(s.l).Log("err", errors.Wrap(err, "parsing feed"))
				return
			}

			t := time.Now()
			for _, notificationService := range s.nr {
				// If there's already a post for today in the cache for this service, we do nothing.
				exists, err := s.cr.EntryExists(t, notificationService.String())
				if err != nil {
					level.Error(s.l).Log("err", err)
					continue
				}
				if exists {
					level.Debug(s.l).Log("msg", "there's already a post today, skipping", "notification_service", notificationService.String())
					continue
				}

				item, isCached, err := s.getNextUncachedFeedItem(items, notificationService.String())
				if err != nil {
					level.Error(s.l).Log("err", err)
					continue
				}
				if !isCached {
					level.Info(s.l).Log("msg", "cache miss, notify", "guid", item.GUID, "notification_service", notificationService.String())
					if err := s.cr.Set(cache.Entry{
						Key:                 item.GUID,
						NotificationService: notificationService.String(),
						Date:                time.Now().Format("2006-01-02"),
					}); err != nil {
						level.Error(s.l).Log("err", err)
						continue
					}
					// If item not in cache yet for this notificatino provider, we can send a notification and add it to the cache
					if err := notificationService.Post(r.Context(), item.Description, item.Author.Name, item.Link); err != nil {
						level.Error(s.l).Log("err", err)
						continue
					}
				}
			}

			w.WriteHeader(http.StatusAccepted)
			return
		}
	}
}

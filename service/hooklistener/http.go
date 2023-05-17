package hooklistener

import (
	"bufio"
	"net/http"
	"os"
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

			f, err := os.OpenFile(s.cacheFilePath, os.O_CREATE|os.O_RDWR, 0644)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				level.Error(s.l).Log("err", err)
				return
			}
			defer f.Close()

			// Read cache file into map
			m := make(map[string]time.Time)

			// Add all unique entries of cache file (they should be unique anyway) to map
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				if scanner.Text() == "" {
					continue
				}
				// We assume all past cache entries were already posted on both services
				for _, notificationService := range []string{"twitter", "mastodon"} {
					cacheTimeStamp, key, err := s.getCacheKey(notificationService, scanner.Text())
					if err != nil {
						level.Error(s.l).Log("err", err)
						continue
					}

					// If URL doesn't exist in cache, set cache entry to cacheKey:time.Time in map
					if _, ok := m[key]; !ok {
						m[key] = cacheTimeStamp
					}
				}

			}

			if err := scanner.Err(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				level.Error(s.l).Log("err", err)
				return
			}

			t := time.Now()
			for _, notificationService := range s.nr {
				// If there's a already a post for today in the cache for this service, we do nothing
				if s.hasPostedToday(m, notificationService.String()) {
					level.Debug(s.l).Log("msg", "there's already a post today, skipping", "notification_service", notificationService.String())
					continue
				}

				item, isCached, err := s.getNextUncachedFeedItem(items, notificationService.String(), m)
				if err != nil {
					level.Error(s.l).Log("err", err)
					continue
				}
				if !isCached {
					level.Info(s.l).Log("msg", "cache miss, notify", "guid", item.GUID, "notification_service", notificationService.String())
					_, err := f.WriteString(t.Format("2006-01-02") + ":" + notificationService.String() + ":" + item.GUID + "\n")
					if err != nil {
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

package hooklistener

import (
	"bufio"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-kit/kit/log/level"
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

			// Read cache file into map
			m := make(map[string]time.Time)
			defer f.Close()

			// Add all unique entries of cache file (they should be unique anyway) to map
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				cacheTimeStamp, key, err := s.getCacheKey(scanner.Text())
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					level.Error(s.l).Log("err", err)
					return
				}

				// If URL doesn't exist in cache, set cache entry to url:time.Time in map
				if _, ok := m[key]; !ok {
					m[key] = cacheTimeStamp
				}
			}

			if err := scanner.Err(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				level.Error(s.l).Log("err", err)
				return
			}

			// If there's a already a tweet for today in the cache, we do nothing
			if s.hasTweetedToday(m) {
				w.WriteHeader(http.StatusAccepted)
				level.Debug(s.l).Log("msg", "there's already a tweet today, skipping")
				return
			}

			t := time.Now()
			item, isCached, err := s.getNextUncachedFeedItem(items, m)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				level.Error(s.l).Log("err", err)
				return
			}
			// If item not in cache yet, we can send a notification and add it to the cache
			if !isCached {
				level.Info(s.l).Log("msg", "cache miss, notify", "guid", item.GUID)
				_, err := f.WriteString(t.Format("2006-01-02") + ":" + item.GUID + "\n")
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					level.Error(s.l).Log("err", err)
					return
				}
				if err := s.nr.Post(item.Description, item.Author.Name, item.Link); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					level.Error(s.l).Log("err", err)
					return
				}
			}
			w.WriteHeader(http.StatusAccepted)
			return
		}
	}
}

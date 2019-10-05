package hooklistener

import (
	"bufio"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
)

// NewHandler initializes a new archiver API handler
func NewHandler(s service) *chi.Mux {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Post("/netlify/{uuid}", netlifyHookHandler(s))
	})

	return r
}

func netlifyHookHandler(s service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var np netlifyPayload
		if err := json.NewDecoder(r.Body).Decode(&np); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			level.Error(s.l).Log("err", err)
			return
		}
		// Checking if UUID is in our whitelist
		valid, err := s.ValidToken(chi.URLParam(r, "uuid"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			level.Error(s.l).Log("err", err)
			return
		}
		// TODO(dewey): This should not be in the handler, but for now it's good enough
		if valid {
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
			m := make(map[string]struct{})
			defer f.Close()

			// Add all unique entries of cache file (they should be unique anyway) to map
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				if _, ok := m[scanner.Text()]; !ok {
					m[scanner.Text()] = struct{}{}
				}
			}

			if err := scanner.Err(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				level.Error(s.l).Log("err", err)
				return
			}

			// If item not in cache yet, we can send a notification
			for _, item := range items {
				if _, ok := m[item.GUID]; !ok {
					level.Info(s.l).Log("msg", "cache miss, notify", "guid", item.GUID)
					_, err := f.WriteString(item.GUID + "\n")
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
					// For each iteration we only send one notification even if there are more cache misses (aka. unsent tweets). This acts
					// as a natural rate limit and jittering and they are more spread out.
					break
				}
			}
			w.WriteHeader(http.StatusAccepted)
			level.Info(s.l).Log("msg", "received valid token on webhook endpoint", "uuid", chi.URLParam(r, "uuid"))
			return
		}
	}
}

type netlifyPayload struct {
	ID                  string        `json:"id"`
	SiteID              string        `json:"site_id"`
	BuildID             string        `json:"build_id"`
	State               string        `json:"state"`
	Name                string        `json:"name"`
	URL                 string        `json:"url"`
	SslURL              string        `json:"ssl_url"`
	AdminURL            string        `json:"admin_url"`
	DeployURL           string        `json:"deploy_url"`
	DeploySslURL        string        `json:"deploy_ssl_url"`
	CreatedAt           time.Time     `json:"created_at"`
	UpdatedAt           time.Time     `json:"updated_at"`
	UserID              string        `json:"user_id"`
	ErrorMessage        interface{}   `json:"error_message"`
	Required            []interface{} `json:"required"`
	RequiredFunctions   []interface{} `json:"required_functions"`
	CommitRef           interface{}   `json:"commit_ref"`
	ReviewID            interface{}   `json:"review_id"`
	Branch              string        `json:"branch"`
	CommitURL           interface{}   `json:"commit_url"`
	Skipped             interface{}   `json:"skipped"`
	Locked              interface{}   `json:"locked"`
	LogAccessAttributes struct {
		Type     string `json:"type"`
		URL      string `json:"url"`
		Endpoint string `json:"endpoint"`
		Path     string `json:"path"`
		Token    string `json:"token"`
	} `json:"log_access_attributes"`
	Title              interface{}   `json:"title"`
	ReviewURL          interface{}   `json:"review_url"`
	PublishedAt        time.Time     `json:"published_at"`
	Context            string        `json:"context"`
	DeployTime         int           `json:"deploy_time"`
	AvailableFunctions []interface{} `json:"available_functions"`
	Summary            struct {
		Status   string `json:"status"`
		Messages []struct {
			Type        string      `json:"type"`
			Title       string      `json:"title"`
			Description string      `json:"description"`
			Details     interface{} `json:"details"`
		} `json:"messages"`
	} `json:"summary"`
	ScreenshotURL    interface{} `json:"screenshot_url"`
	SiteCapabilities struct {
		Title             string `json:"title"`
		AssetAcceleration bool   `json:"asset_acceleration"`
		FormProcessing    bool   `json:"form_processing"`
		CdnPropagation    string `json:"cdn_propagation"`
		BuildGcExchange   string `json:"build_gc_exchange"`
		BuildNodePool     string `json:"build_node_pool"`
		DomainAliases     bool   `json:"domain_aliases"`
		SecureSite        bool   `json:"secure_site"`
		Prerendering      bool   `json:"prerendering"`
		Proxying          bool   `json:"proxying"`
		Ssl               string `json:"ssl"`
		RateCents         int    `json:"rate_cents"`
		YearlyRateCents   int    `json:"yearly_rate_cents"`
		CdnNetwork        string `json:"cdn_network"`
		Ipv6Domain        string `json:"ipv6_domain"`
		BranchDeploy      bool   `json:"branch_deploy"`
		ManagedDNS        bool   `json:"managed_dns"`
		GeoIP             bool   `json:"geo_ip"`
		SplitTesting      bool   `json:"split_testing"`
		ID                string `json:"id"`
	} `json:"site_capabilities"`
	Committer  interface{} `json:"committer"`
	SkippedLog interface{} `json:"skipped_log"`
}
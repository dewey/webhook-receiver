package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/dewey/webhook-receiver/feed"
	"github.com/dewey/webhook-receiver/notification"
	"github.com/dewey/webhook-receiver/service/hooklistener"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/go-chi/chi/v5"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/mattn/go-mastodon"
	"github.com/peterbourgon/ff/v3"
)

type maxBytesHandler struct {
	h http.Handler
	n int64
}

func (h *maxBytesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.n)
	h.h.ServeHTTP(w, r)
}

func main() {
	fs := flag.NewFlagSet("webhook-receiver", flag.ExitOnError)
	var (
		environment              = fs.String("environment", "develop", "the environment we are running in")
		port                     = fs.String("port", "8080", "the port webhook-receiver is running on")
		twitterConsumerKey       = fs.String("twitter-consumer-key", "", "the twitter consumer key")
		twitterConsumerSecretKey = fs.String("twitter-consumer-secret-key", "", "the twitter consumer secret key")
		twitterAccessToken       = fs.String("twitter-access-token", "", "the twitter consumer key")
		twitterAccessTokenSecret = fs.String("twitter-access-token-secret", "", "the twitter consumer secret key")
		twitterUsername          = fs.String("twitter-username", "", "the twitter username you are connecting to")
		mastodonClientKey        = fs.String("mastodon-client-key", "", "the mastodon client key")
		mastodonClientSecret     = fs.String("mastodon-client-secret", "", "the mastodon client secret")
		mastodonAccessToken      = fs.String("mastodon-access-token", "", "the mastodon access token")
		mastodonServer           = fs.String("mastodon-server", "", "the mastodon instance you are using")
		feedURL                  = fs.String("feed-url", "https://annoying.technology/index.xml", "the direct url to the feed index")
		cacheFilePath            = fs.String("cache-file-path", "cache", "the path to the cache file, to prevent duplicate notifications")
		hookToken                = fs.String("hook-token", "changeme", "the secret token for the hook, to prevent other people from hitting the hook")
	)

	ff.Parse(fs, os.Args[1:],
		ff.WithConfigFileFlag("config"),
		ff.WithConfigFileParser(ff.PlainParser),
		ff.WithEnvVarPrefix("WR"),
	)

	// Heroku doesn't support EnvVarPrefixes so we have to overwrite this
	if os.Getenv("PORT") != "" {
		*port = os.Getenv("PORT")
	}

	l := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	switch strings.ToLower(*environment) {
	case "development":
		l = level.NewFilter(l, level.AllowInfo())
	case "prod":
		l = level.NewFilter(l, level.AllowError())
	}
	l = log.With(l, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	var notifiers notification.Notifiers
	// Set up Twitter client
	if *twitterConsumerKey != "" && *twitterAccessTokenSecret != "" && *twitterConsumerSecretKey != "" && *twitterAccessToken != "" {
		config := oauth1.NewConfig(*twitterConsumerKey, *twitterConsumerSecretKey)
		token := oauth1.NewToken(*twitterAccessToken, *twitterAccessTokenSecret)
		httpClient := config.Client(oauth1.NoContext, token)
		client := twitter.NewClient(httpClient)

		// Get user information for setup testing
		user, resp, err := client.Users.Show(&twitter.UserShowParams{
			ScreenName: *twitterUsername,
		})
		if err != nil {
			level.Error(l).Log("err", "error getting user information from twitter")
			return
		}
		if resp.StatusCode != http.StatusOK {
			level.Error(l).Log("err", "status code not 200, check credentials and api.twitterstat.us")
			return
		}
		level.Info(l).Log("msg", "connected to twitter", "twitter_user_id", user.IDStr, "twitter_user", user.ScreenName, "http_status", resp.StatusCode)
		notifiers = append(notifiers, notification.NewTwitterRepository(l, client, user))
	}
	// Setup Mastodon Client
	if *mastodonServer != "" && *mastodonClientKey != "" && *mastodonClientSecret != "" && *mastodonAccessToken != "" {
		cm := mastodon.NewClient(&mastodon.Config{
			Server:       *mastodonServer,
			ClientID:     *mastodonClientKey,
			ClientSecret: *mastodonClientSecret,
			AccessToken:  *mastodonAccessToken,
		})
		clientMastodon, err := cm.GetAccountCurrentUser(context.Background())
		if err != nil {
			level.Error(l).Log("err", "error getting user information from mastodon")
			return
		}
		level.Info(l).Log("msg", "connected to mastodon", "mastodon_user_id", clientMastodon.ID, "mastodon_user", clientMastodon.Username)
		notifiers = append(notifiers, notification.NewMastodonRepository(l, cm))
	}

	// For local development we inject a mock notifier which just prints out a notification. That way we can test the caching
	// logic without setting up real services. The name has to follow the naming convention "mock[\d+]" as defined in service.go
	if *environment == "develop" {
		notifiers = append(notifiers, notification.NewMockRepository(l, "mock1"))
		notifiers = append(notifiers, notification.NewMockRepository(l, "mock2"))
		notifiers = append(notifiers, notification.NewMockRepository(l, "twitter"))
		notifiers = append(notifiers, notification.NewMockRepository(l, "mastodon"))
	}

	if len(notifiers) == 0 {
		level.Error(l).Log("err", "no notifiers are configured. make sure to set up twitter and/or mastodon")
		return
	} else {
		level.Info(l).Log("msg", "configured notifiers", "notifiers", notifiers.String())
	}

	// Set up HTTP API
	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("webhook-receiver"))
	})

	fr := feed.NewRepository(l)
	listenerService := hooklistener.NewService(l, fr, notifiers, *feedURL, *cacheFilePath, *hookToken)

	r.Mount("/incoming-hooks", hooklistener.NewHandler(*listenerService))

	level.Info(l).Log("msg", fmt.Sprintf("webhook-receiver is running on :%s", *port), "environment", *environment)

	// Set up webserver and and set max file limit to 50MB
	err := http.ListenAndServe(fmt.Sprintf(":%s", *port), &maxBytesHandler{h: r, n: (50 * 1024 * 1024)})
	if err != nil {
		level.Error(l).Log("err", err)
		return
	}
}

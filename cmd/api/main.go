package main

import (
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
		port                     = fs.String("port", "8080", "the port archivepipe is running on")
		twitterConsumerKey       = fs.String("twitter-consumer-key", "changeme", "the twitter consumer key")
		twitterConsumerSecretKey = fs.String("twitter-consumer-secret-key", "changeme", "the twitter consumer secret key")
		twitterAccessToken       = fs.String("twitter-access-token", "changeme", "the twitter consumer key")
		twitterAccessTokenSecret = fs.String("twitter-access-token-secret", "changeme", "the twitter consumer secret key")
		twitterUsername          = fs.String("twitter-username", "annoyingfeed", "the twitter username you are connecting to")
		feedURL                  = fs.String("feed-url", "https://annoying.technology/index.xml", "the direct url to the feed index")
		cacheFilePath            = fs.String("cache-file-path", "~/cache", "the path to the cache file, to prevent duplicate notifications")
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

	// Set up Twitter client
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

	// Set up HTTP API
	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("webhook-receiver"))
	})

	fr := feed.NewRepository(l)
	nr := notification.NewRepository(l, client, user)
	listenerService := hooklistener.NewService(l, fr, nr, *feedURL, *cacheFilePath, *hookToken)

	r.Mount("/incoming-hooks", hooklistener.NewHandler(*listenerService))

	level.Info(l).Log("msg", fmt.Sprintf("webhook-receiver is running on :%s", *port), "environment", *environment)

	// Set up webserver and and set max file limit to 50MB
	err = http.ListenAndServe(fmt.Sprintf(":%s", *port), &maxBytesHandler{h: r, n: (50 * 1024 * 1024)})
	if err != nil {
		level.Error(l).Log("err", err)
		return
	}
}

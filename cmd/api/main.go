package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/dewey/webhook-receiver/service/fetcher"
	"github.com/dewey/webhook-receiver/service/listener"
	"github.com/dewey/webhook-receiver/service/notifier"
	"github.com/dghubble/go-twitter/twitter"

	"github.com/go-chi/chi"

	"github.com/dghubble/oauth1"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	_ "github.com/lib/pq"
	"github.com/peterbourgon/ff"
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

	// Twitter client
	client := twitter.NewClient(httpClient)

	// User Show
	user, resp, err := client.Users.Show(&twitter.UserShowParams{
		ScreenName: *twitterUsername,
	})
	if resp.StatusCode != http.StatusOK {
		level.Error(l).Log("err", "status code not 200, check credentials and api.twitterstat.us")
	}
	level.Info(l).Log("msg", "connected to twitter", "twitter_user_id", user.IDStr, "http_status", resp.StatusCode)

	// Set up HTTP API
	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("webhook-receiver"))
	})

	listenerService := listener.NewService(l)
	fetcherService := fetcher.NewService(l)
	notifierService := notifier.NewService(l, client)
	r.Mount("/incoming-hooks", listener.NewHandler(*listenerService, fetcherService, notifierService))

	level.Info(l).Log("msg", fmt.Sprintf("webhook-receiver is running on :%s", *port), "environment", *environment)

	// Set up webserver and and set max file limit to 50MB
	err = http.ListenAndServe(fmt.Sprintf(":%s", *port), &maxBytesHandler{h: r, n: (50 * 1024 * 1024)})
	if err != nil {
		level.Error(l).Log("err", err)
		return
	}
}

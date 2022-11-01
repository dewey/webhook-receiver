package notification

import (
	"context"
	"flag"
	"os"
	"testing"

	"github.com/peterbourgon/ff/v3"

	"github.com/go-kit/log"
	"github.com/mattn/go-mastodon"
)

func Test_mastodonRepository_Post(t *testing.T) {
	// Only for local development to test posting
	t.SkipNow()
	fs := flag.NewFlagSet("webhook-receiver", flag.ExitOnError)
	var (
		mastodonClientKey    = fs.String("mastodon-client-key", "changeme", "the mastodon client key")
		mastodonClientSecret = fs.String("mastodon-client-secret", "changeme", "the mastodon client secret")
		mastodonAccessToken  = fs.String("mastodon-access-token", "changeme", "the mastodon access token")
		mastodonServer       = fs.String("mastodon-server", "changeme", "the mastodon instance you are using")
	)

	ff.Parse(fs, os.Args[1:],
		ff.WithConfigFileFlag("config"),
		ff.WithConfigFileParser(ff.PlainParser),
		ff.WithEnvVarPrefix("WR"),
	)
	type fields struct {
		l log.Logger
		c *mastodon.Client
	}
	type args struct {
		ctx    context.Context
		text   string
		author string
		url    string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "testing tooting",
			fields: fields{
				l: log.NewNopLogger(),
				c: mastodon.NewClient(&mastodon.Config{
					Server:       *mastodonServer,
					ClientID:     *mastodonClientKey,
					ClientSecret: *mastodonClientSecret,
					AccessToken:  *mastodonAccessToken,
				}),
			},
			args: args{
				ctx:    context.TODO(),
				text:   "Testing something",
				author: "Philipp",
				url:    "https://annoying.technology/posts/96c086bc855f1aa8/",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &mastodonRepository{
				l: tt.fields.l,
				c: tt.fields.c,
			}
			if err := s.Post(tt.args.ctx, tt.args.text, tt.args.author, tt.args.url); (err != nil) != tt.wantErr {
				t.Errorf("Post() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

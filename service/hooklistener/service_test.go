package hooklistener

import (
	"github.com/dewey/webhook-receiver/feed"
	"github.com/dewey/webhook-receiver/notification"
	"reflect"
	"testing"
	"time"

	"github.com/go-kit/log"
	"github.com/mmcdole/gofeed"
)

var (
	t1, _ = time.Parse("2006-01-02", "2020-01-02")
	t2, _ = time.Parse("2006-01-02", "2020-01-03")
	t3, _ = time.Parse("2006-01-02", "2020-01-04")
	t4, _ = time.Parse("2006-01-02", "2020-05-22")
	t5, _ = time.Parse("2006-01-02", "2020-05-20")

	cacheMapOne = map[string]time.Time{
		"twitter:https://annoying.technology/posts/1/": time.Time{},
		"twitter:https://annoying.technology/2/":       t1,
		"twitter:https://annoying.technology/posts/3/": t2,
		"twitter:https://annoying.technology/posts/4/": t3,
		"twitter:https://annoying.technology/posts/5/": t4,
	}
	cacheMapTwo = map[string]time.Time{
		"twitter:https://annoying.technology/posts/1/": time.Time{},
		"twitter:https://annoying.technology/2/":       t1,
		"twitter:https://annoying.technology/posts/3/": t2,
		"twitter:https://annoying.technology/posts/4/": t3,
		"twitter:https://annoying.technology/posts/5/": t5,
	}

	timeOut, _ = time.Parse("2006-01-02", "2020-01-04")
)

var tweetedTests = []struct {
	name  string
	cache map[string]time.Time
	out   bool
}{
	{"all tweeted already, don't send out tweet", cacheMapOne, false},
	{"today still has to be tweeted", cacheMapTwo, false},
}

func TestHasPostedToday(t *testing.T) {
	service := NewService(log.NewNopLogger(), nil, nil, "", "", "")
	for _, tt := range tweetedTests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.hasPostedToday(tt.cache, "twitter")
			if got != tt.out {
				t.Errorf("got %t, want %t", got, tt.out)
			}
		})
	}
}

var cacheTests = []struct {
	line    string
	outTime time.Time
	outURL  string
}{
	{"https://annoying.technology/posts/1/", time.Time{}, "https://annoying.technology/posts/1/"},
	{"2020-01-04:https://annoying.technology/posts/4/", timeOut, "https://annoying.technology/posts/4/"},
}

func TestGetCacheKey(t *testing.T) {
	service := NewService(log.NewNopLogger(), nil, nil, "", "", "")
	for _, tt := range cacheTests {
		t.Run("testing cleaning function", func(t *testing.T) {
			outTime, outURL, err := service.getCacheKey(tt.line)
			if err != nil {
				t.Errorf("shouldn't get an error")
			}
			if outTime != tt.outTime {
				t.Errorf("got %s, want %s", outTime, tt.outTime)
			}
			if outURL != tt.outURL {
				t.Errorf("got %s, want %s", outURL, tt.outURL)
			}
		})
	}
}

var isCachedTests = []struct {
	cache               map[string]time.Time
	notificationService string
	items               []*gofeed.Item
	outIsCached         bool
	outNewItem          gofeed.Item
}{
	{cacheMapOne, "twitter", []*gofeed.Item{
		&gofeed.Item{
			GUID: "https://annoying.technology/posts/1/",
		},
		&gofeed.Item{
			GUID: "https://annoying.technology/posts/5/",
		},
	}, true, gofeed.Item{}},
	{cacheMapOne, "twitter", []*gofeed.Item{
		&gofeed.Item{
			GUID: "https://annoying.technology/posts/1/",
		},
		&gofeed.Item{
			GUID: "https://annoying.technology/posts/6/",
		},
	}, false, gofeed.Item{GUID: "https://annoying.technology/posts/6/"}},
	{cacheMapOne, "twitter", []*gofeed.Item{
		&gofeed.Item{
			GUID: "https://annoying.technology/posts/1/",
		},
		&gofeed.Item{
			GUID: "https://annoying.technology/posts/6/",
		},
		&gofeed.Item{
			GUID: "https://annoying.technology/posts/7/",
		},
	}, false, gofeed.Item{GUID: "https://annoying.technology/posts/6/"}},
}

func TestGetNextUncachedFeedItem(t *testing.T) {
	service := NewService(log.NewNopLogger(), nil, nil, "", "", "")
	for _, tt := range isCachedTests {
		t.Run("testing cache fetching function", func(t *testing.T) {
			newItem, isCached, err := service.getNextUncachedFeedItem(tt.items, tt.notificationService, tt.cache)
			if err != nil {
				t.Error("got error that shouldn't be there", err)
			}
			if tt.outIsCached != isCached {
				t.Errorf("got %t, want %t", isCached, tt.outIsCached)
			}
			// If it's not all cached there's a new item
			if !isCached {
				if newItem.GUID != tt.outNewItem.GUID {
					t.Errorf("got %s, want %s", newItem.GUID, tt.outNewItem.GUID)
				}
			}
		})
	}
}

func Test_service_getCacheKey(t *testing.T) {
	t2, err := time.Parse("2006-01-02", "2021-04-29")
	if err != nil {
		t.FailNow()
	}

	type fields struct {
		l             log.Logger
		fr            feed.Repository
		nr            []notification.Repository
		feedURL       string
		cacheFilePath string
		hookToken     string
	}
	type args struct {
		cacheEntry string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    time.Time
		want1   string
		wantErr bool
	}{
		{
			name: "invalid cache format",
			fields: fields{
				l: log.NewNopLogger(),
			},
			args: args{
				cacheEntry: "not a valid cache entry",
			},
			want:    time.Time{},
			want1:   "",
			wantErr: true,
		},
		{
			name: "valid cache format, invalid notification service",
			fields: fields{
				l: log.NewNopLogger(),
			},
			args: args{
				cacheEntry: "example:c1f50e78a65d2ce3",
			},
			want:    time.Time{},
			want1:   "",
			wantErr: true,
		},
		{
			name: "valid legacy cache format",
			fields: fields{
				l: log.NewNopLogger(),
			},
			args: args{
				cacheEntry: "c1f50e78a65d2ce3",
			},
			want:    time.Time{},
			want1:   "twitter:c1f50e78a65d2ce3",
			wantErr: false,
		},
		{
			name: "valid cache format",
			fields: fields{
				l: log.NewNopLogger(),
			},
			args: args{
				cacheEntry: "2021-04-29:mastodon:1594fa5281d264a3",
			},
			want:    t2,
			want1:   "mastodon:1594fa5281d264a3",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				l:             tt.fields.l,
				fr:            tt.fields.fr,
				nr:            tt.fields.nr,
				feedURL:       tt.fields.feedURL,
				cacheFilePath: tt.fields.cacheFilePath,
				hookToken:     tt.fields.hookToken,
			}
			got, got1, err := s.getCacheKey(tt.args.cacheEntry)
			if (err != nil) != tt.wantErr {
				t.Errorf("getCacheKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getCacheKey() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getCacheKey() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

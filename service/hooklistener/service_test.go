package hooklistener

import (
	"testing"
	"time"

	"github.com/go-kit/kit/log"
)

var (
	t1, _ = time.Parse("2006-01-02", "2020-01-02")
	t2, _ = time.Parse("2006-01-02", "2020-01-03")
	t3, _ = time.Parse("2006-01-02", "2020-01-04")
	t4, _ = time.Parse("2006-01-02", "2020-05-22")
	t5, _ = time.Parse("2006-01-02", "2020-05-20")

	cacheMapOne = map[string]time.Time{
		"https://annoying.technology/posts/1/": time.Time{},
		"https://annoying.technology/2/":       t1,
		"https://annoying.technology/posts/3/": t2,
		"https://annoying.technology/posts/4/": t3,
		"https://annoying.technology/posts/5/": t4,
	}
	cacheMapTwo = map[string]time.Time{
		"https://annoying.technology/posts/1/": time.Time{},
		"https://annoying.technology/2/":       t1,
		"https://annoying.technology/posts/3/": t2,
		"https://annoying.technology/posts/4/": t3,
		"https://annoying.technology/posts/5/": t5,
	}

	timeOut, _ = time.Parse("2006-01-02", "2020-01-04")
)

var tweetedTests = []struct {
	name  string
	cache map[string]time.Time
	out   bool
}{
	{"all tweeted already, don't send out tweet", cacheMapOne, true},
	{"today still has to be tweeted", cacheMapTwo, false},
}

func TestHasTweetedToday(t *testing.T) {
	service := NewService(log.NewNopLogger(), nil, nil, "", "", "")
	for _, tt := range tweetedTests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.hasTweetedToday(tt.cache)
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

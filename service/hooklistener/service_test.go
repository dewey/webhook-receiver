package hooklistener

import (
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
	{"all tweeted already, don't send out tweet", cacheMapOne, false},
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

var isCachedTests = []struct {
	cache       map[string]time.Time
	items       []*gofeed.Item
	outIsCached bool
	outNewItem  gofeed.Item
}{
	{cacheMapOne, []*gofeed.Item{
		&gofeed.Item{
			GUID: "https://annoying.technology/posts/1/",
		},
		&gofeed.Item{
			GUID: "https://annoying.technology/posts/5/",
		},
	}, true, gofeed.Item{}},
	{cacheMapOne, []*gofeed.Item{
		&gofeed.Item{
			GUID: "https://annoying.technology/posts/1/",
		},
		&gofeed.Item{
			GUID: "https://annoying.technology/posts/6/",
		},
	}, false, gofeed.Item{GUID: "https://annoying.technology/posts/6/"}},
	{cacheMapOne, []*gofeed.Item{
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

func TestIsCached(t *testing.T) {
	service := NewService(log.NewNopLogger(), nil, nil, "", "", "")
	for _, tt := range isCachedTests {
		t.Run("testing cache checking function", func(t *testing.T) {
			isCached := service.isCached(tt.items, tt.cache)
			if tt.outIsCached != isCached {
				t.Errorf("got %t, want %t", isCached, tt.outIsCached)
			}
		})
	}
}

func TestGetNextUncachedFeedItem(t *testing.T) {
	service := NewService(log.NewNopLogger(), nil, nil, "", "", "")
	for _, tt := range isCachedTests {
		t.Run("testing cache fetching function", func(t *testing.T) {
			newItem, isCached, err := service.getNextUncachedFeedItem(tt.items, tt.cache)
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

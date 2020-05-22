package hooklistener

import (
	"testing"

	"github.com/go-kit/kit/log"
)

var (
	cacheMapOne = map[string]struct{}{
		"https://annoying.technology/posts/1/":            struct{}{},
		"2020-01-02:https://annoying.technology/2/":       struct{}{},
		"2020-01-03:https://annoying.technology/posts/3/": struct{}{},
		"2020-01-04:https://annoying.technology/posts/4/": struct{}{},
		"2020-05-22:https://annoying.technology/posts/5/": struct{}{},
	}
	cacheMapTwo = map[string]struct{}{
		"https://annoying.technology/posts/1/":            struct{}{},
		"2020-01-02:https://annoying.technology/2/":       struct{}{},
		"2020-01-03:https://annoying.technology/posts/3/": struct{}{},
		"2020-01-04:https://annoying.technology/posts/4/": struct{}{},
		"2020-05-20:https://annoying.technology/posts/5/": struct{}{},
	}
)

var flagtests = []struct {
	name  string
	cache map[string]struct{}
	out   bool
}{
	{"all tweeted already, don't send out tweet", cacheMapOne, true},
	{"today still has to be tweeted", cacheMapTwo, false},
}

func TestHasTweetedToday(t *testing.T) {
	service := NewService(log.NewNopLogger(), nil, nil, "", "", "")
	for _, tt := range flagtests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.hasTweetedToday(tt.cache)
			if got != tt.out {
				t.Errorf("got %t, want %t", got, tt.out)
			}
		})
	}
}

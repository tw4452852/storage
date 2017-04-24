package storage

import (
	"io"
	"time"
)

// Init init the dataCenter and repositories
func Init(configPath string) {
	initStorage()
	initRepos(configPath)
}

const (
	TimePattern = "2006-01-02"
)

// Keyer represent a key to post
type Keyer interface {
	Key() string
}

type Staticer interface {
	Static(string) io.ReadCloser
}

// Poster represet a post
type Poster interface {
	Date() time.Time
	Content() string
	Title() string
	Keyer
	Staticer
	Update() error
	Tags() []string
	IsSlide() bool
	StaticList() []string
}

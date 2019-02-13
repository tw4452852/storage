package storage

import (
	"io"
	"time"
)

// Keyer represent a key to post
type Keyer interface {
	// key returns a strong for a key
	Key() string
}

// Staticer represent a handler for static resources (e.g. images)
type Staticer interface {
	// Static receives a path of a source and returns its contents
	Static(string) io.ReadCloser
}

// Poster represet a post item
type Poster interface {
	// Date returns creation time.
	Date() time.Time
	// Content returns main body of a post.
	Content() string
	// Title returns post's headline.
	Title() string
	// Poster is also a keyer which is used to identify this itself.
	Keyer
	// Poster is also a Staticer for handling static resources.
	Staticer
	// Tags get all the tags this post belongs.
	Tags() []string
	// IsSlide reports whether this post is in slide format or not.
	IsSlide() bool
	// StaticList gives a list of all static resources
	StaticList() []string
}

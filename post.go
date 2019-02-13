package storage

import (
	"io"
	"io/ioutil"
	"strings"
	"sync"
	"time"
)

// meta contain the necessary infomations of a post
type meta struct {
	key        string
	title      string
	date       time.Time
	content    string
	tags       []string
	isSlide    bool
	staticList []string
}

// post represent a basic Poster instance
type post struct {
	sync.RWMutex
	meta
}

var _ Poster = &post{}

func newPost(m meta) *post {
	return &post{
		meta: m,
	}
}

// implement Poster's common part
func (p *post) Date() time.Time {
	p.RLock()
	defer p.RUnlock()
	return p.date
}

func (p *post) Content() string {
	p.RLock()
	defer p.RUnlock()
	return p.content
}

func (p *post) Title() string {
	p.RLock()
	defer p.RUnlock()
	return p.title
}

func (p *post) Key() string {
	p.RLock()
	defer p.RUnlock()
	return p.key
}

func (p *post) Tags() []string {
	p.RLock()
	defer p.RUnlock()
	return p.tags
}

func (p *post) IsSlide() bool {
	p.RLock()
	defer p.RUnlock()
	return p.isSlide
}

func (p *post) StaticList() []string {
	p.RLock()
	defer p.RUnlock()
	return p.staticList
}

func (p *post) Static(string) io.ReadCloser {
	return ioutil.NopCloser(strings.NewReader("nop"))
}

func title2Key(title string) string {
	return strings.Replace(title, " ", "_", -1)
}

const imagePrefix = "/images/" //add this prefix to the origin image link
func generateImageLink(key, link string) string {
	return imagePrefix + key + "/" + link
}

// wantChange check whether the image's link need to add prefix
func needChangeImageLink(link string) bool {
	if strings.HasPrefix(link, "http://") ||
		strings.HasPrefix(link, "https://") {
		return false
	}
	return true
}

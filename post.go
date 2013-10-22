package storage

import (
	"html/template"
	"strings"
	"sync"
	"time"
)

// meta contain the necessary infomations of a post
type meta struct {
	key     string
	title   string
	date    time.Time
	content template.HTML
	tags    []string
}

//post represent a poster instance
type post struct { /*{{{*/
	sync.RWMutex
	meta
} /*}}}*/

func newPost() *post { /*{{{*/
	return new(post)
} /*}}}*/

//implement Poster's common part
func (p *post) Date() time.Time { /*{{{*/
	p.RLock()
	defer p.RUnlock()
	return p.date
} /*}}}*/

func (p *post) Content() template.HTML { /*{{{*/
	p.RLock()
	defer p.RUnlock()
	return p.content
} /*}}}*/

func (p *post) Title() template.HTML { /*{{{*/
	p.RLock()
	defer p.RUnlock()
	return template.HTML(p.title)
} /*}}}*/

func (p *post) Key() string { /*{{{*/
	p.RLock()
	defer p.RUnlock()
	return p.key
} /*}}}*/

func (p *post) Tags() []string {
	p.RLock()
	defer p.RUnlock()
	return p.tags
}

func (p *post) update(m *meta) { /*{{{*/
	//update meta
	p.Lock()
	p.meta = *m
	p.Unlock()
	m = nil
} /*}}}*/

func title2Key(title string) string {
	return strings.Replace(title, " ", "_", -1)
}

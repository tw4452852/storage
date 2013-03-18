package storage

import (
	"html/template"
	"io"
)

//Init init the dataCenter and repositories
func Init() { /*{{{*/
	dataCenter = &storage{
		requestCh: make(chan *request),
		data:      make(map[string]Poster),
	}
	go dataCenter.serve()
	initRepos()
} /*}}}*/

const ( /*{{{*/
	TitleAndDateSeperator = "|"
	TimePattern           = "2006-01-02"
	//get repo config file
	ConfigPath = "/app/conf/repos.xml"
) /*}}}*/

//Releaser release a reference
type Releaser interface { /*{{{*/
	Release() string
} /*}}}*/

//Keyer represent a key to post
type Keyer interface { /*{{{*/
	Key() string
} /*}}}*/

//Poster represet a post
type Poster interface { /*{{{*/
	Date() template.HTML
	Content() template.HTML
	Title() template.HTML
	Keyer
	Static(string) io.Reader
	Update(io.Reader) error
} /*}}}*/

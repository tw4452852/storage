package storage

import (
	"errors"
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

//Repository represent a repostory
type Repository interface { /*{{{*/
	//used for setup a repository
	Setup(user, password string) error
	//used for updating a repository
	Refresh()
	//used for uninstall a repostory
	Uninstall()
} /*}}}*/

//used for Init a repository with a root path
type InitFunction func(root string) Repository

var supportedRepoTypes = make(map[string]InitFunction)

//RegisterRepoType register a support repository type
//If there is one, just update it
func RegisterRepoType(key string, f InitFunction) { /*{{{*/
	supportedRepoTypes[key] = f
} /*}}}*/

//UnregisterRepoType unregister a support repository type
func UnregisterRepoType(key string) { /*{{{*/
	delete(supportedRepoTypes, key)
} /*}}}*/

type repos map[string]Repository

func (rs repos) refresh(cfg *Configs) { /*{{{*/
	refreshed := make(map[string]bool)
	for key := range rs {
		refreshed[key] = false
	}

	for _, c := range cfg.Content {
		kind := c.Type
		root := c.Root
		key := kind + "-" + root
		r, found := rs[key]
		if !found {
			if initF, supported := supportedRepoTypes[kind]; supported {
				repo := initF(root)
				if err := repo.Setup(c.User, c.Password); err != nil {
					log.Printf("add repo: setup failed with err(%s)\n", err)
					continue
				}
				log.Printf("add a repo(%q)\n", key)
				rs[key] = repo
				//refresh when init a repo
				repo.Refresh()
			} else {
				log.Printf("add repo: type(%s) isn't supported yet\n",
					kind)
			}
			continue
		}
		r.Refresh()
		refreshed[key] = true
	}

	//uninstall the repos that have been remove
	for key, exist := range refreshed {
		if !exist {
			rs[key].Uninstall()
			delete(rs, key)
		}
	}
} /*}}}*/

var repositories repos

func initRepos(configPath string) { /*{{{*/
	repositories = make(repos)
	go checkConfig(repositories, configPath)
} /*}}}*/

func checkConfig(r repos, configPath string) { /*{{{*/
	//refresh every 10s
	timer := time.NewTicker(10 * time.Second)
	cpath := configPath
	if !filepath.IsAbs(cpath) {
		cpath = filepath.Join(os.Getenv("GOPATH"), cpath)
	}
	for _ = range timer.C {
		cfg, err := getConfig(cpath)
		if err != nil {
			//if there is some error(e.g. file doesn't exist) while reading
			//config file, just skip this refresh
			continue
		}
		r.refresh(cfg)
	}
	panic("not reach")
} /*}}}*/

// meta contain the necessary infomations of a post
type meta struct {
	key     string
	title   string
	date    time.Time
	content template.HTML
}

//post represent a poster instance
type post struct { /*{{{*/
	gen Generator

	sync.RWMutex
	meta
} /*}}}*/

func newPost(gen Generator) *post { /*{{{*/
	if gen == nil {
		panic("a nil Generator in newPost!")
	}
	return &post{
		gen: gen,
	}
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

func (p *post) Update(r io.Reader) error { /*{{{*/
	e, m := p.gen.Generate(r)
	if e != nil {
		return e
	}
	//update meta
	p.Lock()
	p.meta = *m
	p.Unlock()
	m = nil

	return nil
} /*}}}*/

type StaticErr string

//implement io.Reader
func (sr StaticErr) Read(p []byte) (int, error) { /*{{{*/
	log.Println(sr)
	return 0, errors.New(string(sr))
} /*}}}*/

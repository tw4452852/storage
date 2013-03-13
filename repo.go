package storage

import (
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/russross/blackfriday"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//Repository represent a repostory
type Repository interface { /*{{{*/
	//used for setup a repository
	Setup() error
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
				if err := repo.Setup(); err != nil {
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

func initRepos() { /*{{{*/
	repositories = make(repos)
	go checkConfig(repositories)
} /*}}}*/

func checkConfig(r repos) { /*{{{*/
	//refresh every 10s
	timer := time.NewTicker(10 * time.Second)
	for _ = range timer.C {
		cfg, err := getConfig(filepath.Join(os.Getenv("GOPATH"), ConfigPath))
		if err != nil {
			//if there is some error(e.g. file doesn't exist) while reading
			//config file, just skip this refresh
			continue
		}
		r.refresh(cfg)
	}
	panic("not reach")
} /*}}}*/

func generateAll(reader io.Reader) (key, title string, date time.Time, content template.HTML, err error) { /*{{{*/
	c, e := ioutil.ReadAll(reader)
	if e != nil {
		err = e
		return
	}
	//generate title and date
	firstLineIndex := strings.Index(string(c), "\n")
	if firstLineIndex == -1 {
		err = errors.New("generateAll: there must be at least one line\n")
		return
	}
	firstLine := strings.TrimSpace(string(c[:firstLineIndex]))
	remain := strings.TrimSpace(string(c[firstLineIndex+1:]))
	sepIndex := strings.Index(firstLine, TitleAndDateSeperator)
	if sepIndex == -1 {
		err = errors.New("generateAll: can't find seperator for title and date\n")
		return
	}
	t, e := time.Parse(TimePattern, strings.TrimSpace(firstLine[sepIndex+1:]))
	if e != nil {
		err = e
		return
	}

	//generate key
	h := md5.New()
	io.WriteString(h, string(c))
	key = fmt.Sprintf("%x", h.Sum(nil))

	title = strings.TrimSpace(firstLine[:sepIndex])
	date = t
	content = template.HTML(blackfriday.MarkdownCommon([]byte(remain)))
	return
} /*}}}*/

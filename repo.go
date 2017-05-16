package storage

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	rdebug "runtime/debug"
	"time"
)

// Repository represent a repostory
type Repository interface {
	// used for setup a repository
	Setup(user, password string) error
	// used for updating a repository
	Refresh()
	// used for uninstall a repostory
	Uninstall()
}

// used for Init a repository with a root path
type InitFunction func(root string) Repository

var supportedRepoTypes = make(map[string]InitFunction)

// RegisterRepoType register a support repository type
// If there is one, just update it
func RegisterRepoType(key string, f InitFunction) {
	supportedRepoTypes[key] = f
}

// UnregisterRepoType unregister a support repository type
func UnregisterRepoType(key string) {
	delete(supportedRepoTypes, key)
}

type repos map[string]Repository

func (rs repos) refresh(cfg *Configs) {
	// handle github repo update panic
	defer func() {
		if e := recover(); e != nil {
			log.Printf("refresh panic recovered: %s\n%s\n", e, rdebug.Stack())
		}
	}()

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
				// refresh when init a repo
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

	// uninstall the repos that have been remove
	for key, exist := range refreshed {
		if !exist {
			rs[key].Uninstall()
			delete(rs, key)
		}
	}
}

var repositories repos

func initRepos(configPath string) {
	repositories = make(repos)
	go checkConfig(repositories, configPath)
}

func checkConfig(r repos, configPath string) {
	// refresh every 10s
	timer := time.NewTicker(10 * time.Second)
	cpath := configPath
	if !filepath.IsAbs(cpath) {
		cpath = filepath.Join(os.Getenv("GOPATH"), cpath)
	}
	for range timer.C {
		cfg, err := getConfig(cpath)
		if err != nil {
			// if there is some error(e.g. file doesn't exist) while reading
			// config file, just skip this refresh
			continue
		}
		r.refresh(cfg)
	}
	panic("not reach")
}

type StaticErr string

// implement io.Reader
func (sr StaticErr) Read(p []byte) (int, error) {
	log.Println(sr)
	return 0, errors.New(string(sr))
}

func (sr StaticErr) Close() error {
	return nil
}

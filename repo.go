package storage

import (
	"log"
	"time"
)

// Repository represent a repostory
type Repository interface {
	// Install the repository
	Install(user, password string) error
	// Update the repository's contents in storage
	Refresh(s Storager)
	// Uninstall the repostory
	Uninstall(s Storager)
}

// Creator creates a repository with a root path
type Creator func(root string) (Repository, error)

var supportedRepoTypes = make(map[string]Creator)

// RegisterRepoType register a support repository type
// If there is one, just update it
func RegisterRepoType(t string, f Creator) {
	supportedRepoTypes[t] = f
}

// UnregisterRepoType unregister a support repository type
func UnregisterRepoType(t string) {
	delete(supportedRepoTypes, t)
}

func newRepos(configPath string, s Storager) ([]Repository, error) {
	cfg, err := getConfig(configPath)
	if err != nil {
		return nil, err
	}

	var rs []Repository
	for _, c := range cfg {
		kind := c.Type
		root := c.Root
		if create, supported := supportedRepoTypes[kind]; supported {
			repo, err := create(root)
			if err != nil {
				log.Printf("create repo failed: %s\n", err)
				continue
			}

			if err := repo.Install(c.User, c.Password); err != nil {
				log.Printf("install repo failed: %s\n", err)
				continue
			}

			rs = append(rs, repo)
			log.Printf("add a repo, type:%s, root:%s\n", kind, root)
		} else {
			log.Printf("add repo: type(%s) isn't supported yet\n", kind)
		}
	}

	startRepoChecker(rs, s)

	return rs, nil
}

func startRepoChecker(rs []Repository, s Storager) {
	for _, repo := range rs {
		go func(repo Repository) {
			c := time.Tick(1 * time.Second)
			for range c {
				repo.Refresh(s)
			}
		}(repo)
	}
}

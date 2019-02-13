package storage

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

func init() {
	RegisterRepoType("local", newLocalRepo)
}

type localRepo struct {
	root  string
	posts map[string]*localPost
}

func newLocalRepo(root string) (Repository, error) {
	// root must exit
	fi, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	// root must be a dir
	if !fi.IsDir() {
		return nil, errors.New("you can't specify a file as a repo root")
	}
	return &localRepo{
		root:  root,
		posts: make(map[string]*localPost),
	}, nil
}

// implement the Repository interface
func (lr *localRepo) Install(user, password string) error {
	// TODO: verify user and password pair
	return nil
}

func (lr *localRepo) Uninstall(s Storager) {
	// delete repo's post in the dataCenter
	cleans := make([]Keyer, 0, len(lr.posts))
	for _, p := range lr.posts {
		cleans = append(cleans, p)
	}
	if err := s.Remove(cleans...); err != nil {
		log.Printf("remove all the posts in local repo(%s) failed: %s\n",
			lr.root, err)
	}
}

func (lr *localRepo) Refresh(s Storager) {
	// delete the removed files
	lr.clean(s)
	// add newer post and update the exist post
	lr.update(s)
}

// clean the noexist posts
func (lr *localRepo) clean(s Storager) {
	cleans := make([]Keyer, 0)
	for relPath, p := range lr.posts {
		absPath := filepath.Join(lr.root, relPath)
		_, err := os.Stat(absPath)
		if err != nil && os.IsNotExist(err) {
			cleans = append(cleans, p)
			delete(lr.posts, relPath)
		}
	}
	if len(cleans) != 0 {
		if err := s.Remove(cleans...); err != nil {
			log.Printf("remove local post failed: %s\n", err)
		}
	}
}

// update add new post or update the exist ones
func (lr *localRepo) update(s Storager) {
	if err := filepath.Walk(lr.root, func(path string, info os.FileInfo, err error) error {
		// only focus on regular files
		if info.IsDir() {
			return nil
		}
		relPath, _ := filepath.Rel(lr.root, path)
		post, found := lr.posts[relPath]
		if !found {
			post = newLocalPost(path)
			if post == nil {
				return nil
			}
			lr.posts[relPath] = post
			dprintf("Add a new local post(%s)\n", path)
		}
		// update an existing one
		if e := post.update(s); e != nil {
			log.Printf("Update a local post(%s) failed: %s\n", path, e)
		}
		return nil
	}); err != nil {
		log.Printf("Walk local repo(%s) error: %s\n",
			lr.root, err)
	}
}

// represet a local post
type localPost struct {
	Poster
	path       string
	gen        Generator
	lastUpdate time.Time
}

func newLocalPost(path string) *localPost {
	gen := FindGenerator(path)
	if gen == nil {
		return nil
	}
	return &localPost{
		path: path,
		gen:  gen,
	}
}

func (lp *localPost) update(s Storager) error {
	file, err := os.Open(lp.path)
	if err != nil {
		return err
	}
	defer file.Close()
	fi, err := file.Stat()
	if err != nil {
		return err
	}
	if ut := fi.ModTime(); ut.After(lp.lastUpdate) {
		p, err := lp.gen.Generate(file, lp)
		if err != nil {
			return err
		}
		// remove the old one if any
		if lp.Poster != nil {
			err = s.Remove(lp)
			if err != nil {
				return err
			}
		}
		// add the new one
		lp.Poster = p
		err = s.Add(lp)
		if err != nil {
			return nil
		}

		lp.lastUpdate = ut
	}
	return nil
}

// Implement localPost's Static interface
func (lp *localPost) Static(path string) io.ReadCloser {
	path = filepath.FromSlash(path)
	if !filepath.IsAbs(path) {
		path = filepath.Join(filepath.Dir(lp.path), path)
	}
	file, err := os.Open(path)
	if err != nil {
		return StaticErr(fmt.Sprintf("open %q file failed: %s\n",
			path, err))
	}
	return file
}

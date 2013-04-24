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

func init() { /*{{{*/
	RegisterRepoType("local", NewLocalRepo)
} /*}}}*/

type localRepo struct { /*{{{*/
	root  string
	posts map[string]*localPost
} /*}}}*/

func NewLocalRepo(root string) Repository { /*{{{*/
	return &localRepo{
		root:  root,
		posts: make(map[string]*localPost),
	}
} /*}}}*/

//implement the Repository interface
func (lr *localRepo) Setup(user, password string) error { /*{{{*/
	//root Must be a dir
	fi, err := os.Stat(lr.root)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return errors.New("you can't specify a file as a repo root")
	}
	return nil
} /*}}}*/

func (lr *localRepo) Uninstall() { /*{{{*/
	//delete repo's post in the dataCenter
	cleans := make([]Keyer, 0, len(lr.posts))
	for _, p := range lr.posts {
		cleans = append(cleans, p)
	}
	if err := Remove(cleans...); err != nil {
		log.Printf("remove all the posts in local repo(%s) failed: %s\n",
			lr.root, err)
	}
} /*}}}*/

func (lr *localRepo) Refresh() { /*{{{*/
	//delete the removed files
	lr.clean()
	//add newer post and update the exist post
	lr.update()
} /*}}}*/

//clean the noexist posts
func (lr *localRepo) clean() { /*{{{*/
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
		if err := Remove(cleans...); err != nil {
			log.Printf("remove local post failed: %s\n", err)
		}
	}
} /*}}}*/

//update add new post or update the exist ones
func (lr *localRepo) update() { /*{{{*/
	updateLocalPost := func(lp *localPost) {
		if err := lp.update(); err != nil {
			log.Printf("update local post(%s) failed: %s\n", lp.path, err)
		}
	}

	if err := filepath.Walk(lr.root, func(path string, info os.FileInfo, err error) error {
		//only watch the special filetype
		if info.IsDir() || !filetypeFilter(path) {
			return nil
		}
		relPath, _ := filepath.Rel(lr.root, path)
		post, found := lr.posts[relPath]
		if !found {
			lp := newLocalPost(path)
			lr.posts[relPath] = lp
			updateLocalPost(lp)
			return nil
		}
		//update a exist one
		updateLocalPost(post)
		return nil
	}); err != nil {
		log.Printf("update local repo(%s) error: %s\n",
			lr.root, err)
	}
} /*}}}*/

//represet a local post
type localPost struct { /*{{{*/
	path       string
	lastUpdate time.Time
	*post
} /*}}}*/

func newLocalPost(path string) *localPost { /*{{{*/
	return &localPost{
		path: path,
		post: newPost(),
	}
} /*}}}*/

func (lp *localPost) update() error { /*{{{*/
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
		if err := lp.Update(file); err != nil {
			return err
		}
		lp.lastUpdate = ut
		//update the content in dataCenter
		if err := Add(lp); err != nil {
			log.Printf("update a local post failed: %s\n", err)
		}
	}
	return nil
} /*}}}*/

//Implement localPost's Static interface
func (lp *localPost) Static(path string) io.Reader { /*{{{*/
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
} /*}}}*/

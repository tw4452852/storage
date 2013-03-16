package storage

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
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
func (lr *localRepo) Setup() error { /*{{{*/
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
	//nothing to do
} /*}}}*/

func (lr *localRepo) Refresh() { /*{{{*/
	//delete the removed files
	lr.clean()
	//add newer post and update the exist post
	lr.update()
} /*}}}*/

//clean the noexist posts
func (lr *localRepo) clean() { /*{{{*/
	cleans := make([]string, 0)
	for relPath := range lr.posts {
		absPath := filepath.Join(lr.root, relPath)
		_, err := os.Stat(absPath)
		if err != nil && os.IsNotExist(err) {
			cleans = append(cleans, relPath)
		}
	}
	for _, relPath := range cleans {
		lp := lr.posts[relPath]
		if err := Remove(lp); err != nil {
			log.Printf("remove local post failed: %s\n", err)
			continue
		}
		log.Printf("remove a local post: %s\n", lp.path)
		delete(lr.posts, relPath)
	}
} /*}}}*/

//update add new post or update the exist ones
func (lr *localRepo) update() { /*{{{*/
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
			if err := lp.update(); err != nil {
				log.Printf("update local post(%s) failed: %s\n", lp.path, err)
			}
			return nil
		}
		//update a exist one
		if err := post.update(); err != nil {
			log.Printf("update local post(%s) failed: %s\n", post.path, err)
		}
		return nil
	}); err != nil {
		log.Printf("update local repo(%s) error: %s\n",
			lr.root, err)
	}
} /*}}}*/

//supported filetype
var filters = []*regexp.Regexp{ /*{{{*/
	regexp.MustCompile(".*.md$"),
} /*}}}*/

//filter file type , return pass
func filetypeFilter(path string) (passed bool) { /*{{{*/
	for _, filter := range filters {
		if filter.MatchString(path) {
			return true
		}
	}
	return false
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
		log.Printf("update a local post: path(%s), key(%x), date(%s)\n",
			lp.path, lp.Key(), lp.Date())
	}
	return nil
} /*}}}*/

//Implement localPost's Static interface
func (lp *localPost) Static(path string) io.Reader { /*{{{*/
	path = filepath.FromSlash(path)
	if !filepath.IsAbs(path) {
		path = filepath.Join(lp.path, path)
	}
	file, err := os.Open(path)
	if err != nil {
		return StaticErr(fmt.Sprintf("open %q file failed: %s\n",
			path, err))
	}
	return file
} /*}}}*/

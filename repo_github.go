package storage

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"path"
	"sort"

	"github.com/google/go-github/v29/github"
	"github.com/gregjones/httpcache"
)

func init() {
	RegisterRepoType("github", newGithubRepo)
}

var nilError = errors.New("nil error")

type githubRepo struct {
	client   *github.Client
	owner    string
	name     string
	posts    map[string]*githubPost
	lastSHA1 string
}

func newGithubRepo(name string) (Repository, error) {
	return &githubRepo{
		name:  name,
		posts: make(map[string]*githubPost),
	}, nil
}

// Implement the Repository interface
func (gr *githubRepo) Install(user, password string) error {
	// TODO:	oauth2
	gr.client = github.NewClient(&http.Client{
		Transport: &github.BasicAuthTransport{
			Username:  user,
			Password:  password,
			Transport: httpcache.NewMemoryCacheTransport(),
		}})
	gr.owner = user
	return nil
}

func (gr *githubRepo) Uninstall(s Storager) {
	// delete repo's post in the dataCenter
	cleans := make([]Keyer, 0, len(gr.posts))
	for _, p := range gr.posts {
		cleans = append(cleans, p)
	}
	if err := s.Remove(cleans...); err != nil {
		log.Printf("remove all the posts in github repo(%s/%s) failed: %s\n",
			gr.owner, gr.name, err)
	}
}

func (gr *githubRepo) Refresh(s Storager) {
	// get master's sha1
	sha1, res, err := gr.client.Repositories.GetCommitSHA1(context.Background(), gr.owner, gr.name, "master", gr.lastSHA1)
	if err != nil {
		log.Printf("failed to get SHA1 of master: error[%#v], res[%#v]\n", err, res)
		return
	}
	// skip Refresh if there's nothing new
	if sha1 == gr.lastSHA1 {
		return
	}
	// get that commit according to the sha1
	commit, res, err := gr.client.Repositories.GetCommit(context.Background(), gr.owner, gr.name, sha1)
	if err != nil {
		log.Printf("failed to get commit of master: error[%#v], res[%#v]\n", err, res)
		return
	}
	// get all the files according to tree's sha1
	treeSha1 := commit.GetCommit().GetTree().GetSHA()
	tree, res, err := gr.client.Git.GetTree(context.Background(), gr.owner, gr.name, treeSha1, true)
	if err != nil {
		log.Printf("failed to get tree of master: error[%#v], res[%#v]\n", err, res)
		return
	}
	treeArray := tree.Entries
	paths := make([]string, 0)
	for i := range treeArray {
		path := treeArray[i].GetPath()
		if FindGenerator(path) == nil {
			continue
		}
		paths = append(paths, path)
	}
	sort.Strings(paths)
	// delete the no exist posts
	gr.clean(s, paths)
	// add new post and update the exist ones
	gr.update(s, paths)

	gr.lastSHA1 = sha1
}

// the paths has been sorted in increasing order
func (gr *githubRepo) clean(s Storager, paths []string) {
	cleans := make([]Keyer, 0)
	for relPath, p := range gr.posts {
		i := sort.SearchStrings(paths, relPath)
		if i >= len(paths) || paths[i] != relPath {
			cleans = append(cleans, p)
			delete(gr.posts, relPath)
		}
	}
	if len(cleans) != 0 {
		if err := s.Remove(cleans...); err != nil {
			log.Printf("remove github post failed: %s\n", err)
		}
	}
}

// the paths has been sorted in increasing order
func (gr *githubRepo) update(s Storager, paths []string) {
	for _, path := range paths {
		post, found := gr.posts[path]
		if !found {
			post = newGithubPost(path, gr)
			gr.posts[path] = post
			dprintf("Add a new github post(%s)\n", path)
		}
		// update a exist one
		if e := post.update(s); e != nil {
			log.Printf("Update a github post(%s) failed: %s\n", path, e)
		}
	}
}

type githubPost struct {
	Poster
	repo *githubRepo
	path string
	gen  Generator
}

func newGithubPost(path string, repo *githubRepo) *githubPost {
	return &githubPost{
		repo: repo,
		path: path,
		gen:  FindGenerator(path),
	}
}

func (gp *githubPost) update(s Storager) error {
	rc, err := gp.repo.client.Repositories.DownloadContents(context.Background(), gp.repo.owner, gp.repo.name, gp.path, nil)
	if err != nil {
		return err
	}

	p, err := gp.gen.Generate(rc, gp)
	if err != nil {
		return err
	}
	// remove the old one if any
	if gp.Poster != nil {
		err = s.Remove(gp)
		if err != nil {
			return err
		}
	}
	gp.Poster = p
	// add the new one
	err = s.Add(gp)
	if err != nil {
		return err
	}
	dprintf("update a github post(%s)\n", gp.path)
	return nil
}

func (gp *githubPost) Static(p string) io.ReadCloser {
	p = path.Join(path.Dir(gp.path), p)
	rc, err := gp.repo.client.Repositories.DownloadContents(context.Background(), gp.repo.owner, gp.repo.name, p, nil)
	if err != nil {
		log.Printf("failed to get static resource[%s]: %v\n", p, err)
		return nil
	}
	return rc
}

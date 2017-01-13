package storage

import (
	"encoding/base64"
	"errors"
	"fmt"
	ghc "github.com/alcacoop/go-github-client/client"
	"io"
	"log"
	"net/http"
	"path"
	"sort"
	"strings"
	"sync"
	"time"
)

func init() {
	RegisterRepoType("github", NewGithubRepo)
}

var nilError = errors.New("nil error")

type githubRepo struct {
	client *githubClient
	name   string
	user   string
	posts  map[string]*githubPost
}

//A wrapper of *ghc.GithubClient
//Used to avoid the exceed the limit
type githubClient struct {
	lock  sync.RWMutex
	cache map[string]*githubResource
	*ghc.GithubClient
}

//githubResource represet a github content
//according to the url, etag used for a
//conditional request
type githubResource struct {
	ghc.JsonMap
	etag string
}

func newGithubClient(client *ghc.GithubClient) *githubClient {
	gc := &githubClient{
		cache:        make(map[string]*githubResource),
		GithubClient: client,
	}
	go gc.refresh()
	return gc
}

//refresh periodically the cache
func (gc *githubClient) refresh() {
	timer := time.NewTicker(1 * time.Minute)
	for range timer.C {
		for url := range gc.cache {
			gr, err := execApi(gc, url, true)
			if err != nil {
				log.Printf("refresh github post(%s) failed: %s\n",
					url, err)
				continue
			}
			if err == nil && gr == nil {
				//nothing to be updated
				continue
			}
			gc.lock.Lock()
			gc.cache[url] = gr
			gc.lock.Unlock()
		}
	}
}

//get the appointed resource according to the url
//If there is a cache, just return it, otherwise,
//emit a api request and update the cache
//if it is a conditional request, return nil, nil when succeed
func (gc *githubClient) get(url string) (ghc.JsonMap, error) {
	gc.lock.RLock()
	gr, found := gc.cache[url]
	if found {
		gc.lock.RUnlock()
		return gr.JsonMap, nil
	}
	//if not found, release the Rlock first,
	//prepare to emit a api request
	gc.lock.RUnlock()
	gr, err := execApi(gc, url, false)
	if err != nil {
		return nil, err
	}
	//save result into cache
	gc.lock.Lock()
	defer gc.lock.Unlock()
	gc.cache[url] = gr
	return gr.JsonMap, nil
}

func execApi(client *githubClient, url string, isCondition bool) (*githubResource, error) {
	req, err := client.NewAPIRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if isCondition {
		etag := client.cache[url].etag
		if etag == "" {
			return nil,
				fmt.Errorf("there is no etag for condition request(%s)", url)
		}
		req.Header.Set("If-None-Match", etag)
	}
	res, err := client.RunRequest(req, new(http.Client))
	if err != nil {
		return nil, err
	}

	//check if nothing is modified
	if isCondition &&
		res.RawHttpResponse.Status == "304 Not Modified" {
		return nil, nil
	}

	etag := res.RawHttpResponse.Header.Get("ETag")
	if etag == "" {
		return nil, errors.New("ETag is null")
	}
	m, err := res.JsonMap()
	if err != nil {
		return nil, err
	}
	return &githubResource{
		JsonMap: m,
		etag:    etag,
	}, nil
}

func NewGithubRepo(name string) Repository {
	return &githubRepo{
		name:  name,
		posts: make(map[string]*githubPost),
	}
}

//Implement the Repository interface
func (gr *githubRepo) Setup(user, password string) error {
	//TODO:	oauth2
	client, err := ghc.NewGithubClient(user, password,
		ghc.AUTH_USER_PASSWORD)
	if err != nil {
		return err
	}
	gr.client = newGithubClient(client)
	gr.user = user
	return nil
}

func (gr *githubRepo) Uninstall() {
	//delete repo's post in the dataCenter
	cleans := make([]Keyer, 0, len(gr.posts))
	for _, p := range gr.posts {
		cleans = append(cleans, p)
	}
	if err := Remove(cleans...); err != nil {
		log.Printf("remove all the posts in github repo(%s/%s) failed: %s\n",
			gr.user, gr.name, err)
	}
}

func (gr *githubRepo) Refresh() {
	//get the master branch post list
	//1.get master branch tree sha
	//2.filter tree to get support file
	//if there is some error happened, just abort and do nothing
	master, err := gr.client.get(
		"repos/" + gr.user + "/" + gr.name + "/branches/master")
	if err != nil {
		log.Println(err)
		return
	}
	c := master.GetMap("commit")
	if c == nil {
		log.Println(nilError, master)
		return
	}
	cc := c.GetMap("commit")
	if cc == nil {
		log.Println(nilError, c)
		return
	}
	t := cc.GetMap("tree")
	if c == nil {
		log.Println(nilError, cc)
		return
	}
	sha := t.GetString("sha")
	if sha == "" {
		log.Println(nilError, cc)
		return
	}
	tree, err := gr.client.get(
		"repos/" + gr.user + "/" + gr.name + "/git/trees/" + sha + "?recursive=1")
	if err != nil {
		log.Println(err)
		return
	}
	treeArray := tree.GetArray("tree")
	if treeArray == nil {
		log.Println(nilError, tree)
		return
	}
	paths := make([]string, 0)
	for i := range treeArray {
		path := treeArray.GetObject(i).GetString("path")
		if FindGenerator(path) == nil {
			continue
		}
		paths = append(paths, path)
	}
	sort.Strings(paths)
	//delete the no exist posts
	gr.clean(paths)
	//add new post and update the exist ones
	gr.update(paths)
}

//the paths has been sorted in increasing order
func (gr *githubRepo) clean(paths []string) {
	cleans := make([]Keyer, 0)
	for relPath, p := range gr.posts {
		i := sort.SearchStrings(paths, relPath)
		if i >= len(paths) || paths[i] != relPath {
			cleans = append(cleans, p)
			delete(gr.posts, relPath)
		}
	}
	if len(cleans) != 0 {
		if err := Remove(cleans...); err != nil {
			log.Printf("remove github post failed: %s\n", err)
		}
	}
}

//the paths has been sorted in increasing order
func (gr *githubRepo) update(paths []string) {
	for _, path := range paths {
		post, found := gr.posts[path]
		if !found {
			gp := newGithubPost(path, gr)
			gr.posts[path] = gp
			if debug {
				log.Printf("Add a new github post(%s)\n", path)
			}
			if e := gp.Update(); e != nil {
				log.Printf("Add a new github post(%s) failed: %s\n", path, e)
			}
			continue
		}
		//update a exist one
		if e := post.Update(); e != nil {
			log.Printf("Update a github post(%s) failed: %s\n", path, e)
		}
	}
}

func (gr *githubRepo) static(path string) io.Reader {
	m, err := gr.client.get(
		"repos/" + gr.user + "/" + gr.name + "/contents/" + path)
	if err != nil {
		return StaticErr(fmt.Sprintf("get static file %q failed: %s\n",
			path, err))
	}
	return base64.NewDecoder(base64.StdEncoding,
		strings.NewReader(strings.Replace(m.GetString("content"), "\n", "", -1)))
}

type githubPost struct {
	repo *githubRepo
	path string
	sha  string
	*post
	Generator
}

func newGithubPost(path string, repo *githubRepo) *githubPost {
	return &githubPost{
		repo:      repo,
		path:      path,
		post:      newPost(),
		Generator: FindGenerator(path),
	}
}

func (gp *githubPost) Update() error {
	ms, err := gp.repo.client.get(
		"repos/" + gp.repo.user + "/" + gp.repo.name + "/contents/" + gp.path)
	if err != nil {
		return err
	}
	sha, encodedContent := ms.GetString("sha"), ms.GetString("content")
	if sha == gp.sha {
		//no need to update
		return nil
	}
	encodedContent = strings.Replace(encodedContent, "\n", "", -1)
	var m *meta
	err, m = gp.Generate(base64.NewDecoder(base64.StdEncoding,
		strings.NewReader(encodedContent)), gp)
	if err != nil {
		return err
	}
	gp.update(m)
	if debug {
		log.Printf("update a github post(%s)\n", gp.path)
	}
	//add it to the dataCenter
	if err = Add(gp); err != nil {
		return err
	}
	//update it sha
	gp.sha = sha
	return nil
}

func (gp *githubPost) Static(p string) io.Reader {
	p = path.Join(path.Dir(gp.path), p)
	return gp.repo.static(p)
}

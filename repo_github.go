package storage

import (
	"encoding/base64"
	"fmt"
	ghc "github.com/alcacoop/go-github-client/client"
	"io"
	"log"
	"net/http"
	"path"
	"sort"
	"strings"
)

func init() {
	RegisterRepoType("github", NewGithubRepo)
}

type githubRepo struct {
	client *ghc.GithubClient
	name   string
	user   string
	posts  map[string]*githubPost
}

func NewGithubRepo(name string) Repository { /*{{{*/
	return &githubRepo{
		name:  name,
		posts: make(map[string]*githubPost),
	}
} /*}}}*/

//Implement the Repository interface
func (gr *githubRepo) Setup(user, password string) error { /*{{{*/
	//TODO:	oauth2
	client, err := ghc.NewGithubClient(user, password,
		ghc.AUTH_USER_PASSWORD)
	if err != nil {
		return err
	}
	gr.client = client
	gr.user = user
	return nil
} /*}}}*/

func (gr *githubRepo) Uninstall() { /*{{{*/
	//nothing to clear
} /*}}}*/

func execAPI(client *ghc.GithubClient, url string) (ghc.JsonMap, error) { /*{{{*/
	req, err := client.NewAPIRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	res, err := client.RunRequest(req, new(http.Client))
	if err != nil {
		return nil, err
	}
	return res.JsonMap()
} /*}}}*/

func (gr *githubRepo) Refresh() { /*{{{*/
	//get the master branch post list
	//1.get master branch tree sha
	//2.filter tree to get support file
	//if there is some error happened, just abort and do nothing
	master, err := execAPI(gr.client,
		"repos/"+gr.user+"/"+gr.name+"/branches/master")
	if err != nil {
		log.Println(err)
		return
	}
	sha := master.GetMap("commit").GetMap("commit").GetMap("tree").GetString("sha")
	tree, err := execAPI(gr.client,
		"repos/"+gr.user+"/"+gr.name+"/git/trees/"+sha+"?recursive=1")
	if err != nil {
		log.Println(err)
		return
	}
	treeArray := tree.GetArray("tree")
	paths := make([]string, 0)
	for i := range treeArray {
		path := treeArray.GetObject(i).GetString("path")
		if !filetypeFilter(path) {
			continue
		}
		paths = append(paths, path)
	}
	sort.Strings(paths)
	//delete the no exist posts
	gr.clean(paths)
	//add new post and update the exist ones
	gr.update(paths)
} /*}}}*/

//the paths has been sorted in increasing order
func (gr *githubRepo) clean(paths []string) { /*{{{*/
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
} /*}}}*/

//the paths has been sorted in increasing order
func (gr *githubRepo) update(paths []string) { /*{{{*/
	updateGithubPost := func(gp *githubPost, path string) {
		m, err := execAPI(gr.client,
			"repos/"+gr.user+"/"+gr.name+"/contents/"+path)
		if err != nil {
			log.Printf("get github post(%s) content failed: %s\n",
				path, err)
			return
		}
		if err := gp.update(m.GetString("sha"), m.GetString("content")); err != nil {
			log.Printf("update github post(%s) failed: %s\n", path, err)
		}
	}

	for _, path := range paths {
		post, found := gr.posts[path]
		if !found {
			gp := newGithubPost(path, gr)
			gr.posts[path] = gp
			log.Printf("add a new github post(%s)\n", path)
			updateGithubPost(gp, path)
			continue
		}
		//update a exist one
		updateGithubPost(post, path)
	}
} /*}}}*/

func (gr *githubRepo) static(path string) io.Reader { /*{{{*/
	m, err := execAPI(gr.client,
		"repos/"+gr.user+"/"+gr.name+"/contents/"+path)
	if err != nil {
		return StaticErr(fmt.Sprintf("get static file %q failed: %s\n",
			path, err))
	}
	return base64.NewDecoder(base64.StdEncoding,
		strings.NewReader(strings.Replace(m.GetString("content"), "\n", "", -1)))
} /*}}}*/

type githubPost struct { /*{{{*/
	repo *githubRepo
	path string
	sha  string
	*post
} /*}}}*/

func newGithubPost(path string, repo *githubRepo) *githubPost { /*{{{*/
	return &githubPost{
		repo: repo,
		path: path,
		post: newPost(),
	}
} /*}}}*/

func (gp *githubPost) update(sha, encodedContent string) error { /*{{{*/
	if sha == gp.sha {
		//no need to update
		return nil
	}
	encodedContent = strings.Replace(encodedContent, "\n", "", -1)
	err := gp.Update(base64.NewDecoder(base64.StdEncoding,
		strings.NewReader(encodedContent)))
	if err != nil {
		return err
	}
	log.Printf("update a github post(%s)\n", gp.path)
	//add it to the dataCenter
	if err = Add(gp); err != nil {
		return err
	}
	//update it sha
	gp.sha = sha
	return nil
} /*}}}*/

func (gp *githubPost) Static(p string) io.Reader { /*{{{*/
	p = path.Join(path.Dir(gp.path), p)
	return gp.repo.static(p)
} /*}}}*/

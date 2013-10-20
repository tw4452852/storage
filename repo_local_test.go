package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

var repoRoot = filepath.FromSlash(filepath.Join(os.Getenv("GOPATH"),
	"src/github.com/tw4452852/storage/testdata/localRepo/"))

func TestLocalSetup(t *testing.T) { /*{{{*/
	cases := []struct {
		root   string
		expect error
	}{
		{
			"./testdata/noexist/",
			pathNotFound,
		},
		{
			"./testdata/localRepo.file",
			errors.New("you can't specify a file as a repo root"),
		},
		{
			"./testdata/localRepo/",
			nil,
		},
	}

	for _, c := range cases {
		lr := NewLocalRepo(c.root)
		if e := matchError(c.expect, lr.Setup("", "")); e != nil {
			t.Error(e)
		}
		r := lr.(*localRepo)
		if r.root != c.root {
			t.Errorf("expect repo root(%s), but get(%s)\n",
				c.root, r.root)
		}
	}
} /*}}}*/

func TestLocalRepo(t *testing.T) { /*{{{*/
	repo := NewLocalRepo(repoRoot)
	if err := repo.Setup("", ""); err != nil {
		t.Fatal(err)
	}
	repo.Uninstall()
	lr := repo.(*localRepo)
	cases := []struct {
		prepare func()
		check   func()
	}{
		{
			prepare: nil,
			check: func() {
				expect := map[string]*localPost{
					"1.md":   newLocalPost(filepath.Join(repoRoot, "1.md")),
					"1.prst": newLocalPost(filepath.Join(repoRoot, "1.prst")),
					"level1" + string(filepath.Separator) + "1.md": newLocalPost(filepath.Join(repoRoot, "level1/1.md")),
				}
				lr.update()
				if err := checkLocalPosts(expect, lr.posts); err != nil {
					t.Error(err)
				}
			},
		},

		{
			prepare: func() {
				lr.posts["1.md"] = newLocalPost(filepath.Join(repoRoot, "1.md"))
				lr.posts["noexist.md"] = newLocalPost(filepath.Join(repoRoot, "noexist.md"))
				lr.posts["level1/noexist.md"] = newLocalPost(filepath.Join(repoRoot, "level1/noexist.md"))
			},
			check: func() {
				expect := map[string]*localPost{
					"1.md":   newLocalPost(filepath.Join(repoRoot, "1.md")),
					"1.prst": newLocalPost(filepath.Join(repoRoot, "1.prst")),
					"level1" + string(filepath.Separator) + "1.md": newLocalPost(filepath.Join(repoRoot, "level1/1.md")),
				}
				lr.clean()
				if err := checkLocalPosts(expect, lr.posts); err != nil {
					t.Error(err)
				}
			},
		},
	}

	for _, c := range cases {
		if c.prepare != nil {
			c.prepare()
		}
		if c.check != nil {
			c.check()
		}
	}
} /*}}}*/

func checkLocalPosts(expect, real map[string]*localPost) error { /*{{{*/
	if len(real) != len(expect) {
		return fmt.Errorf("length of posts isn't equal: expect %v but get %v\n",
			expect, real)
	}
	for k, v := range expect {
		r, found := real[k]
		if !found {
			return fmt.Errorf("can't find expect %v in real\n", *v)
		}
		if v.path != r.path {
			return fmt.Errorf("want path %q, but get %q\n", v.path, r.path)
		}
	}
	return nil
} /*}}}*/

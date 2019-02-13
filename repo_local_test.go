package storage

import (
	"errors"
	"fmt"
	"path/filepath"
	"testing"
)

func TestNewLocalRepo(t *testing.T) {
	for name, c := range map[string]struct {
		root   string
		expect error
	}{
		"noExistDir": {
			root:   "./testdata/noexist/",
			expect: pathNotFound,
		},
		"fileRoot": {
			root:   "./testdata/localRepo.file",
			expect: errors.New("you can't specify a file as a repo root"),
		},
		"normal": {
			root: "./testdata/localRepo/",
		},
	} {
		c := c
		t.Run(name, func(t *testing.T) {
			lr, e := newLocalRepo(c.root)
			if e = matchError(c.expect, e); e != nil {
				t.Fatal(e)
			}
			if r, ok := lr.(*localRepo); ok {
				if r.root != c.root {
					t.Errorf("expect repo root(%s), but get(%s)\n",
						c.root, r.root)
				}
			}
		})
	}
}

func TestLocalRepoRefresh(t *testing.T) {
	repo, err := newLocalRepo("./testdata/localRepo/")
	if err != nil {
		t.Fatal(err)
	}
	if err = repo.Install("", ""); err != nil {
		t.Fatal(err)
	}
	lr := repo.(*localRepo)

	for name, c := range map[string]struct {
		prepare map[string]*localPost
		expect  map[string]*localPost
	}{
		"update": {
			prepare: map[string]*localPost{},
			expect: map[string]*localPost{
				"1.md":      newLocalPost(filepath.Join("./testdata/localRepo/", "1.md")),
				"1.article": newLocalPost(filepath.Join("./testdata/localRepo/", "1.article")),
				"1.slide":   newLocalPost(filepath.Join("./testdata/localRepo/", "1.slide")),
				"level1" + string(filepath.Separator) + "1.md": newLocalPost(filepath.Join("./testdata/localRepo/", "level1/1.md")),
			},
		},

		"clean": {
			prepare: map[string]*localPost{
				"1.md":              newLocalPost(filepath.Join("./testdata/localRepo/", "1.md")),
				"noexist.md":        newLocalPost(filepath.Join("./testdata/localRepo/", "noexist.md")),
				"level1/noexist.md": newLocalPost(filepath.Join("./testdata/localRepo/", "level1/noexist.md")),
			},
			expect: map[string]*localPost{
				"1.md":      newLocalPost(filepath.Join("./testdata/localRepo/", "1.md")),
				"1.article": newLocalPost(filepath.Join("./testdata/localRepo/", "1.article")),
				"1.slide":   newLocalPost(filepath.Join("./testdata/localRepo/", "1.slide")),
				"level1" + string(filepath.Separator) + "1.md": newLocalPost(filepath.Join("./testdata/localRepo/", "level1/1.md")),
			},
		},
	} {
		c := c
		t.Run(name, func(t *testing.T) {
			lr.posts = c.prepare
			lr.Refresh(&nopStorage{})
			if err := checkLocalPosts(c.expect, lr.posts); err != nil {
				t.Error(err)
			}
		})
	}
}

func checkLocalPosts(expect, real map[string]*localPost) error {
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
}

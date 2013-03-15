package storage

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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
		if e := matchError(c.expect, lr.Setup()); e != nil {
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
	if err := repo.Setup(); err != nil {
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
					"1.md": newLocalPost(filepath.Join(repoRoot, "1.md")),
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
					"1.md": newLocalPost(filepath.Join(repoRoot, "1.md")),
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

func TestLocalPostUpdate(t *testing.T) { /*{{{*/
	type Expect struct {
		path, title, date, content string
	}
	type Case struct {
		prepare   func()
		clean     func()
		path      string
		updateErr error
		expect    *Expect
	}
	cases := []*Case{
		{
			nil,
			nil,
			"./testdata/noexist/1.md",
			pathNotFound,
			nil,
		},
		{
			func() {
				ioutil.WriteFile(filepath.Join(repoRoot, "11.md"),
					[]byte("hello world | 2012-12-01 \n# title hello world \n"), 0777)
			},
			func() {
				os.Remove(filepath.Join(repoRoot, "11.md"))
			},
			filepath.Join(repoRoot, "11.md"),
			nil,
			&Expect{
				path:    filepath.Join(repoRoot, "11.md"),
				title:   "hello world",
				date:    "2012-12-01",
				content: "<h1>title hello world</h1>\n",
			},
		},
		{
			func() {
				ioutil.WriteFile(filepath.Join(repoRoot, "11.md"),
					[]byte("hello world | 2012-12-01 \n "), 0777)
			},
			func() {
				os.Remove(filepath.Join(repoRoot, "11.md"))
			},
			filepath.Join(repoRoot, "11.md"),
			nil,
			&Expect{
				path:    filepath.Join(repoRoot, "11.md"),
				title:   "hello world",
				date:    "2012-12-01",
				content: "",
			},
		},
		{
			func() {
				ioutil.WriteFile(filepath.Join(repoRoot, "11.md"),
					[]byte(" hello world | 2012-12-01"), 0777)
			},
			func() {
				os.Remove(filepath.Join(repoRoot, "11.md"))
			},
			filepath.Join(repoRoot, "11.md"),
			errors.New("generateAll: there must be at least one line"),
			nil,
		},
		{
			func() {
				ioutil.WriteFile(filepath.Join(repoRoot, "11.md"),
					[]byte(" hello world & 2012-12-01\n"), 0777)
			},
			func() {
				os.Remove(filepath.Join(repoRoot, "11.md"))
			},
			filepath.Join(repoRoot, "11.md"),
			errors.New("generateAll: can't find seperator"),
			nil,
		},
		{
			func() {
				ioutil.WriteFile(filepath.Join(repoRoot, "11.md"),
					[]byte(" hello world || 2012-12-01\n"), 0777)
			},
			func() {
				os.Remove(filepath.Join(repoRoot, "11.md"))
			},
			filepath.Join(repoRoot, "11.md"),
			errors.New("parsing time"),
			nil,
		},
	}

	runCase := func(c *Case) error {
		if c.clean != nil {
			defer c.clean()
		}
		if c.prepare != nil {
			c.prepare()
		}
		lp := newLocalPost(c.path)
		if err := matchError(c.updateErr, lp.update()); err != nil {
			return err
		}
		if c.updateErr != nil && c.expect == nil {
			return nil
		}
		real := &Expect{
			path:    lp.path,
			title:   string(lp.Title()),
			date:    string(lp.Date()),
			content: string(lp.Content()),
		}
		if *real != *c.expect {
			return fmt.Errorf("expect %v, but get %v\n", *c.expect, *real)
		}
		return nil
	}

	for i, c := range cases {
		if err := runCase(c); err != nil {
			t.Errorf("case %d error: %s\n", i, err)
		}
	}
} /*}}}*/

func TestLocalPostStatic(t *testing.T) {
	type Expect struct {
		path string
	}
	type Case struct {
		prepare   func()
		clean     func()
		path      string
		updateErr error
		expect    *Expect
	}
	cases := []*Case{
		{
			func() {
				ioutil.WriteFile(filepath.Join(repoRoot, "11.md"),
					[]byte("hello world | 2012-12-01 \n![1](/1/1.png)\n"), 0777)
			},
			func() {
				os.Remove(filepath.Join(repoRoot, "11.md"))
			},
			filepath.Join(repoRoot, "11.md"),
			nil,
			&Expect{
				"/1/1.png",
			},
		},
		{
			func() {
				ioutil.WriteFile(filepath.Join(repoRoot, "11.md"),
					[]byte("hello world | 2012-12-01 \n![1](1/1.png)\n"), 0777)
			},
			func() {
				os.Remove(filepath.Join(repoRoot, "11.md"))
			},
			filepath.Join(repoRoot, "11.md"),
			nil,
			&Expect{
				filepath.Join(filepath.Join(repoRoot, "11.md"), "/1/1.png"),
			},
		},
	}
	runCase := func(c *Case) error {
		if c.clean != nil {
			defer c.clean()
		}
		if c.prepare != nil {
			c.prepare()
		}
		lp := newLocalPost(c.path)
		if err := matchError(c.updateErr, lp.update()); err != nil {
			return err
		}
		if c.updateErr != nil && c.expect == nil {
			return nil
		}
		expect := lp.Key() + c.expect.path
		content := string(lp.Content())
		if !strings.Contains(content, expect) {
			return fmt.Errorf("can't find %q in %q\n", expect, content)
		}
		return nil
	}

	for i, c := range cases {
		if err := runCase(c); err != nil {
			t.Errorf("case %d error: %s\n", i, err)
		}
	}
}

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

func TestMarkDownGenerate(t *testing.T) { /*{{{*/
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
		if err := matchError(c.updateErr, lp.Update()); err != nil {
			return err
		}
		if c.updateErr != nil && c.expect == nil {
			return nil
		}
		real := &Expect{
			path:    lp.path,
			title:   string(lp.Title()),
			date:    lp.Date().Format(TimePattern),
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

func TestMarkDownImage(t *testing.T) { /*{{{*/
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
				"1/1.png",
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
		if err := matchError(c.updateErr, lp.Update()); err != nil {
			return err
		}
		if c.updateErr != nil && c.expect == nil {
			return nil
		}
		expect := imagePrefix + lp.Key() + "/" + c.expect.path
		content := string(lp.Content())
		if !strings.Contains(content, expect) {
			return fmt.Errorf("can't find (%s) in (%s)\n", expect, content)
		}
		return nil
	}

	for i, c := range cases {
		if err := runCase(c); err != nil {
			t.Errorf("case %d error: %s\n", i, err)
		}
	}
} /*}}}*/

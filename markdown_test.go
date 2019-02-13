package storage

import (
	"errors"
	"strings"
	"testing"
)

func TestMarkDownMatch(t *testing.T) {
	for name, c := range map[string]struct {
		path   string
		expect bool
	}{
		"match": {
			path:   "a/b/c.md",
			expect: true,
		},
		"unmatch": {
			path:   "d/e/f.mm",
			expect: false,
		},
	} {
		c := c
		t.Run(name, func(t *testing.T) {
			if got := (markdownGenerator{}).Match(c.path); got != c.expect {
				t.Errorf("got %v, but want %v\n", got, c.expect)
			}
		})
	}
}

func TestMarkDownGenerate(t *testing.T) {
	for name, c := range map[string]struct {
		input        string
		expectErr    error
		expectResult Poster
	}{
		"normal": {
			input: "hello world | 2012-12-01 |tag1, tag2\n# title hello world \n",
			expectResult: newPost(meta{
				key:     "hello_world",
				title:   "hello world",
				date:    parseTime("2012-12-01"),
				content: "<h1>title hello world</h1>\n",
				tags:    []string{"tag1", "tag2"},
			}),
		},
		"noTag": {
			input: "hello world | 2012-12-01 | \n ",
			expectResult: newPost(meta{
				key:   "hello_world",
				title: "hello world",
				date:  parseTime("2012-12-01"),
			}),
		},
		"noContent": {
			input:     "hello world | 2012-12-01 | tag1",
			expectErr: errors.New("generateAll: there must be at least one line"),
		},
		"noTitle": {
			input:     "hello world & 2012-12-01\n",
			expectErr: errors.New("generateAll: can't find title, date and tags"),
		},
		"noTime": {
			input:     "hello world || 2012-12-01\n",
			expectErr: errors.New("parsing time"),
		},
		"oneImageLink": {
			input: "hello world | 2012-12-01 | \n![1](/1/1.png)\n",
			expectResult: newPost(meta{
				key:        "hello_world",
				title:      "hello world",
				date:       parseTime("2012-12-01"),
				content:    "<p><img src=\"/images/hello_world//1/1.png\" alt=\"1\" /></p>\n",
				staticList: []string{"/images/hello_world//1/1.png"},
			}),
		},
		"twoImageLinks": {
			input: "hello world | 2012-12-01 |tag1\n![1](1/1.png)\n![2](http://2/2.png)\n",
			expectResult: newPost(meta{
				key:        "hello_world",
				title:      "hello world",
				date:       parseTime("2012-12-01"),
				tags:       []string{"tag1"},
				content:    "<p><img src=\"/images/hello_world/1/1.png\" alt=\"1\" />\n<img src=\"http://2/2.png\" alt=\"2\" /></p>\n",
				staticList: []string{"/images/hello_world/1/1.png"},
			}),
		},
	} {
		c := c
		t.Run(name, func(t *testing.T) {
			got, err := markdownGenerator{}.Generate(strings.NewReader(c.input), nil)
			if err = matchError(c.expectErr, err); err != nil {
				t.Error(err)
			}
			if !isPosterEqual(got, c.expectResult) {
				t.Errorf("\n\tgot result: %#v,\n\tbut want %#v\n", got, c.expectResult)
			}
		})
	}
}

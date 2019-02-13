package storage

import (
	"os"
	"testing"
)

func TestArticleMatch(t *testing.T) {
	for name, c := range map[string]struct {
		path   string
		expect bool
	}{
		"match": {
			path:   "a/b/c.article",
			expect: true,
		},
		"unmatch": {
			path:   "d/e/f.art",
			expect: false,
		},
	} {
		c := c
		t.Run(name, func(t *testing.T) {
			if got := articleGenerator.Match(c.path); got != c.expect {
				t.Errorf("got %v, but want %v\n", got, c.expect)
			}
		})
	}
}

func TestArticleGenerate(t *testing.T) {
	for name, c := range map[string]struct {
		inputfilePath string
		expectErr     error
		expectResult  Poster
	}{
		"normal": {
			inputfilePath: "./testdata/localRepo/1.article",
			expectResult: newPost(meta{
				key:        "Title",
				title:      "Title",
				date:       parseTime("2006-01-02"),
				content:    "<h2>Subtitle</h2><p>Some Text</p><h4 id=\"TOC_1.1.\">Subsection</h4><ul><li>bullets</li><li>more bullets</li><li>a bullet with</li></ul><h4 id=\"TOC_1.1.1.\">Sub-subsection</h4><p>Some More text</p><div class=\"code\"><pre>Preformatted text\nis indented (however you like)</pre></div><p>Further Text, including invocations like:</p><div class=\"code\">\n\n\n<pre><span num=\"7\">func main() {</span>\n<span num=\"8\">    fmt.Println(&#34;hello tw&#34;)</span>\n<span num=\"9\">}</span>\n</pre>\n\n\n</div><div class=\"playground\">\n\n\n<pre><span num=\"1\">package main</span>\n<span num=\"2\"></span>\n<span num=\"3\">import (</span>\n<span num=\"4\">    &#34;fmt&#34;</span>\n<span num=\"5\">)</span>\n<span num=\"6\"></span>\n<span num=\"7\">func main() {</span>\n<span num=\"8\">    fmt.Println(&#34;hello tw&#34;)</span>\n<span num=\"9\">}</span>\n</pre>\n\n\n</div><div class=\"image\">\n<img src=\"/images/Title/image.jpg\">\n</div><div class=\"image\">\n<img src=\"http://foo/image.jpg\">\n</div><div class=\"iframe\">\n<iframe src=\"http://foo\"frameborder=\"0\" allowfullscreen mozallowfullscreen webkitallowfullscreen></iframe>\n</div><p class=\"link\"><a href=\"http://foo\" target=\"_blank\">label</a></p><html><head>test</head><body><h1>hello tw</h1></body></html>\n<p>Again, more text</p>",
				tags:       []string{"foo", "bar", "baz"},
				staticList: []string{"/images/Title/image.jpg"},
			}),
		},
	} {
		c := c
		t.Run(name, func(t *testing.T) {
			f, err := os.Open(c.inputfilePath)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			got, err := articleGenerator.Generate(f, ts)
			if err = matchError(c.expectErr, err); err != nil {
				t.Fatal(err)
			}
			if !isPosterEqual(got, c.expectResult) {
				t.Errorf("result mismatch:\ngot: %#v\nwant:%#v\n", got, c.expectResult)
			}
		})
	}
}

func TestSlideMatch(t *testing.T) {
	for name, c := range map[string]struct {
		path   string
		expect bool
	}{
		"match": {
			path:   "a/b/c.slide",
			expect: true,
		},
		"unmatch": {
			path:   "d/e/f.sli",
			expect: false,
		},
	} {
		c := c
		t.Run(name, func(t *testing.T) {
			if got := slideGenerator.Match(c.path); got != c.expect {
				t.Errorf("got %v, but want %v\n", got, c.expect)
			}
		})
	}
}
func TestSlideGenerate(t *testing.T) {
	for name, c := range map[string]struct {
		inputfilePath string
		expectErr     error
		expectResult  Poster
	}{
		"normal": {
			inputfilePath: "./testdata/localRepo/1.slide",
			expectResult: newPost(meta{
				key:        "Title",
				title:      "Title",
				date:       parseTime("2006-01-02"),
				content:    "<section class='slides layout-widescreen'>\n<article>\n<h1>Title</h1><h3>Subtitle</h3><h3>2 January 2006</h3><div class=\"presenter\"><p>Author Name</p><p>Job title, Company</p></div></article>\n<article><h3>Title of slide or section (must have asterisk)</h3><p>Some Text</p><h2id=\"TOC_1.1.\">1.1.Subsection</h2><ul><li>bullets</li><li>more bullets</li><li>a bullet with</li></ul><h3id=\"TOC_1.1.1.\">1.1.1.Sub-subsection</h3><p>Some More text</p><div class=\"code\"><pre>Preformatted text\nis indented (however you like)</pre></div><p>Further Text, including invocations like:</p><div class=\"code\" contenteditable=\"true\" spellcheck=\"false\">\n\n\n<pre><span num=\"7\">func main() {</span>\n<span num=\"8\">    fmt.Println(&#34;hello tw&#34;)</span>\n<span num=\"9\">}</span>\n</pre>\n\n\n</div><div class=\"codeplayground\" contenteditable=\"true\" spellcheck=\"false\">\n\n\n<pre><span num=\"1\">package main</span>\n<span num=\"2\"></span>\n<span num=\"3\">import (</span>\n<span num=\"4\">    &#34;fmt&#34;</span>\n<span num=\"5\">)</span>\n<span num=\"6\"></span>\n<span num=\"7\">func main() {</span>\n<span num=\"8\">    fmt.Println(&#34;hello tw&#34;)</span>\n<span num=\"9\">}</span>\n</pre>\n\n\n</div><div class=\"image\">\n<img src=\"/images/Title/image.jpg\">\n</div><div class=\"image\">\n<img src=\"http://foo/image.jpg\">\n</div><iframe src=\"http://foo\"></iframe><p class=\"link\"><a href=\"http://foo\" target=\"_blank\">label</a></p><html><head>test</head><body><h1>hello tw</h1></body></html>\n<p>Again, more text</p></article>\n<article>\n<h3>Thank you</h1><div class=\"presenter\"><p>Author Name</p><p>Job title, Company</p><p class=\"link\"><a href=\"mailto:joe@example.com\" target=\"_blank\">joe@example.com</a></p><p class=\"link\"><a href=\"http://url/\" target=\"_blank\">http://url/</a></p><p class=\"link\"><a href=\"http://twitter.com/twitter_name\" target=\"_blank\">@twitter_name</a></p></div></article>",
				tags:       []string{"foo", "bar", "baz"},
				staticList: []string{"/images/Title/image.jpg"},
				isSlide:    true,
			}),
		},
	} {
		c := c
		t.Run(name, func(t *testing.T) {
			f, err := os.Open(c.inputfilePath)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			got, err := slideGenerator.Generate(f, ts)
			if err = matchError(c.expectErr, err); err != nil {
				t.Fatal(err)
			}
			if !isPosterEqual(got, c.expectResult) {
				t.Errorf("result mismatch:\ngot: %#v\nwant:%#v\n", got, c.expectResult)
			}
		})
	}
}

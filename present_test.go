package storage

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestPresentGenerate(t *testing.T) {
	type Expect struct {
		path, title, date, content string
		tags                       []string
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
			"./testdata/noexist/1.prst",
			pathNotFound,
			nil,
		},
		{
			nil,
			nil,
			filepath.Join(repoRoot, "1.prst"),
			nil,
			&Expect{
				path:  filepath.Join(repoRoot, "1.prst"),
				title: "Title",
				date:  "2006-01-02",
				content: `
  <h2>Subtitle</h2>





  <p>
    Some Text
  </p>




  <h4 id="TOC_1.1.">Subsection</h4>

  <ul>

    <li>bullets</li>

    <li>more bullets</li>

    <li>a bullet with</li>

  </ul>

  <h4 id="TOC_1.1.1.">Sub-subsection</h4>


  <p>
    Some More text
  </p>



  <div class="code"><pre>Preformatted text
is indented (however you like)</pre></div>



  <p>
    Further Text, including invocations like:
  </p>



	<div class="code">


<pre><span num="7">func main() {</span>
<span num="8">    fmt.Println(&#34;hello tw&#34;)</span>
<span num="9">}</span>
</pre>


</div>



	<div class="playground">


<pre><span num="1">package main</span>
<span num="2"></span>
<span num="3">import (</span>
<span num="4">    &#34;fmt&#34;</span>
<span num="5">)</span>
<span num="6"></span>
<span num="7">func main() {</span>
<span num="8">    fmt.Println(&#34;hello tw&#34;)</span>
<span num="9">}</span>
</pre>


</div>


<div class="image">
  <img src="/images/Title/image.jpg">
</div>

<div class="image">
  <img src="http://foo/image.jpg">
</div>

<div class="iframe">
  <iframe src="http://foo" frameborder="0" allowfullscreen mozallowfullscreen webkitallowfullscreen></iframe>
</div>
<p class="link"><a href="http://foo" target="_blank">label</a></p><html>
	<head>
		test
	</head>
	<body>
		<h1>hello tw</h1>
	</body>
</html>


  <p>
    Again, more text
  </p>
`,
				tags: []string{"foo", "bar", "baz"},
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
		real := &Expect{
			path:    lp.path,
			title:   string(lp.Title()),
			date:    lp.Date().Format(TimePattern),
			content: string(lp.Content()),
			tags:    lp.Tags(),
		}
		if real.path != c.expect.path {
			return fmt.Errorf("path not equal\n")
		}
		if real.title != c.expect.title {
			return fmt.Errorf("title not equal\n")
		}
		if real.date != c.expect.date {
			return fmt.Errorf("date not equal\n")
		}
		r := strings.Replace(real.content, " ", "", -1)
		r = strings.Replace(r, "\n", "", -1)
		e := strings.Replace(c.expect.content, " ", "", -1)
		e = strings.Replace(e, "\n", "", -1)
		if r != e {
			return fmt.Errorf("content not equal:\n\t%q\n\t%q\n", r, e)
		}
		if !reflect.DeepEqual(real.tags, c.expect.tags) {
			return fmt.Errorf("tags not equal: %v - %v\n", real.tags,
				c.expect.tags)
		}
		return nil
	}

	for i, c := range cases {
		if err := runCase(c); err != nil {
			t.Errorf("case %d error: %s\n", i, err)
		}
	}
}

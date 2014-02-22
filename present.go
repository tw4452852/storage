package storage

import (
	"bytes"
	"code.google.com/p/go.tools/present"
	"html/template"
	"io"
	"io/ioutil"
	"regexp"
)

var (
	articleTmpl *template.Template
)

const (
	articleTmplString = `
{{/* This doc template is given to the present tool to format articles.  */}}

{{define "root"}}

  {{with .Subtitle}}<h2>{{.}}</h2>{{end}}
  {{if .Doc | sectioned}}
    {{range .Sections}}
      {{elem $.Template .}}
    {{end}}
  {{else}}
    {{with index .Sections 0}}
      {{range .Elem}}
        {{elem $.Template .}}
      {{end}}
    {{end}}
  {{end}}

{{end}}

{{define "TOC"}}
  <ul>
  {{range .}}
    <li><a href="#TOC_{{.FormattedNumber}}">{{.Title}}</a></li>
    {{with .Sections}}{{template "TOC" .}}{{end}}
  {{end}}
  </ul>
{{end}}

{{define "newline"}}
{{/* No automatic line break. Paragraphs are free-form. */}}
{{end}}

{{define "section"}}
  <h4 id="TOC_{{.FormattedNumber}}">{{.Title}}</h4>
  {{range .Elem}}{{elem $.Template .}}{{end}}
{{end}}

{{define "list"}}
  <ul>
  {{range .Bullet}}
    <li>{{style .}}</li>
  {{end}}
  </ul>
{{end}}

{{define "text"}}
  {{if .Pre}}
  <div class="code"><pre>{{range .Lines}}{{.}}{{end}}</pre></div>
  {{else}}
  <p>
    {{range $i, $l := .Lines}}{{if $i}}{{template "newline"}}
    {{end}}{{style $l}}{{end}}
  </p>
  {{end}}
{{end}}

{{define "code"}}
  {{if .Play}}
	<div class="playground">{{.Text}}</div>
  {{else}}
	<div class="code">{{.Text}}</div>
  {{end}}
{{end}}

{{define "image"}}
<div class="image">
  <img src="{{.URL}}"{{with .Height}} height="{{.}}"{{end}}{{with .Width}} width="{{.}}"{{end}}>
</div>
{{end}}

{{define "iframe"}}
<div class="iframe">
  <iframe src="{{.URL}}"{{with .Height}} height="{{.}}"{{end}}{{with .Width}} width="{{.}}"{{end}} frameborder="0" allowfullscreen mozallowfullscreen webkitallowfullscreen></iframe>
</div>
{{end}}

{{define "link"}}<p class="link"><a href="{{.URL}}" target="_blank">{{style .Label}}</a></p>{{end}}

{{define "html"}}{{.HTML}}{{end}}
`
)

func init() {
	RegisterGenerator(ArticleGenerator{regexp.MustCompile(".*.article$")})

	// init presentTmpl
	funcMap := template.FuncMap{
		"sectioned": func(d *present.Doc) bool {
			return len(d.Sections) > 1
		},
	}
	var e error
	articleTmpl, e = present.Template().Funcs(funcMap).Parse(articleTmplString)
	if e != nil {
		panic(e)
	}

	// enable playgroud
	present.PlayEnabled = true
}

type ArticleGenerator presentGenerator

func (ag ArticleGenerator) Generate(input io.Reader, s Staticer) (error, *meta) {
	return presentGenerator(ag).generate(input, s, articleTmpl)
}

func (ag ArticleGenerator) Match(filename string) bool {
	return presentGenerator(ag).match(filename)
}

type presentGenerator struct {
	matcher *regexp.Regexp
}

func (p presentGenerator) match(filename string) bool {
	return p.matcher.MatchString(filename)
}

func (presentGenerator) generate(input io.Reader, s Staticer, tmpl *template.Template) (error, *meta) {
	ctx := &present.Context{func(filename string) ([]byte, error) {
		r := s.Static(filename)
		if closer, ok := r.(io.Closer); ok {
			defer closer.Close()
		}
		return ioutil.ReadAll(r)
	}}
	doc, err := ctx.Parse(input, "", 0)
	if err != nil {
		return err, nil
	}
	key := title2Key(doc.Title)
	fixImageLink(doc, key)
	// TODO: buffer pool
	b := new(bytes.Buffer)
	err = doc.Render(b, tmpl)
	if err != nil {
		return err, nil
	}
	return nil, &meta{
		title:   doc.Title,
		date:    doc.Time,
		key:     key,
		content: template.HTML(b.String()),
		tags:    doc.Tags,
	}
}

func fixImageLink(doc *present.Doc, key string) {
	var checkElem func(present.Elem) present.Elem
	checkElem = func(e present.Elem) present.Elem {
		if s, ok := e.(present.Section); ok {
			for i, e := range s.Elem {
				s.Elem[i] = checkElem(e)
			}
		}
		if image, ok := e.(present.Image); ok {
			image.URL = generateImageLink(key, image.URL)
			return image
		}
		return e
	}
	for i, s := range doc.Sections {
		doc.Sections[i] = checkElem(s).(present.Section)
	}
}

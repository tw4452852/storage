package storage

import (
	"bytes"
	"code.google.com/p/go.tools/present"
	"html/template"
	"io"
	"regexp"
	"strings"
)

var presentTmpl *template.Template

const presentTmplString = `
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

func init() {
	RegisterGenerator(presentGenerator{regexp.MustCompile(".*.prst$")})

	// init presentTmpl
	funcMap := template.FuncMap{
		"sectioned": func(d *present.Doc) bool {
			return len(d.Sections) > 1
		},
	}
	var e error
	presentTmpl, e = present.Template().Funcs(funcMap).Parse(presentTmplString)
	if e != nil {
		panic(e)
	}
}

type presentGenerator struct {
	matcher *regexp.Regexp
}

func (p presentGenerator) Match(filename string) bool {
	return p.matcher.MatchString(filename)
}

func (presentGenerator) Generate(input io.Reader) (error, *meta) {
	doc, err := present.Parse(input, "", 0)
	if err != nil {
		return err, nil
	}
	b := new(bytes.Buffer)
	err = doc.Render(b, presentTmpl)
	if err != nil {
		return err, nil
	}
	key := strings.Replace(doc.Title, " ", "_", -1)
	return nil, &meta{
		title:   doc.Title,
		date:    doc.Time,
		key:     key,
		content: template.HTML(b.String()),
	}
}

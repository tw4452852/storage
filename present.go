package storage

import (
	"bytes"
	"html/template"
	"io"
	"io/ioutil"
	"strings"

	"golang.org/x/tools/present"
)

var (
	articleTmpl *template.Template
	slideTmpl   *template.Template
)

const (
	articleTmplString = `
{{- /* This doc template is given to the present tool to format articles.  */ -}}
{{- define "root" -}}
{{- with .Subtitle -}}<h2>{{- . -}}</h2>{{- end -}}
{{- if .Doc | sectioned -}}
{{- range .Sections -}}
{{- elem $.Template . -}}
{{- end -}}
{{- else -}}
{{- with index .Sections 0 -}}
{{- range .Elem -}}
{{- elem $.Template . -}}
{{- end -}}
{{- end -}}
{{- end -}}
{{- end -}}
{{- define "TOC" -}}
<ul>
{{- range . -}}
<li><a href="#TOC_{{- .FormattedNumber -}}">{{- .Title -}}</a></li>
{{- with .Sections -}}{{- template "TOC" . -}}{{- end -}}
{{- end -}}
</ul>
{{- end -}}
{{- define "newline" -}}
{{- /* No automatic line break. Paragraphs are free-form. */ -}}
{{- end -}}
{{- define "section" -}}
<h4 id="TOC_{{- .FormattedNumber -}}">{{- .Title -}}</h4>
{{- range .Elem -}}{{- elem $.Template . -}}{{- end -}}
{{- end -}}
{{- define "list" -}}
<ul>
{{- range .Bullet -}}
<li>{{- style . -}}</li>
{{- end -}}
</ul>
{{- end -}}
{{- define "text" -}}
{{- if .Pre -}}
<div class="code"><pre>{{- range .Lines -}}{{- . -}}{{- end -}}</pre></div>
{{- else -}}
<p>
{{- range $i, $l := .Lines -}}{{- if $i -}}{{- template "newline" -}}
{{- end -}}{{- style $l -}}{{- end -}}
</p>
{{- end -}}
{{- end -}}
{{- define "code" -}}
{{- if .Play -}}
<div class="playground">{{- .Text -}}</div>
{{- else -}}
<div class="code">{{- .Text -}}</div>
{{- end -}}
{{- end -}}
{{- define "image" -}}
<div class="image">
<img src="{{- .URL -}}"{{- with .Height -}} height="{{- . -}}"{{- end -}}{{- with .Width -}} width="{{- . -}}"{{- end -}}>
</div>
{{- end -}}
{{- define "iframe" -}}
<div class="iframe">
<iframe src="{{- .URL -}}"{{- with .Height -}} height="{{- . -}}"{{- end -}}{{- with .Width -}} width="{{- . -}}"{{- end -}} frameborder="0" allowfullscreen mozallowfullscreen webkitallowfullscreen></iframe>
</div>
{{- end -}}
{{- define "link" -}}<p class="link"><a href="{{- .URL -}}" target="_blank">{{- style .Label -}}</a></p>{{- end -}}
{{- define "html" -}}{{- .HTML -}}{{- end -}}
	`

	slideTmplString = `
{{- define "root" -}}
<section class='slides layout-widescreen'>
<article>
<h1>{{- .Title -}}</h1>
{{- with .Subtitle -}}<h3>{{- . -}}</h3>{{- end -}}
{{- if not .Time.IsZero -}}<h3>{{- .Time.Format "2 January 2006" -}}</h3>{{- end -}}
{{- range .Authors -}}
<div class="presenter">
{{- range .TextElem -}}{{- elem $.Template . -}}{{- end -}}
</div>
{{- end -}}
</article>
{{- range $i, $s := .Sections -}}
<!-- start of slide {{- $s.Number -}} -->
<article>
{{- if $s.Elem -}}
<h3>{{- $s.Title -}}</h3>
{{- range $s.Elem -}}{{- elem $.Template . -}}{{- end -}}
{{- else -}}
<h2>{{- $s.Title -}}</h2>
{{- end -}}
</article>
<!-- end of slide {{- $i -}} -->
{{- end -}}{{- /* of Slide block */ -}}
<article>
<h3>Thank you</h1>
{{- range .Authors -}}
<div class="presenter">
{{- range .Elem -}}{{- elem $.Template . -}}{{- end -}}
</div>
{{- end -}}
</article>
{{- end -}}
{{- define "newline" -}}
<br>
{{- end -}}
{{- define "section" -}}
<h{{- len .Number -}} id="TOC_{{- .FormattedNumber -}}">{{- .FormattedNumber -}} {{- .Title -}}</h{{- len .Number -}}>
{{- range .Elem -}}{{- elem $.Template . -}}{{- end -}}
{{- end -}}
{{- define "list" -}}
<ul>
{{- range .Bullet -}}
<li>{{- style . -}}</li>
{{- end -}}
</ul>
{{- end -}}
{{- define "text" -}}
{{- if .Pre -}}
<div class="code"><pre>{{- range .Lines -}}{{- . -}}{{- end -}}</pre></div>
{{- else -}}
<p>
{{- range $i, $l := .Lines -}}{{- if $i -}}{{- template "newline" -}}
{{- end -}}{{- style $l -}}{{- end -}}
</p>
{{- end -}}
{{- end -}}
{{- define "code" -}}
<div class="code{{- if .Play -}} playground{{- end -}}" contenteditable="true" spellcheck="false">{{- .Text -}}</div>
{{- end -}}
{{- define "image" -}}
<div class="image">
<img src="{{- .URL -}}"{{- with .Height -}} height="{{- . -}}"{{- end -}}{{- with .Width -}} width="{{- . -}}"{{- end -}}>
</div>
{{- end -}}
{{- define "iframe" -}}
<iframe src="{{- .URL -}}"{{- with .Height -}} height="{{- . -}}"{{- end -}}{{- with .Width -}} width="{{- . -}}"{{- end -}}></iframe>
{{- end -}}
{{- define "link" -}}<p class="link"><a href="{{- .URL -}}" target="_blank">{{- style .Label -}}</a></p>{{- end -}}
{{- define "html" -}}{{- .HTML -}}{{- end -}}
	`
)

var (
	articleGenerator = presentGenerator{
		match: func(filename string) bool {
			return strings.HasSuffix(filename, ".article")
		},
	}
	slideGenerator = presentGenerator{
		match: func(filename string) bool {
			return strings.HasSuffix(filename, ".slide")
		},
	}
)

func init() {
	funcMap := template.FuncMap{
		"sectioned": func(d *present.Doc) bool {
			return len(d.Sections) > 1
		},
	}

	// init articleTmpl and slideTmpl
	var e error
	articleTmpl, e = present.Template().Funcs(funcMap).Parse(articleTmplString)
	if e != nil {
		panic(e)
	}
	slideTmpl, e = present.Template().Funcs(funcMap).Parse(slideTmplString)
	if e != nil {
		panic(e)
	}

	// enable playgroud
	present.PlayEnabled = true

	articleGenerator.tmpl = articleTmpl
	slideGenerator.tmpl = slideTmpl

	RegisterGenerator(articleGenerator)
	RegisterGenerator(slideGenerator)
}

type presentGenerator struct {
	match func(string) bool
	tmpl  *template.Template
}

func (p presentGenerator) Match(filename string) bool {
	return p.match(filename)
}

func (p presentGenerator) Generate(input io.Reader, s Staticer) (Poster, error) {
	return p.generate(input, s, p.tmpl)
}

func (presentGenerator) generate(input io.Reader, s Staticer, tmpl *template.Template) (Poster, error) {
	ctx := &present.Context{func(filename string) ([]byte, error) {
		r := s.Static(filename)
		defer r.Close()
		return ioutil.ReadAll(r)
	}}
	doc, err := ctx.Parse(input, "", 0)
	if err != nil {
		return nil, err
	}
	key := title2Key(doc.Title)

	images := fixImageLink(doc, key)

	// TODO: buffer pool
	b := new(bytes.Buffer)
	err = doc.Render(b, tmpl)
	if err != nil {
		return nil, err
	}

	return newPost(meta{
		title:      doc.Title,
		date:       doc.Time,
		key:        key,
		content:    bytes2String(b.Bytes()),
		tags:       doc.Tags,
		isSlide:    tmpl == slideTmpl,
		staticList: images,
	}), nil
}

func fixImageLink(doc *present.Doc, key string) (images []string) {
	var checkElem func(present.Elem) present.Elem
	checkElem = func(e present.Elem) present.Elem {
		if s, ok := e.(present.Section); ok {
			for i, e := range s.Elem {
				s.Elem[i] = checkElem(e)
			}
		}
		if image, ok := e.(present.Image); ok {
			if needChangeImageLink(image.URL) {
				image.URL = generateImageLink(key, image.URL)
				images = append(images, image.URL)
				return image
			}
		}
		return e
	}
	for i, s := range doc.Sections {
		doc.Sections[i] = checkElem(s).(present.Section)
	}
	return
}

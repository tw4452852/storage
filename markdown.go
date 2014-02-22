package storage

import (
	"bytes"
	"errors"
	"github.com/russross/blackfriday"
	"html/template"
	"io"
	"io/ioutil"
	"regexp"
	"strings"
	"time"
)

const (
	htmlFlags = blackfriday.HTML_USE_XHTML |
		blackfriday.HTML_USE_SMARTYPANTS |
		blackfriday.HTML_SMARTYPANTS_FRACTIONS |
		blackfriday.HTML_SMARTYPANTS_LATEX_DASHES
	extensions = blackfriday.EXTENSION_NO_INTRA_EMPHASIS |
		blackfriday.EXTENSION_TABLES |
		blackfriday.EXTENSION_FENCED_CODE |
		blackfriday.EXTENSION_AUTOLINK |
		blackfriday.EXTENSION_STRIKETHROUGH |
		blackfriday.EXTENSION_SPACE_HEADERS
	seperator = "|"
)

func init() {
	RegisterGenerator(MarkdownGenerator{regexp.MustCompile(".*.md$")})
}

type MarkdownGenerator struct {
	matcher *regexp.Regexp
}

func (m MarkdownGenerator) Match(filename string) bool {
	return m.matcher.MatchString(filename)
}

func (MarkdownGenerator) Generate(input io.Reader, static Staticer) (error, *meta) {
	c, e := ioutil.ReadAll(input)
	if e != nil {
		return e, nil
	}
	//title
	firstLineIndex := strings.Index(string(c), "\n")
	if firstLineIndex == -1 {
		return errors.New("generateAll: there must be at least one line\n"), nil
	}
	firstLine := strings.TrimSpace(string(c[:firstLineIndex]))
	titleDateTags := strings.Split(firstLine, seperator)
	if len(titleDateTags) != 3 {
		return errors.New("generateAll: can't find title, date and tags\n"), nil
	}
	title := strings.TrimSpace(titleDateTags[0])
	//date
	t, e := time.Parse(TimePattern, strings.TrimSpace(titleDateTags[1]))
	if e != nil {
		return e, nil
	}
	//key
	key := title2Key(title)
	//tags
	tags := []string{}
	tagsString := strings.TrimSpace(titleDateTags[2])
	if tagsString != "" {
		tags = strings.Split(tagsString, ",")
		for i, tag := range tags {
			tags[i] = strings.TrimSpace(tag)
		}
	}
	//content
	remain := strings.TrimSpace(string(c[firstLineIndex+1:]))
	content := template.HTML(markdown([]byte(remain), key))

	return nil, &meta{
		key:     key,
		title:   title,
		date:    t,
		content: content,
		tags:    tags,
	}
}

type myRender struct {
	key string //myself post key
	blackfriday.Renderer
}

func (mr *myRender) BlockCode(out *bytes.Buffer, text []byte, lang string) {
	out.WriteString(`<div class="code">`)
	mr.Renderer.BlockCode(out, text, lang)
	out.WriteString(`</div>`)
}

//add prefix to img link
func (mr *myRender) Image(out *bytes.Buffer, link, title, alt []byte) {
	mr.Renderer.Image(out, []byte(generateImageLink(mr.key, string(link))), title, alt)
}

func markdown(input []byte, key string) []byte { /*{{{*/
	// set up the HTML renderer
	renderer := &myRender{
		key:      key,
		Renderer: blackfriday.HtmlRenderer(htmlFlags, "", ""),
	}
	return blackfriday.Markdown(input, renderer, extensions)
} /*}}}*/

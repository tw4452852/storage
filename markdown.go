package storage

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/russross/blackfriday"
	"html/template"
	"io"
	"io/ioutil"
	"regexp"
	"strings"
	"time"
)

const (
	imagePrefix = "/images/" //add this prefix to the origin image link
	htmlFlags   = blackfriday.HTML_USE_XHTML |
		blackfriday.HTML_USE_SMARTYPANTS |
		blackfriday.HTML_SMARTYPANTS_FRACTIONS |
		blackfriday.HTML_SMARTYPANTS_LATEX_DASHES
	extensions = blackfriday.EXTENSION_NO_INTRA_EMPHASIS |
		blackfriday.EXTENSION_TABLES |
		blackfriday.EXTENSION_FENCED_CODE |
		blackfriday.EXTENSION_AUTOLINK |
		blackfriday.EXTENSION_STRIKETHROUGH |
		blackfriday.EXTENSION_SPACE_HEADERS
)

func init() {
	RegisterGenerator(markdownGenerator{regexp.MustCompile(".*.md$")})
}

type markdownGenerator struct {
	matcher *regexp.Regexp
}

func (m markdownGenerator) Match(filename string) bool {
	return m.matcher.MatchString(filename)
}

func (markdownGenerator) Generate(input io.Reader) (error, *meta) {
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
	sepIndex := strings.Index(firstLine, TitleAndDateSeperator)
	if sepIndex == -1 {
		return errors.New("generateAll: can't find seperator for title and date\n"), nil
	}
	title := strings.TrimSpace(firstLine[:sepIndex])

	//date
	t, e := time.Parse(TimePattern, strings.TrimSpace(firstLine[sepIndex+1:]))
	if e != nil {
		return e, nil
	}

	//key
	h := md5.New()
	io.WriteString(h, firstLine)
	key := fmt.Sprintf("%x", h.Sum(nil))

	//content
	remain := strings.TrimSpace(string(c[firstLineIndex+1:]))
	content := template.HTML(markdown([]byte(remain), key))

	return nil, &meta{
		key:     key,
		title:   title,
		date:    t,
		content: content,
	}
}

type myRender struct {
	key string //myself post key
	blackfriday.Renderer
}

//add prefix to img link
func (mr *myRender) Image(out *bytes.Buffer, link, title, alt []byte) {
	if wantChange(link) {
		mr.Renderer.Image(out, []byte(imagePrefix+mr.key+"/"+string(link)), title, alt)
	} else {
		mr.Renderer.Image(out, link, title, alt)
	}
}

//wantChange check whether the image's link need to add prefix
func wantChange(link []byte) bool {
	if bytes.HasPrefix(link, []byte("http://")) ||
		bytes.HasPrefix(link, []byte("https://")) {
		return false
	}
	return true
}

func markdown(input []byte, key string) []byte { /*{{{*/
	// set up the HTML renderer
	renderer := &myRender{
		key:      key,
		Renderer: blackfriday.HtmlRenderer(htmlFlags, "", ""),
	}

	return blackfriday.Markdown(input, renderer, extensions)
} /*}}}*/

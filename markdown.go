package storage

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/russross/blackfriday"
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
	seperator   = "|"
	timePattern = "2006-01-02"
)

func init() {
	RegisterGenerator(markdownGenerator{})
}

type markdownGenerator struct{}

func (m markdownGenerator) Match(filename string) bool {
	return strings.HasSuffix(filename, ".md")
}

func (markdownGenerator) Generate(input io.Reader, _ Staticer) (Poster, error) {
	c, e := ioutil.ReadAll(input)
	if e != nil {
		return nil, e
	}
	// title
	firstLineIndex := strings.Index(string(c), "\n")
	if firstLineIndex == -1 {
		return nil, errors.New("generateAll: there must be at least one line\n")
	}
	firstLine := strings.TrimSpace(string(c[:firstLineIndex]))
	titleDateTags := strings.Split(firstLine, seperator)
	if len(titleDateTags) != 3 {
		return nil, errors.New("generateAll: can't find title, date and tags\n")
	}
	title := strings.TrimSpace(titleDateTags[0])
	// date
	t, e := time.Parse(timePattern, strings.TrimSpace(titleDateTags[1]))
	if e != nil {
		return nil, e
	}
	// key
	key := title2Key(title)
	// tags
	var tags []string
	tagsString := strings.TrimSpace(titleDateTags[2])
	if tagsString != "" {
		tags = strings.Split(tagsString, ",")
		for i, tag := range tags {
			tags[i] = strings.TrimSpace(tag)
		}
	}
	// content
	remain := bytes.TrimSpace(c[firstLineIndex+1:])
	renderer := &myRender{
		key:      key,
		Renderer: blackfriday.HtmlRenderer(htmlFlags, "", ""),
	}
	content := blackfriday.Markdown(remain, renderer, extensions)

	return newPost(meta{
		key:        key,
		title:      title,
		date:       t,
		content:    bytes2String(content),
		tags:       tags,
		isSlide:    false,
		staticList: renderer.images,
	}), nil
}

type myRender struct {
	images []string // collect image links
	key    string   // myself post key
	blackfriday.Renderer
}

func (mr *myRender) BlockCode(out *bytes.Buffer, text []byte, lang string) {
	out.WriteString(`<div class="code">`)
	mr.Renderer.BlockCode(out, text, lang)
	out.WriteString(`</div>`)
}

// add prefix to img link
func (mr *myRender) Image(out *bytes.Buffer, link, title, alt []byte) {
	if slink := string(link); needChangeImageLink(slink) {
		imageLink := generateImageLink(mr.key, slink)
		link = []byte(imageLink)
		mr.images = append(mr.images, imageLink)
	}
	mr.Renderer.Image(out, link, title, alt)
}

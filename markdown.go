package storage

import (
	"bytes"
	"github.com/russross/blackfriday"
)

//a wrapper of blackfriday.Renderer
type myRender struct {
	imgPrefix string
	blackfriday.Renderer
}

//add prefix to img link
func (mr *myRender) Image(out *bytes.Buffer, link, title, alt []byte) {
	mr.Renderer.Image(out, []byte(mr.imgPrefix+string(link)), title, alt)
}

func markdown(input []byte, imgPrefix string) []byte { /*{{{*/
	// set up the HTML renderer
	htmlFlags := 0
	htmlFlags |= blackfriday.HTML_USE_XHTML
	htmlFlags |= blackfriday.HTML_USE_SMARTYPANTS
	htmlFlags |= blackfriday.HTML_SMARTYPANTS_FRACTIONS
	htmlFlags |= blackfriday.HTML_SMARTYPANTS_LATEX_DASHES
	renderer := &myRender{
		imgPrefix: imgPrefix,
		Renderer:  blackfriday.HtmlRenderer(htmlFlags, "", ""),
	}

	// set up the parser
	extensions := 0
	extensions |= blackfriday.EXTENSION_NO_INTRA_EMPHASIS
	extensions |= blackfriday.EXTENSION_TABLES
	extensions |= blackfriday.EXTENSION_FENCED_CODE
	extensions |= blackfriday.EXTENSION_AUTOLINK
	extensions |= blackfriday.EXTENSION_STRIKETHROUGH
	extensions |= blackfriday.EXTENSION_SPACE_HEADERS

	return blackfriday.Markdown(input, renderer, extensions)
} /*}}}*/

package markdown

import (
	"bytes"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

type Renderer struct {
	engine goldmark.Markdown
	policy *bluemonday.Policy
}

func New() *Renderer {
	return &Renderer{
		engine: goldmark.New(
			goldmark.WithExtensions(extension.GFM, extension.Linkify, extension.Table),
			goldmark.WithParserOptions(parser.WithAutoHeadingID()),
			goldmark.WithRendererOptions(html.WithUnsafe()),
		),
		policy: bluemonday.UGCPolicy(),
	}
}

func (r *Renderer) Render(text string) (string, error) {
	var buf bytes.Buffer
	if err := r.engine.Convert([]byte(text), &buf); err != nil {
		return "", err
	}
	return r.policy.Sanitize(buf.String()), nil
}

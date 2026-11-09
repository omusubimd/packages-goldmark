package ruby

import (
	"io"

	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/renderer"
	"github.com/yuin/goldmark/v2/renderer/html"
)

type rubyHTMLRendererExtension struct{}

// NewHTMLRenderer returns a Goldmark HTML extension for ruby nodes.
func NewHTMLRenderer() html.Extension { return &rubyHTMLRendererExtension{} }

func (*rubyHTMLRendererExtension) RendererOptions(*html.Config) []html.Option {
	return []html.Option{html.WithNodeRenderers(map[ast.NodeKind]html.NodeRenderer{
		KindRuby:        html.NodeRendererFunc(renderRuby),
		KindRubySegment: html.NodeRendererFunc(renderSegment),
	})}
}

func renderRuby(w io.Writer, _ []byte, _ ast.Node, entering bool, _ renderer.Context) (ast.WalkStatus, error) {
	tag := "</ruby>"
	if entering {
		tag = "<ruby>"
	}
	_, err := io.WriteString(w, tag)
	return ast.WalkContinue, err
}

func renderSegment(w io.Writer, source []byte, n ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	if entering {
		return ast.WalkContinue, nil
	}
	if _, err := io.WriteString(w, "<rt>"); err != nil {
		return ast.WalkStop, err
	}
	if _, err := n.(*RubySegmentNode).RubyText.WriteTo(html.ContextTextWriter(rc), source); err != nil {
		return ast.WalkStop, err
	}
	_, err := io.WriteString(w, "</rt>")
	return ast.WalkContinue, err
}

// HTMLRenderer is the default ruby HTML renderer extension.
var HTMLRenderer = NewHTMLRenderer()

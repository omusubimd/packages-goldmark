package internal

import (
	"bytes"

	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/text"
	"github.com/yuin/goldmark/v2/util"
)

// Labels are inline fragments: heading, list and HTML block syntax must not
// change their interpretation. Each segment gets its own delimiter context.
func newLabelParser() parser.Parser {
	return parser.New(
		parser.WithDefaultParsers(false),
		parser.WithBlockParsers(util.Prioritized[parser.BlockParser](&labelParagraphParser{}, 100)),
		parser.WithInlineParsers(
			util.Prioritized(parser.NewCodeSpanParser(), 100),
			util.Prioritized(parser.NewLinkParser(), 200),
			util.Prioritized(parser.NewAutoLinkParser(), 300),
			util.Prioritized(parser.NewRawHTMLParser(), 400),
			util.Prioritized(parser.NewEmphasisParser(), 500),
		),
	)
}

type labelParagraphParser struct{}

func (*labelParagraphParser) Trigger() []byte { return nil }
func (*labelParagraphParser) Open(_ ast.Node, r text.Reader, _ parser.Context) (ast.Node, parser.State) {
	_, segment := r.PeekLine()
	n := ast.NewParagraph()
	n.AppendSource(segment)
	r.AdvanceToEOL()
	return n, parser.NoChildren
}
func (*labelParagraphParser) Continue(ast.Node, text.Reader, parser.Context) parser.State {
	return parser.Close
}
func (*labelParagraphParser) Close(ast.Node, text.Reader, parser.Context) {}
func (*labelParagraphParser) CanInterruptParagraph() bool                 { return false }
func (*labelParagraphParser) CanAcceptIndentedLine() bool                 { return true }

func appendLabel(parent ast.Node, p parser.Parser, source []byte, offset int, outer parser.Context) {
	pc := parser.NewContext()
	for _, definition := range outer.LinkDefinitions() {
		pc.AddLinkDefinition(definition)
	}
	doc := p.Parse(source, parser.WithContext(pc))
	paragraph := doc.FirstChild()
	if paragraph == nil {
		parent.AppendChild(ast.NewText(text.NewSingleLineValue(source, text.IdentityDecoder)))
		return
	}
	// Fragment-backed values cannot refer to the document's byte slice. Own the
	// decoded values before moving nodes, so every renderer/AST consumer can use
	// the original document source safely.
	_ = ast.Walk(paragraph, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if pos := n.Pos(); pos >= 0 {
			n.(interface{ SetPos(int) }).SetPos(offset + pos)
		}
		single := func(v text.SingleLineValue) text.SingleLineValue {
			return text.NewSingleLineValue(v.Value(source), text.IdentityDecoder)
		}
		multi := func(v text.MultiLineValue) text.MultiLineValue {
			return text.NewMultiLineValue(v.Value(source), text.IdentityDecoder)
		}
		switch v := n.(type) {
		case *ast.Text:
			v.Value = single(v.Value)
		case *ast.CodeSpan:
			v.Value = text.NewSingleLineValue(v.Value.Value(source), text.IdentityDecoder)
		case *ast.Link:
			v.Destination = single(v.Destination)
			v.Title = multi(v.Title)
		case *ast.Image:
			v.Destination = single(v.Destination)
			v.Title = multi(v.Title)
		case *ast.AutoLink:
			v.Destination = single(v.Destination)
			v.Label = single(v.Label)
			v.Text = single(v.Text)
		case *ast.RawHTML:
			v.Value = multi(v.Value)
		}
		return ast.WalkContinue, nil
	})
	for paragraph.FirstChild() != nil {
		child := paragraph.FirstChild()
		paragraph.RemoveChild(child)
		parent.AppendChild(child)
	}
	// Goldmark trims paragraph-final whitespace; ruby labels preserve it.
	trimmed := bytes.TrimRight(source, " \t")
	if len(trimmed) < len(source) {
		parent.AppendChild(ast.NewText(text.NewSingleLineValue(source[len(trimmed):], text.IdentityDecoder)))
	}
}

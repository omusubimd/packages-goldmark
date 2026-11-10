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
func newLabelParser(config *parser.Config) parser.Parser {
	var emphasisOptions []parser.EmphasisOption
	if config.ParseDelimiterFunc != nil {
		emphasisOptions = append(emphasisOptions, parser.WithParseDelimiterFunc(config.ParseDelimiterFunc))
	}
	var options []parser.Option
	if config.EscapedSpace {
		options = append(options, parser.WithEscapedSpace())
	}
	options = append(options,
		parser.WithDefaultParsers(false),
		parser.WithBlockParsers(util.Prioritized[parser.BlockParser](&labelParagraphParser{}, 100)),
		parser.WithInlineParsers(
			util.Prioritized(parser.NewCodeSpanParser(), 100),
			util.Prioritized(parser.NewLinkParser(), 200),
			util.Prioritized(parser.NewAutoLinkParser(), 300),
			util.Prioritized(parser.NewRawHTMLParser(), 400),
			util.Prioritized(parser.NewEmphasisParser(emphasisOptions...), 500),
		),
	)
	return parser.New(options...)
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

func appendLabel(parent ast.Node, p parser.Parser, source []byte, offset int, outer parser.Context, decoder text.Decoder) {
	pc := parser.NewContext()
	for _, definition := range outer.LinkDefinitions() {
		pc.AddLinkDefinition(definition)
	}
	doc := p.Parse(source, parser.WithContext(pc))
	paragraph := doc.FirstChild()
	if paragraph == nil {
		parent.AppendChild(ast.NewText(text.NewSingleLineValue(source, decoder)))
		return
	}
	// Goldmark trims paragraph-final whitespace. Restore the raw suffix before
	// decoding: an escaped final space must be handled together with its slash,
	// while a literal escaped backslash followed by a space must stay literal.
	trimmed := bytes.TrimRight(source, " \t")
	if len(trimmed) < len(source) {
		suffix := source[len(trimmed):]
		if last, ok := paragraph.LastChild().(*ast.Text); ok {
			last.Value = text.NewSingleLineValue(last.Value.Str(source)+string(suffix), decoder)
		} else {
			paragraph.AppendChild(ast.NewText(text.NewSingleLineValue(suffix, decoder)))
		}
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
}

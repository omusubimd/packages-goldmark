// Package indent adds Omusubi's ideographic-space paragraph syntax to Goldmark v2.
package indent

import (
	"bytes"

	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/text"
	"github.com/yuin/goldmark/v2/util"
)

const marker = "\u3000"

type indentExtension struct{}

// NewParser returns an extension that starts a paragraph at a U+3000 marker.
// It emits ordinary Goldmark nodes and needs no renderer extension.
func NewParser() parser.Extension { return &indentExtension{} }

// Parser is the default indent parser extension.
var Parser = NewParser()

func (*indentExtension) ParserOptions(*parser.Config) []parser.Option {
	return []parser.Option{
		parser.WithBlockParsers(util.Prioritized[parser.BlockParser](&indentParser{paragraph: parser.NewParagraphParser()}, 950)),
		parser.WithASTTransformers(util.Prioritized[parser.ASTTransformer](&removeSeparators{}, 100)),
	}
}

type indentParser struct{ paragraph parser.BlockParser }

// Goldmark dispatches by the first UTF-8 byte; Open checks the entire marker.
func (*indentParser) Trigger() []byte             { return []byte{0xe3} }
func (*indentParser) CanInterruptParagraph() bool { return true }
func (*indentParser) CanAcceptIndentedLine() bool { return false }

func (*indentParser) Open(_ ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, _ := reader.PeekLine()
	offset := pc.BlockOffset()
	if offset < 0 || offset > len(line) || !bytes.HasPrefix(line[offset:], []byte(marker)) {
		return nil, parser.NoChildren
	}
	reader.Advance(offset + len(marker))
	content, segment := reader.PeekLine()
	if util.IsBlank(content) {
		reader.AdvanceToEOL()
		return newSeparator(), parser.NoChildren
	}
	// Remove exactly the marker and the preceding Markdown indentation. Spaces
	// after it belong to the paragraph, not a new block indentation level.
	paragraph := ast.NewParagraph()
	paragraph.AppendSource(segment)
	reader.AdvanceToEOL()
	return paragraph, parser.NoChildren
}

func (p *indentParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	if _, ok := node.(*separator); ok {
		return parser.Close
	}
	return p.paragraph.Continue(node, reader, pc)
}
func (p *indentParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	if _, ok := node.(*ast.Paragraph); ok {
		p.paragraph.Close(node, reader, pc)
	}
}

// A marker-only line separates blocks without inventing an empty paragraph.
// Keeping a temporary non-paragraph node prevents a following setext underline
// from reaching across the separator to the preceding paragraph.
type separator struct{ ast.BaseBlock }

var kindSeparator = ast.NewNodeKind("OmusubiIndentSeparator")

func newSeparator() *separator                 { n := &separator{}; n.Init(n); return n }
func (*separator) Kind() ast.NodeKind          { return kindSeparator }
func (n *separator) Dump([]byte) *ast.NodeDump { return ast.NewNodeDump(n, nil) }

type removeSeparators struct{}

func (*removeSeparators) Transform(doc *ast.Document, _ text.Reader, _ parser.Context) {
	var nodes []*separator
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if node, ok := n.(*separator); ok && entering {
			nodes = append(nodes, node)
		}
		return ast.WalkContinue, nil
	})
	for _, node := range nodes {
		node.Parent().RemoveChild(node)
	}
}

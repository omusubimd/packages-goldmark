package internal

import (
	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/text"
)

// RubyTransformer moves segments without a reading outside ruby nodes and
// groups each consecutive run of annotated segments into one ruby node.
type RubyTransformer struct{}

func (*RubyTransformer) Transform(doc *ast.Document, _ text.Reader, _ parser.Context) {
	var rubies []*RubyNode
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if ruby, ok := n.(*RubyNode); ok && entering {
			rubies = append(rubies, ruby)
		}
		return ast.WalkContinue, nil
	})
	for _, ruby := range rubies {
		parent := ruby.Parent()
		var run *RubyNode
		for ruby.FirstChild() != nil {
			segment := ruby.FirstChild().(*RubySegmentNode)
			ruby.RemoveChild(segment)
			if segment.RubyText.IsEmpty() {
				run = nil
				for segment.FirstChild() != nil {
					child := segment.FirstChild()
					segment.RemoveChild(child)
					parent.InsertBefore(ruby, child)
				}
			} else {
				if run == nil {
					run = NewRubyNode()
					run.SetPos(ruby.Pos())
					parent.InsertBefore(ruby, run)
				}
				run.AppendChild(segment)
			}
		}
		parent.RemoveChild(ruby)
	}
}

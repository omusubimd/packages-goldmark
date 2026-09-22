package internal

import (
	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/text"
)

var (
	KindRuby        = ast.NewNodeKind("Ruby")        // KindRuby は Ruby ノードの種類を表します。
	KindRubySegment = ast.NewNodeKind("RubySegment") // KindRubySegment は RubySegment ノードの種類を表します。
)

// RubyNode は Ruby ノードを表します。
//
// Ruby ノードは、子として一つ以上の RubySegment ノードを持ちます。
type RubyNode struct {
	ast.BaseInline
}

func (n *RubyNode) Kind() ast.NodeKind {
	return KindRuby
}

func (n *RubyNode) Dump(source []byte) *ast.NodeDump {
	return ast.NewNodeDump(n, nil)
}

func NewRubyNode() *RubyNode {
	n := &RubyNode{}
	n.Init(n)

	return n
}

// RubySegmentNode は Ruby ノードの子であり、ルビセグメントを表します。
//
// 一つのルビセグメントはベーステキストとルビテキスト（読み）の組です。
//
// ベーステキストは、ルビの対象となる文字列です。例えば、「青空」の「青」に「あお」とルビを振る場合、ベーステキストは「青」となります。ベーステキストはインライン要素です。したがって、強調やその他のインライン要素でありえます。
type RubySegmentNode struct {
	ast.BaseInline

	// RubyText はルビのテキストを表します。
	//
	// ルビのテキストは、ベーステキストの読みを表します。例えば、「青空」の「青」に「あお」とルビを振る場合、ルビのテキストは「あお」となります。
	//
	// ルビのテキストはプレーンテキストです。
	RubyText text.SingleLineValue
}

func (n *RubySegmentNode) Kind() ast.NodeKind {
	return KindRubySegment
}

func (n *RubySegmentNode) Dump(source []byte) *ast.NodeDump {
	return ast.NewNodeDump(n, map[string]any{
		"RubyText": n.RubyText.Str(source),
	})
}

func NewRubySegmentNode(rubyText text.SingleLineValue) *RubySegmentNode {
	n := &RubySegmentNode{
		RubyText: rubyText,
	}
	n.Init(n)
	return n
}

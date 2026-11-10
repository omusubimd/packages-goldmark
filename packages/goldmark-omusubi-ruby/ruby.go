// Package ruby adds Omusubi ruby syntax to Goldmark v2.
package ruby

import (
	"github.com/omusubimd/packages-goldmark/packages/goldmark-omusubi-ruby/internal"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/util"
)

type RubyNode = internal.RubyNode
type RubySegmentNode = internal.RubySegmentNode

var KindRuby = internal.KindRuby
var KindRubySegment = internal.KindRubySegment

type rubyParserExtension struct{}

// NewParser returns a Goldmark parser extension for Omusubi ruby syntax.
func NewParser() parser.Extension { return &rubyParserExtension{} }
func (*rubyParserExtension) ParserOptions(config *parser.Config) []parser.Option {
	return []parser.Option{
		parser.WithInlineParsers(util.Prioritized[parser.InlineParser](internal.NewRubyParserWithConfig(config), 199)),
		parser.WithASTTransformers(util.Prioritized[parser.ASTTransformer](&internal.RubyTransformer{}, 100)),
	}
}

// Parser is the default ruby parser extension.
var Parser = NewParser()

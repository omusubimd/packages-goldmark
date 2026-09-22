package internal_test

import (
	"testing"

	"github.com/omusubimd/packages-goldmark/goldmark-omusubi-ruby/internal"
	"github.com/stretchr/testify/assert"
	"github.com/yuin/goldmark/v2/text"
)

func TestRubyNode(t *testing.T) {
	t.Run("NewRubyNode 関数が期待した値を返す", func(t *testing.T) {
		rubyNode := internal.NewRubyNode()

		assert.NotNil(t, rubyNode)
	})

	rubyNode := internal.NewRubyNode()

	t.Run("Kind メソッドが期待した値を返す", func(t *testing.T) {
		kind := rubyNode.Kind()

		assert.Equal(t, internal.KindRuby, kind)
	})

	t.Run("Dump メソッドが期待した値を返す", func(t *testing.T) {
		dump := rubyNode.Dump([]byte(`[伍]<<ご>>`))

		assert.Equal(t, rubyNode, dump.Node)
	})
}

func TestRubySegmentNode(t *testing.T) {
	t.Run("NewRubySegmentNode 関数が期待した値を返す", func(t *testing.T) {
		rubyText := text.NewSingleLineValue([]byte("ご"), text.IdentityDecoder)
		rubySegmentNode := internal.NewRubySegmentNode(rubyText)

		assert.NotNil(t, rubySegmentNode)
		assert.Equal(t, rubyText, rubySegmentNode.RubyText)
	})

	rubySegmentNode := internal.NewRubySegmentNode(text.NewSingleLineValue([]byte("ご"), text.IdentityDecoder))

	t.Run("Kind メソッドが期待した値を返す", func(t *testing.T) {
		kind := rubySegmentNode.Kind()

		assert.Equal(t, internal.KindRubySegment, kind)
	})

	t.Run("Dump メソッドが期待した値を返す", func(t *testing.T) {
		dump := rubySegmentNode.Dump([]byte(`[伍]<<ご>>`))

		assert.Equal(t, rubySegmentNode, dump.Node)
	})
}

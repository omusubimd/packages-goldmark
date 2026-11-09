package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/text"
	"github.com/yuin/goldmark/v2/util"
)

func TestRubyParser(t *testing.T) {
	t.Run("NewRubyParser 関数が期待した値を返す", func(t *testing.T) {
		p := NewRubyParser()

		assert.NotNil(t, p)
	})

	p := NewRubyParser()

	t.Run("Trigger メソッドが期待した値を返す", func(t *testing.T) {
		trigger := p.Trigger()

		assert.Equal(t, []byte{'['}, trigger)
	})

	t.Run("Parse メソッドが期待した値を返す", func(t *testing.T) {
		cases := []struct {
			name     string
			source   string
			expected ast.Node
		}{
			{
				name:   "単一セグメントからなるルビ構文",
				source: "[青空]<<あおぞら>>",
				expected: ast.N(
					NewRubyNode(),
					
					ast.N(
						NewRubySegmentNode(text.NewSingleLineValue([]byte("あおぞら"), text.IdentityDecoder)),
						"青空",
					),
				),
			},
			{
				name:   "複数セグメントからなるルビ構文",
				source: "[花|緑|青]<<はな|ろく|しょう>>",
				expected: ast.N(
					NewRubyNode(),

					ast.N(
						NewRubySegmentNode(text.NewSingleLineValue([]byte("はな"), text.IdentityDecoder)),
						"花",
					),

					ast.N(
						NewRubySegmentNode(text.NewSingleLineValue([]byte("ろく"), text.IdentityDecoder)),
						"緑",
					),

					ast.N(
						NewRubySegmentNode(text.NewSingleLineValue([]byte("しょう"), text.IdentityDecoder)),
						"青",
					),
				),
			},
			{
				name:   "ルビのスキップを含むルビ構文",
				source: "[青天|の|霹靂]<<せいてん||へきれき>>",
				expected: ast.N(
					ast.NewParagraph(),
					ast.N(
						NewRubyNode(),
						ast.N(
							NewRubySegmentNode(text.NewSingleLineValue([]byte("せいてん"), text.IdentityDecoder)),
							"青天",
						),
					),
					"の",
					ast.N(
						NewRubyNode(),
						ast.N(
							NewRubySegmentNode(text.NewSingleLineValue([]byte("へきれき"), text.IdentityDecoder)),
							"霹靂",
						),
					),
				),
			},
		}

		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				source := []byte(c.source)
				parser := parser.New(parser.WithInlineParsers(util.Prioritized[parser.InlineParser](p, 199)))
				doc := parser.Parse(source)

				actualRuby := doc.FirstChild().FirstChild()

				assertNodeEqual(t, source, c.expected, actualRuby)
			})
		}
	})
}

func assertNodeEqual(
	t *testing.T,
	source []byte,
	expected ast.Node,
	actual ast.Node,
) {
	t.Helper()

	assert.Equal(t, expected.Kind(), actual.Kind())

	switch expected := expected.(type) {
	case *ast.Text:
		actual, ok := actual.(*ast.Text)
		require.True(t, ok)

		assert.Equal(
			t,
			expected.Value.Str(nil),
			actual.Value.Str(source),
		)

	case *RubySegmentNode:
		actual, ok := actual.(*RubySegmentNode)
		require.True(t, ok)

		assert.Equal(
			t,
			expected.RubyText.Str(nil),
			actual.RubyText.Str(source),
		)
	}

	expectedChild := expected.FirstChild()
	actualChild := actual.FirstChild()

	for expectedChild != nil && actualChild != nil {
		assertNodeEqual(t, source, expectedChild, actualChild)

		expectedChild = expectedChild.NextSibling()
		actualChild = actualChild.NextSibling()
	}

	assert.Nil(t, expectedChild)
	assert.Nil(t, actualChild)
}

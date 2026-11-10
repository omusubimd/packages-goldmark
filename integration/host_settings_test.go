package integration_test

import (
	"bytes"
	indent "github.com/omusubimd/packages-goldmark/packages/goldmark-omusubi-indent"
	ruby "github.com/omusubimd/packages-goldmark/packages/goldmark-omusubi-ruby"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"
	"testing"
)

func TestRubyUsesHostSettingsWithIndent(t *testing.T) {
	for _, extensions := range [][]parser.Extension{{indent.Parser, ruby.Parser}, {ruby.Parser, indent.Parser}} {
		p := parser.New(parser.WithExtensions(extensions...), parser.WithParseDelimiterFunc(parser.ParseDelimiterSimple), parser.WithEscapedSpace())
		source := []byte("　あ*「い」*う [あ*「い」*う|a\\ b]<<r|s>>")
		var out bytes.Buffer
		if err := html.New(html.WithExtensions(ruby.HTMLRenderer)).Render(&out, source, p.Parse(source)); err != nil {
			t.Fatal(err)
		}
		want := "<p>あ<em>「い」</em>う <ruby>あ<em>「い」</em>う<rt>r</rt>ab<rt>s</rt></ruby></p>\n"
		if out.String() != want {
			t.Fatalf("got %q, want %q", out.String(), want)
		}
	}
}

// Package integration_test verifies the independently consumable extensions together.
package integration_test

import (
	"bytes"
	"testing"

	indent "github.com/omusubimd/packages-goldmark/packages/goldmark-omusubi-indent"
	ruby "github.com/omusubimd/packages-goldmark/packages/goldmark-omusubi-ruby"
	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/extension"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"
)

func TestIndentWithRuby(t *testing.T) {
	for _, extensions := range [][]parser.Extension{
		{indent.Parser, ruby.Parser}, {ruby.Parser, indent.Parser},
	} {
		p := parser.New(parser.WithExtensions(extensions...))
		r := html.New(html.WithExtensions(ruby.HTMLRenderer))
		for _, c := range []struct{ input, want string }{
			{"前文\n　[**青**|空]<<あお|そら>>\n　[青天|の|霹靂]<<せいてん||へきれき>>", "<p>前文</p>\n<p><ruby><strong>青</strong><rt>あお</rt>空<rt>そら</rt></ruby></p>\n<p><ruby>青天<rt>せいてん</rt></ruby>の<ruby>霹靂<rt>へきれき</rt></ruby></p>\n"},
			{"> 　[青空]<<あおぞら>>\n> 　次の段落", "<blockquote>\n<p><ruby>青空<rt>あおぞら</rt></ruby></p>\n<p>次の段落</p>\n</blockquote>\n"},
			{"　[青空]<<あおぞら>>\n　\n---", "<p><ruby>青空<rt>あおぞら</rt></ruby></p>\n<hr>\n"},
			{"```md\n　[青空]<<あおぞら>>\n```", "<pre><code class=\"language-md\">　[青空]&lt;&lt;あおぞら&gt;&gt;\n</code></pre>\n"},
		} {
			source := []byte(c.input)
			doc := p.Parse(source)
			var out bytes.Buffer
			if err := r.Render(&out, source, doc); err != nil {
				t.Fatal(err)
			}
			if out.String() != c.want {
				t.Fatalf("%q: got %q, want %q", c.input, out.String(), c.want)
			}
			_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
				if entering && n.Kind().String() == "OmusubiIndentSeparator" {
					t.Fatal("temporary node leaked")
				}
				return ast.WalkContinue, nil
			})
		}
	}
}

func TestIndentPreservesInlineExtensions(t *testing.T) {
	// Unlike ruby label fragments, indent paragraph bodies use the host's inline parsers.
	p := parser.New(parser.WithExtensions(indent.Parser, extension.StrikethroughParser))
	r := html.New(html.WithExtensions(extension.StrikethroughHTMLRenderer))
	source := []byte("a\n　~~deleted~~")
	var out bytes.Buffer
	if err := r.Render(&out, source, p.Parse(source)); err != nil {
		t.Fatal(err)
	}
	if out.String() != "<p>a</p>\n<p><del>deleted</del></p>\n" {
		t.Fatal(out.String())
	}
}

func TestRubyIndentBoundaryMatrix(t *testing.T) {
	for _, extensions := range [][]parser.Extension{{indent.Parser, ruby.Parser}, {ruby.Parser, indent.Parser}} {
		p := parser.New(parser.WithExtensions(extensions...))
		r := html.New(html.WithXHTML(), html.WithExtensions(ruby.HTMLRenderer))
		for _, c := range []struct{ name, input, want string }{
			{"paragraph", "前文\n　[青天|の|霹靂]<<せいてん||へきれき>>", "<p>前文</p>\n<p><ruby>青天<rt>せいてん</rt></ruby>の<ruby>霹靂<rt>へきれき</rt></ruby></p>\n"},
			{"blank-setext", "　[青空]<<あおぞら>>\n　\n---", "<p><ruby>青空<rt>あおぞら</rt></ruby></p>\n<hr />\n"},
			{"quote", "> 　[青空]<<あおぞら>>\n> 　\n> ---", "<blockquote>\n<p><ruby>青空<rt>あおぞら</rt></ruby></p>\n<hr />\n</blockquote>\n"},
			{"loose-list", "- [青空]<<あおぞら>>\n  　\n  次", "<ul>\n<li>\n<p><ruby>青空<rt>あおぞら</rt></ruby></p>\n<p>次</p>\n</li>\n</ul>\n"},
			{"empty-readings", "　[青|空]<<|>>\n　[海]<<うみ>>", "<p>青空</p>\n<p><ruby>海<rt>うみ</rt></ruby></p>\n"},
			{"inline-code", "　`[青]<<あお>>`\n　[海]<<うみ>>", "<p><code>[青]&lt;&lt;あお&gt;&gt;</code></p>\n<p><ruby>海<rt>うみ</rt></ruby></p>\n"},
			{"fenced-code", "```md\n　[青]<<あお>>\n```", "<pre><code class=\"language-md\">　[青]&lt;&lt;あお&gt;&gt;\n</code></pre>\n"},
			{"malformed-recovery", "　[青|空]<<あおぞら>>\n　[海]<<うみ>>", "<p>[青|空]&lt;&lt;あおぞら&gt;&gt;</p>\n<p><ruby>海<rt>うみ</rt></ruby></p>\n"},
			{"escaped-reading", "　[A&amp;B]<<&lt;script&gt;>>", "<p><ruby>A&amp;B<rt>&lt;script&gt;</rt></ruby></p>\n"},
		} {
			t.Run(c.name, func(t *testing.T) {
				source := []byte(c.input)
				var out bytes.Buffer
				if err := r.Render(&out, source, p.Parse(source)); err != nil {
					t.Fatal(err)
				}
				if out.String() != c.want {
					t.Fatalf("got %q, want %q", out.String(), c.want)
				}
			})
		}
	}
}

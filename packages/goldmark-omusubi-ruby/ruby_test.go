package ruby_test

import (
	"bytes"
	"encoding/json"
	"os"
	"sync"
	"testing"

	ruby "github.com/omusubimd/packages-goldmark/packages/goldmark-omusubi-ruby"
	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"
)

func TestUnifiedFixtures(t *testing.T) {
	data, err := os.ReadFile("testdata/ruby.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []struct{ Input, Output, GoldmarkOutput string }
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}
	p := parser.New(parser.WithExtensions(ruby.Parser))
	r := html.New(html.WithExtensions(ruby.HTMLRenderer))
	for _, fixture := range fixtures {
		t.Run(fixture.Input, func(t *testing.T) {
			var out bytes.Buffer
			source := []byte(fixture.Input)
			if err := r.Render(&out, source, p.Parse(source)); err != nil {
				t.Fatal(err)
			}
			expected := fixture.Output
			if fixture.GoldmarkOutput != "" {
				expected = fixture.GoldmarkOutput
			}
			if out.String() != expected {
				t.Fatalf("got %q, want %q", out.String(), expected)
			}
		})
	}
}

func TestMarkdownIntegration(t *testing.T) {
	cases := []struct{ input, output string }{
		{"[ a ]<<A>>", "<p><ruby> a <rt>A</rt></ruby></p>\n"},
		{"[ ]<< >>", "<p><ruby> <rt> </rt></ruby></p>\n"},
		{"[<https://example.com>]<<URL>>", "<p><ruby><a href=\"https://example.com\">https://example.com</a><rt>URL</rt></ruby></p>\n"},
		{"[# a]<<A>>", "<p><ruby># a<rt>A</rt></ruby></p>\n"},
		{"[a| ]<<A|>>", "<p><ruby>a<rt>A</rt></ruby> </p>\n"},
		{"before [a]<<A>> after", "<p>before <ruby>a<rt>A</rt></ruby> after</p>\n"},
		{"# [a]<<A>>", "<h1><ruby>a<rt>A</rt></ruby></h1>\n"},
		{"- [a]<<A>>", "<ul>\n<li><ruby>a<rt>A</rt></ruby></li>\n</ul>\n"},
		{"> [a]<<A>>", "<blockquote>\n<p><ruby>a<rt>A</rt></ruby></p>\n</blockquote>\n"},
		{"`[a]<<A>>`", "<p><code>[a]&lt;&lt;A&gt;&gt;</code></p>\n"},
		{"\\[a]<<あ>>", "<p>[a]&lt;&lt;あ&gt;&gt;</p>\n"},
		{"[link](https://example.com)", "<p><a href=\"https://example.com\">link</a></p>\n"},
		{"[a]<<あ>い>>", "<p>[a]&lt;&lt;あ&gt;い&gt;&gt;</p>\n"},
		{"[a]<<\\*&#65;&amp;>>", "<p><ruby>a<rt>*A&amp;</rt></ruby></p>\n"},
		{"[**a|b**]<<A|B>>", "<p><ruby>**a<rt>A</rt>b**<rt>B</rt></ruby></p>\n"},
		{"[a\nb]<<あ>>", "<p>[a\nb]&lt;&lt;あ&gt;&gt;</p>\n"},
		{"[a]<<あ\nい>>", "<p>[a]&lt;&lt;あ\nい&gt;&gt;</p>\n"},
	}
	p := parser.New(parser.WithExtensions(ruby.NewParser()))
	r := html.New(html.WithExtensions(ruby.NewHTMLRenderer()))
	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			source := []byte(c.input)
			var out bytes.Buffer
			if err := r.Render(&out, source, p.Parse(source)); err != nil {
				t.Fatal(err)
			}
			if out.String() != c.output {
				t.Fatalf("got %q, want %q", out.String(), c.output)
			}
		})
	}
}

func TestAST(t *testing.T) {
	source := []byte("prefix [**青**|の|空]<<あお||そら>> suffix")
	doc := parser.New(parser.WithExtensions(ruby.Parser)).Parse(source)
	paragraph := doc.FirstChild()
	first := paragraph.FirstChild().NextSibling()
	n, ok := first.(*ruby.RubyNode)
	if !ok || n.Kind() != ruby.KindRuby {
		t.Fatalf("expected ruby, got %T", first)
	}
	segment := n.FirstChild().(*ruby.RubySegmentNode)
	if segment.Kind() != ruby.KindRubySegment || segment.RubyText.Value(source) != "あお" {
		t.Fatal("incorrect segment")
	}
	found := false
	for child := segment.FirstChild(); child != nil; child = child.NextSibling() {
		if _, ok := child.(*ast.Strong); ok {
			found = true
		}
	}
	if !found {
		t.Fatal("label emphasis missing")
	}
	if text := n.NextSibling().(*ast.Text).Value.Value(source); text != "の" {
		t.Fatalf("skipped label: %q", text)
	}
	if _, ok := n.NextSibling().NextSibling().(*ruby.RubyNode); !ok {
		t.Fatal("second ruby missing")
	}
}

func TestConcurrentConversion(t *testing.T) {
	p := parser.New(parser.WithExtensions(ruby.Parser))
	r := html.New(html.WithExtensions(ruby.HTMLRenderer))
	var wg sync.WaitGroup
	for range 16 {
		wg.Go(func() {
			source := []byte("[**青**|空]<<あお|そら>>")
			var out bytes.Buffer
			if err := r.Render(&out, source, p.Parse(source)); err != nil {
				t.Error(err)
			}
			if out.String() != "<p><ruby><strong>青</strong><rt>あお</rt>空<rt>そら</rt></ruby></p>\n" {
				t.Error(out.String())
			}
		})
	}
	wg.Wait()
}

func TestUnsafeHTML(t *testing.T) {
	source := []byte("[<b>青</b>]<<&lt;script&gt;>>")
	p := parser.New(parser.WithExtensions(ruby.Parser))
	for _, c := range []struct {
		unsafe   bool
		expected string
	}{
		{false, "<p><ruby><!-- raw HTML omitted -->青<!-- raw HTML omitted --><rt>&lt;script&gt;</rt></ruby></p>\n"},
		{true, "<p><ruby><b>青</b><rt>&lt;script&gt;</rt></ruby></p>\n"},
	} {
		opts := []html.Option{html.WithExtensions(ruby.HTMLRenderer)}
		if c.unsafe {
			opts = append(opts, html.WithUnsafe())
		}
		var out bytes.Buffer
		if err := html.New(opts...).Render(&out, source, p.Parse(source)); err != nil {
			t.Fatal(err)
		}
		if out.String() != c.expected {
			t.Fatalf("got %q, want %q", out.String(), c.expected)
		}
	}
}

func FuzzConversion(f *testing.F) {
	for _, seed := range []string{"[青|空]<<あお|そら>>", "[a]<<>>", "[a||b]<<||>>", "[**a**]<<&lt;x&gt;>>", "[a]<<a>b>>", "[a]", "[a\nb]<<c>>", "[ ]<< >>", "[a]<<\x00>>"} {
		f.Add(seed)
	}
	p := parser.New(parser.WithExtensions(ruby.Parser))
	r := html.New(html.WithExtensions(ruby.HTMLRenderer))
	f.Fuzz(func(t *testing.T, input string) {
		source := []byte(input)
		var first, second bytes.Buffer
		if err := r.Render(&first, source, p.Parse(source)); err != nil {
			t.Fatal(err)
		}
		if err := r.Render(&second, source, p.Parse(source)); err != nil {
			t.Fatal(err)
		}
		if first.String() != second.String() {
			t.Fatal("conversion is not deterministic")
		}
	})
}

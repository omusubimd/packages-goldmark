package indent_test

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"

	indent "github.com/omusubimd/packages-goldmark/packages/goldmark-omusubi-indent"
	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"
)

type fixture struct {
	Input          string  `json:"input"`
	UnifiedOutput  string  `json:"unifiedOutput"`
	GoldmarkOutput *string `json:"goldmarkOutput"`
	Difference     string  `json:"difference"`
}

func TestUnifiedFixtures(t *testing.T) {
	data, err := os.ReadFile("testdata/unified.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []fixture
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}
	p := parser.New(parser.WithExtensions(indent.Parser))
	r := html.New(html.WithXHTML())
	for _, f := range fixtures {
		t.Run(f.Input, func(t *testing.T) {
			source := []byte(f.Input)
			var out bytes.Buffer
			if err := r.Render(&out, source, p.Parse(source)); err != nil {
				t.Fatal(err)
			}
			want := f.UnifiedOutput
			if f.GoldmarkOutput != nil {
				if f.Difference == "" {
					t.Fatal("unexplained compatibility difference")
				}
				want = *f.GoldmarkOutput
			}
			// The default engines differ in whether a document ends with a newline.
			if strings.Trim(out.String(), "\r\n") != strings.Trim(want, "\r\n") {
				t.Fatalf("got %q, want %q (difference: %s)", out.String(), want, f.Difference)
			}
		})
	}
}

func render(t *testing.T, p parser.Parser, input string, opts ...html.Option) string {
	t.Helper()
	source := []byte(input)
	var out bytes.Buffer
	if err := html.New(opts...).Render(&out, source, p.Parse(source)); err != nil {
		t.Fatal(err)
	}
	return out.String()
}

func TestNativeMarkdownUnaffected(t *testing.T) {
	extended := parser.New(parser.WithExtensions(indent.NewParser()))
	native := parser.New()
	for _, input := range []string{
		"# heading\n\ntext *emphasis* **strong** [link](/url)\nnext",
		"- one\n- two\n\n> quote", "[ref]: /url\n\n[ref]",
		"```md\n　inside\n```", strings.Repeat(" ", 4) + "　code\n    next", "\t　code", "`　inline`", "text　inside",
		"<div>\n　inside\n</div>", "<script>\n　inside\n</script>",
		"、other UTF-8 punctuation\nあいう",
	} {
		t.Run(input, func(t *testing.T) {
			if got, want := render(t, extended, input, html.WithUnsafe()), render(t, native, input, html.WithUnsafe()); got != want {
				t.Fatalf("got %q, want native %q", got, want)
			}
			source := []byte(input)
			assertSameAST(t, source, extended.Parse(source), native.Parse(source))

		})
	}
	input := "a\n　b"
	if got := render(t, native, input); got != "<p>a\n　b</p>\n" {
		t.Fatal(got)
	}
	if got := render(t, extended, input); got != "<p>a</p>\n<p>b</p>\n" {
		t.Fatal(got)
	}
}

func TestASTAndSource(t *testing.T) {
	source := []byte("前\n　**青**\n次\n　\n　終")
	doc := parser.New(parser.WithExtensions(indent.Parser)).Parse(source)
	if doc.ChildCount() != 3 {
		t.Fatalf("paragraph count: %d", doc.ChildCount())
	}
	second := doc.FirstChild().NextSibling().(*ast.Paragraph)
	if second.Source()[0].Start != len("前\n　") {
		t.Fatal("source position does not exclude marker")
	}
	if _, ok := second.FirstChild().(*ast.Strong); !ok {
		t.Fatal("strong node missing")
	}
	if err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			switch n.Kind() {
			case ast.KindDocument, ast.KindParagraph, ast.KindText, ast.KindStrong:
			default:
				t.Fatalf("unexpected custom node: %v", n.Kind())
			}
		}
		return ast.WalkContinue, nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestMarkerOnlyDoesNotBridgeSetext(t *testing.T) {
	p := parser.New(parser.WithExtensions(indent.Parser))
	for _, c := range []struct{ input, want string }{
		{"one\n　\n---", "<p>one</p>\n<hr>\n"},
		{"one\n　\n===", "<p>one</p>\n<p>===</p>\n"},
		{"　\n　", ""},
	} {
		if got := render(t, p, c.input); got != c.want {
			t.Fatalf("%q: got %q, want %q", c.input, got, c.want)
		}
	}
}

func TestConcurrentParsing(t *testing.T) {
	p := parser.New(parser.WithExtensions(indent.Parser))
	var wg sync.WaitGroup
	for range 16 {
		wg.Go(func() {
			if got := render(t, p, "　a\n　b\n　\nend"); got != "<p>a</p>\n<p>b</p>\n<p>end</p>\n" {
				t.Error(got)
			}
		})
	}
	wg.Wait()
}

func FuzzParsing(f *testing.F) {
	for _, s := range []string{"a\n　b", "　", "- a\n  　b", "> 　a", "　a\n---", "a\n　\n---", "```\n　a\n```", "a\r\n　b", "\xe3\x80", "　\x00"} {
		f.Add(s)
	}
	extended := parser.New(parser.WithExtensions(indent.Parser))
	native := parser.New()
	f.Fuzz(func(t *testing.T, input string) {
		original := []byte(input)
		source := bytes.Clone(original)
		first := render(t, extended, input)
		if second := render(t, extended, input); first != second {
			t.Fatal("nondeterministic output")
		}
		doc := extended.Parse(source)
		if !bytes.Equal(original, source) {
			t.Fatal("source modified")
		}
		_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
			if entering && n.Kind().String() == "OmusubiIndentSeparator" {
				t.Fatal("temporary node leaked")
			}
			return ast.WalkContinue, nil
		})
		if !strings.Contains(input, "　") {
			if want := render(t, native, input); first != want {
				t.Fatalf("unmarked input changed: %q != %q", first, want)
			}
		}
	})
}

func assertSameAST(t *testing.T, source []byte, got, want ast.Node) {
	t.Helper()
	if got.Kind() != want.Kind() || got.Pos() != want.Pos() || got.ChildCount() != want.ChildCount() {
		t.Fatalf("AST shape changed: %v != %v", got.Kind(), want.Kind())
	}
	if !reflect.DeepEqual(got.Dump(source).Properties, want.Dump(source).Properties) {
		t.Fatalf("AST properties changed for %v", got.Kind())
	}
	if b, ok := got.(ast.BlockNode); ok {
		w := want.(ast.BlockNode)
		if !reflect.DeepEqual(b.Source(), w.Source()) || b.HasBlankPreviousLines() != w.HasBlankPreviousLines() {
			t.Fatalf("block source changed for %v", got.Kind())
		}
	}
	gc, wc := got.FirstChild(), want.FirstChild()
	for gc != nil {
		assertSameAST(t, source, gc, wc)
		gc, wc = gc.NextSibling(), wc.NextSibling()
	}
}

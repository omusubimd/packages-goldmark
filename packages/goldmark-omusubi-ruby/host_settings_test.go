package ruby_test

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	ruby "github.com/omusubimd/packages-goldmark/packages/goldmark-omusubi-ruby"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"
	"github.com/yuin/goldmark/v2/text"
)

func convertWith(t *testing.T, p parser.Parser, input string) string {
	t.Helper()
	source := []byte(input)
	var out bytes.Buffer
	if err := html.New(html.WithExtensions(ruby.HTMLRenderer)).Render(&out, source, p.Parse(source)); err != nil {
		t.Fatal(err)
	}
	return out.String()
}

func TestHostDelimiterSettings(t *testing.T) {
	disable := func(reader text.Reader, minimum int, processor parser.DelimiterProcessor, pc parser.Context) *parser.Delimiter {
		d := parser.ParseDelimiter(reader, minimum, processor, pc)
		if d != nil {
			d.CanOpen = false
			d.CanClose = false
		}
		return d
	}
	p := parser.New(parser.WithExtensions(ruby.Parser), parser.WithParseDelimiterFunc(disable))
	for _, c := range []struct{ input, want string }{
		{"*outer* [*base*]<<r>>", "<p>*outer* <ruby>*base*<rt>r</rt></ruby></p>\n"},
		{"[**base**|_tail_]<<b|t>>", "<p><ruby>**base**<rt>b</rt>_tail_<rt>t</rt></ruby></p>\n"},
		{"[**base**]<<>>", "<p>**base**</p>\n"},
	} {
		t.Run(c.input, func(t *testing.T) {
			if got := convertWith(t, p, c.input); got != c.want {
				t.Fatalf("got %q, want %q", got, c.want)
			}
		})
	}
	// Reusing the exported extension must not share a host's configuration.
	normal := parser.New(parser.WithExtensions(ruby.Parser))
	if got := convertWith(t, normal, "[*base*]<<r>>"); got != "<p><ruby><em>base</em><rt>r</rt></ruby></p>\n" {
		t.Fatal(got)
	}
}

func TestHostEscapedSpaceSettings(t *testing.T) {
	p := parser.New(parser.WithExtensions(ruby.Parser), parser.WithEscapedSpace())
	for _, c := range []struct{ input, want string }{
		{`a\ b [a\ b]<<r\ s>>`, "<p>ab <ruby>ab<rt>rs</rt></ruby></p>\n"},
		{`[a\ ]<<r>>`, "<p><ruby>a<rt>r</rt></ruby></p>\n"},
		{`[a\\ ]<<r>>`, "<p><ruby>a\\ <rt>r</rt></ruby></p>\n"},
		{"[`a\\ b`]<<r>>", "<p><ruby><code>a\\ b</code><rt>r</rt></ruby></p>\n"},
		{`[a\ b]<<>>`, "<p>ab</p>\n"},
	} {
		t.Run(c.input, func(t *testing.T) {
			if got := convertWith(t, p, c.input); got != c.want {
				t.Fatalf("got %q, want %q", got, c.want)
			}
		})
	}
	normal := parser.New(parser.WithExtensions(ruby.Parser))
	if got := convertWith(t, normal, `[a\ b]<<r>>`); got != "<p><ruby>a\\ b<rt>r</rt></ruby></p>\n" {
		t.Fatal(got)
	}
}

func TestHostCJKDelimiterSettings(t *testing.T) {
	p := parser.New(parser.WithExtensions(ruby.Parser), parser.WithParseDelimiterFunc(parser.ParseDelimiterSimple))
	input := "あ*「い」*う [あ*「い」*う]<<r>>"
	want := "<p>あ<em>「い」</em>う <ruby>あ<em>「い」</em>う<rt>r</rt></ruby></p>\n"
	if got := convertWith(t, p, input); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func FuzzHostSettings(f *testing.F) {
	for _, seed := range []string{"[あ*「い」*う]<<r>>", `[a\ b]<<r\ s>>`, `[a\ ]<<r>>`, `[a\\ ]<<r>>`, "[`a\\ b`]<<r>>", "[a|b]<<>>", "あ*「い」*う", "[a]<<\x00>>"} {
		f.Add(seed)
	}
	p := parser.New(parser.WithExtensions(ruby.Parser), parser.WithParseDelimiterFunc(parser.ParseDelimiterSimple), parser.WithEscapedSpace())
	native := parser.New(parser.WithParseDelimiterFunc(parser.ParseDelimiterSimple), parser.WithEscapedSpace())
	f.Fuzz(func(t *testing.T, input string) {
		first := convertWith(t, p, input)
		if second := convertWith(t, p, input); first != second {
			t.Fatal("host configuration is not deterministic")
		}
		if !strings.Contains(input, "[") {
			if want := convertWith(t, native, input); first != want {
				t.Fatalf("unmarked input differs from host: %q != %q", first, want)
			}
		}
	})
}

func TestConcurrentHostSettings(t *testing.T) {
	p := parser.New(parser.WithExtensions(ruby.Parser), parser.WithParseDelimiterFunc(parser.ParseDelimiterSimple), parser.WithEscapedSpace())
	const input = "[あ*「い」*う|a\\ b]<<r|s>>"
	const want = "<p><ruby>あ<em>「い」</em>う<rt>r</rt>ab<rt>s</rt></ruby></p>\n"
	var wg sync.WaitGroup
	for range 16 {
		wg.Go(func() {
			if got := convertWith(t, p, input); got != want {
				t.Error(got)
			}
		})
	}
	wg.Wait()
}

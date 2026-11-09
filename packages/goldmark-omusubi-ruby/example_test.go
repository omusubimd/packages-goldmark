package ruby_test

import (
	ruby "github.com/omusubimd/packages-goldmark/packages/goldmark-omusubi-ruby"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"
	"os"
)

func Example() {
	p := parser.New(parser.WithExtensions(ruby.Parser))
	r := html.New(html.WithExtensions(ruby.HTMLRenderer))
	source := []byte("[青天|の|霹靂]<<せいてん||へきれき>>")
	if err := r.Render(os.Stdout, source, p.Parse(source)); err != nil {
		panic(err)
	}
	// Output: <p><ruby>青天<rt>せいてん</rt></ruby>の<ruby>霹靂<rt>へきれき</rt></ruby></p>
}

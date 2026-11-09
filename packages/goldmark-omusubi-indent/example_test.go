package indent_test

import (
	indent "github.com/omusubimd/packages-goldmark/packages/goldmark-omusubi-indent"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"
	"os"
)

func Example() {
	p := parser.New(parser.WithExtensions(indent.Parser))
	source := []byte("こんにちは\n　おむすび\n　**大作戦**")
	if err := html.New().Render(os.Stdout, source, p.Parse(source)); err != nil {
		panic(err)
	}
	// Output:
	// <p>こんにちは</p>
	// <p>おむすび</p>
	// <p><strong>大作戦</strong></p>
}

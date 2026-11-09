package internal

import (
	"bytes"

	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/text"
)

type RubyParser struct{}

func (p *RubyParser) Trigger() []byte {
	return []byte{'['}
}

func (p *RubyParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, segment := block.PeekLine()
	if len(line) < 7 || line[0] != '[' {
		return nil
	}

	labelEnd := bytes.IndexByte(line[1:], ']')
	if labelEnd < 1 {
		return nil
	}
	labelEnd++
	if labelEnd+3 > len(line) || line[labelEnd+1] != '<' || line[labelEnd+2] != '<' {
		return nil
	}

	rubyStart := labelEnd + 3
	rubyLength := bytes.Index(line[rubyStart:], []byte(">>"))
	if rubyLength < 0 {
		return nil
	}
	rubyEnd := rubyStart + rubyLength

	labelStart := 1
	rubyPartStart := rubyStart
	ruby := NewRubyNode()
	for {
		labelPartEnd := labelEnd
		if i := bytes.IndexByte(line[labelStart:labelEnd], '|'); i >= 0 {
			labelPartEnd = labelStart + i
		}
		if labelPartEnd == labelStart {
			return nil
		}

		rubyPartEnd := rubyEnd
		if i := bytes.IndexByte(line[rubyPartStart:rubyEnd], '|'); i >= 0 {
			rubyPartEnd = rubyPartStart + i
		}

		baseSegment := text.NewSegment(segment.Start+labelStart, segment.Start+labelPartEnd)
		rubySegment := NewRubySegmentNode(text.NewSingleLineValueFromSegment(
			text.NewSegment(segment.Start+rubyPartStart, segment.Start+rubyPartEnd),
			block.Decoder(),
		))
		rubySegment.AppendChild(ast.NewText(text.NewSingleLineValueFromSegment(baseSegment, block.Decoder())))
		ruby.AppendChild(rubySegment)

		labelsDone := labelPartEnd == labelEnd
		rubyPartsDone := rubyPartEnd == rubyEnd
		if labelsDone || rubyPartsDone {
			if labelsDone != rubyPartsDone {
				return nil
			}
			break
		}
		labelStart = labelPartEnd + 1
		rubyPartStart = rubyPartEnd + 1
	}

	block.Advance(rubyEnd + 2)
	return ruby
}

func NewRubyParser() *RubyParser {
	return &RubyParser{}
}

package internal

import (
	"bytes"

	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/text"
)

type RubyParser struct {
	labels parser.Parser
}

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
	rubyLength := bytes.IndexByte(line[rubyStart:], '>')
	if rubyLength < 0 || rubyStart+rubyLength+1 >= len(line) || line[rubyStart+rubyLength+1] != '>' {
		return nil
	}
	rubyEnd := rubyStart + rubyLength

	labels := bytes.Split(line[1:labelEnd], []byte{'|'})
	readings := bytes.Split(line[rubyStart:rubyEnd], []byte{'|'})
	if len(labels) != len(readings) {
		return nil
	}
	for _, label := range labels {
		if len(label) == 0 {
			return nil
		}
	}

	labelStart, readingStart := 1, rubyStart
	ruby := NewRubyNode()
	for i, label := range labels {
		reading := readings[i]
		rubySegment := NewRubySegmentNode(text.NewSingleLineValueFromSegment(
			text.NewSegment(segment.Start+readingStart, segment.Start+readingStart+len(reading)),
			block.Decoder(),
		))
		appendLabel(rubySegment, p.labels, label, segment.Start+labelStart, pc, block.Decoder())
		ruby.AppendChild(rubySegment)
		labelStart += len(label) + 1
		readingStart += len(reading) + 1
	}

	block.Advance(rubyEnd + 2)
	return ruby
}

func NewRubyParser() *RubyParser {
	return NewRubyParserWithConfig(&parser.Config{})
}

// NewRubyParserWithConfig applies the host's supported inline settings to labels.
func NewRubyParserWithConfig(config *parser.Config) *RubyParser {
	return &RubyParser{labels: newLabelParser(config)}
}

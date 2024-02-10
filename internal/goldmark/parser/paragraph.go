package parser

import (
	"github.com/rprtr258/flatnotes/internal/goldmark/ast"
	"github.com/rprtr258/flatnotes/internal/goldmark/text"
	"github.com/rprtr258/flatnotes/internal/goldmark/util"
)

type paragraphParser struct{}

var defaultParagraphParser = &paragraphParser{}

// NewParagraphParser returns a new BlockParser that
// parses paragraphs.
func NewParagraphParser() BlockParser {
	return defaultParagraphParser
}

func (b *paragraphParser) Trigger() []byte {
	return nil
}

func (b *paragraphParser) Open(parent ast.Node, reader text.Reader, pc Context) (ast.Node, State) {
	_, segment := reader.PeekLine()
	segment = segment.TrimLeftSpace(reader.Source())
	if segment.IsEmpty() {
		return nil, NoChildren
	}
	node := ast.NewParagraph()
	node.Lines().Append(segment)
	reader.Advance(segment.Len() - 1)
	return node, NoChildren
}

func (b *paragraphParser) Continue(node ast.Node, reader text.Reader, pc Context) State {
	line, segment := reader.PeekLine()
	if util.IsBlank(line) {
		return Close
	}
	node.Lines().Append(segment)
	reader.Advance(segment.Len() - 1)
	return Continue | NoChildren
}

func (b *paragraphParser) Close(node ast.Node, reader text.Reader, pc Context) {
	lines := node.Lines()
	if len(lines.Values) != 0 {
		// trim leading spaces
		for i, l := range lines.Values {
			lines.Values[i] = l.TrimLeftSpace(reader.Source())
		}

		// trim trailing spaces
		length := len(lines.Values)
		lastLine := node.Lines().Values[length-1]
		node.Lines().Values[length-1] = lastLine.TrimRightSpace(reader.Source())
	}
	if len(lines.Values) == 0 {
		node.Parent().RemoveChild(node.Parent(), node)
		return
	}
}

func (b *paragraphParser) CanInterruptParagraph() bool {
	return false
}

func (b *paragraphParser) CanAcceptIndentedLine() bool {
	return false
}

package parser

import (
	"github.com/rprtr258/flatnotes/internal/goldmark/ast"
	"github.com/rprtr258/flatnotes/internal/goldmark/text"
	"github.com/rprtr258/flatnotes/internal/goldmark/util"
)

type linkReferenceParagraphTransformer struct{}

// LinkReferenceParagraphTransformer is a ParagraphTransformer implementation
// that parses and extracts link reference from paragraphs.
var LinkReferenceParagraphTransformer = &linkReferenceParagraphTransformer{}

func (p *linkReferenceParagraphTransformer) Transform(node *ast.Paragraph, reader text.Reader, pc Context) {
	lines := node.Lines()
	block := text.NewBlockReader(reader.Source(), lines)
	removes := [][2]int{}
	for {
		start, end := parseLinkReferenceDefinition(block, pc)
		if start > -1 {
			if start == end {
				end++
			}
			removes = append(removes, [2]int{start, end})
			continue
		}
		break
	}

	offset := 0
	for _, remove := range removes {
		if len(lines.Values) == 0 {
			break
		}
		s := lines.Values[remove[1]-offset : len(lines.Values)]
		lines.SetSliced(0, remove[0]-offset)
		lines.Append(s...)
		offset = remove[1]
	}

	if len(lines.Values) == 0 {
		t := ast.NewTextBlock()
		t.SetBlankPreviousLines(node.HasBlankPreviousLines())
		node.Parent().ReplaceChild(node.Parent(), node, t)
		return
	}

	node.SetLines(lines)
}

func parseLinkReferenceDefinition(block text.Reader, pc Context) (start, end int) {
	block.SkipSpaces()
	line, _ := block.PeekLine()
	if line == nil {
		return -1, -1
	}

	startLine, _ := block.Position()
	width, pos := util.IndentWidth(line, 0)
	if width > 3 {
		return -1, -1
	}

	if width != 0 {
		pos++
	}
	if line[pos] != '[' {
		return -1, -1
	}

	block.Advance(pos + 1)
	segments, found := block.FindClosure('[', ']', linkFindClosureOptions)
	if !found {
		return -1, -1
	}

	var label []byte
	if len(segments.Values) == 1 {
		label = block.Value(segments.Values[0])
	} else {
		for _, segment := range segments.Values {
			label = append(label, block.Value(segment)...)
		}
	}
	if util.IsBlank(label) {
		return -1, -1
	}
	if block.Peek() != ':' {
		return -1, -1
	}
	block.Advance(1)
	block.SkipSpaces()
	destination, ok := parseLinkDestination(block)
	if !ok {
		return -1, -1
	}
	line, _ = block.PeekLine()
	isNewLine := line == nil || util.IsBlank(line)

	endLine, _ := block.Position()
	_, spaces, _ := block.SkipSpaces()
	opener := block.Peek()
	if opener != '"' && opener != '\'' && opener != '(' {
		if !isNewLine {
			return -1, -1
		}
		ref := NewReference(label, destination, nil)
		pc.AddReference(ref)
		return startLine, endLine + 1
	}
	if spaces == 0 {
		return -1, -1
	}
	block.Advance(1)
	closer := opener
	if opener == '(' {
		closer = ')'
	}
	segments, found = block.FindClosure(opener, closer, linkFindClosureOptions)
	if !found {
		if !isNewLine {
			return -1, -1
		}
		ref := NewReference(label, destination, nil)
		pc.AddReference(ref)
		block.AdvanceLine()
		return startLine, endLine + 1
	}
	var title []byte
	if len(segments.Values) == 1 {
		title = block.Value(segments.Values[0])
	} else {
		for _, s := range segments.Values {
			title = append(title, block.Value(s)...)
		}
	}

	line, _ = block.PeekLine()
	if line != nil && !util.IsBlank(line) {
		if !isNewLine {
			return -1, -1
		}
		ref := NewReference(label, destination, title)
		pc.AddReference(ref)
		return startLine, endLine
	}

	endLine, _ = block.Position()
	ref := NewReference(label, destination, title)
	pc.AddReference(ref)
	return startLine, endLine + 1
}

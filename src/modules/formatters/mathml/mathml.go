package mathml

import (
	"../.."
	"errors"
	"log"
)

func Provide() modules.Formatter {
	return MathmlFormatter{}
}

type MathmlFormatter struct{}

type TagType int

const (
	NEW_TAG TagType = iota
	CLOSE_TAG
	OPEN_TAG
)

type NodeType int

const (
	TEXT_NODE NodeType = iota
	TAG_NODE
)

type LatexNode struct {
	tag      string
	children []LatexNode
	nodetype NodeType
}

func (_ MathmlFormatter) Format(content string) string {
	if tree, err := buildTree(content); err == nil {
		return convert(tree)
	}
	return content
}

func buildTree(content string) (LatexNode, error) {
	parseError := false

	stack := make([]LatexNode, 0)
	i := 0
	curTag := LatexNode{tag: "root", nodetype: TAG_NODE}
	for {
		j, tagType := index(content, i)
		if j < 0 {
			curTag.children = append(curTag.children, LatexNode{
				tag:      content[i:],
				nodetype: TEXT_NODE,
			})
			break
		}
		curTag.children = append(curTag.children, LatexNode{
			tag:      content[i:j],
			nodetype: TEXT_NODE,
		})
		i = j
		switch tagType {
		case NEW_TAG:
			// found child tag in current tag
			stack = append(stack, curTag)
			j, tagType = index(content, i)
			if j < 0 || tagType != OPEN_TAG {
				parseError = true
				break
			}
			curTag = LatexNode{
				tag:      content[i:j],
				nodetype: TAG_NODE,
			}
			continue
		case CLOSE_TAG:
			// Closed current tag: recurse on child ta
			parent := stack[len(stack)-1]
			parent.children = append(parent.children, curTag)
			curTag = parent
			if len(stack) == 1 {
				break
			}
			stack = stack[:len(stack)-1]
		default:
			parseError = true
			break
		}
	}
	if parseError {
		log.Println("[mathml.Format] parse error")
		return curTag, errors.New("Parse error")
	}
	return curTag, nil
}

func index(str string, offset int) (int, TagType) {
	for i := offset; i < len(str); i++ {
		switch str[i] {
		case '\\':
			return i, NEW_TAG
		case '{':
			return i, OPEN_TAG
		case '}':
			return i, CLOSE_TAG
		}
	}
	return -1, CLOSE_TAG
}

func convert(tree LatexNode) string {
	return ""
}

package lexer

import "fmt"

type Node struct {
	Self     Token
	Children []*Node
}

func NewNode(token Token) *Node {
	return &Node{Self: token}
}

func (node *Node) FindAncestor(possibleAncestor *Node) *Node {
	for i := len(possibleAncestor.Children) - 1; i >= 0; i-- {
		if possibleAncestor.Children[i].Self.isOneOfKinds(
			HEADING_1,
			HEADING_2,
			HEADING_3,
			HEADING_4,
			HEADING_5,
			// DASH,
			// NUMBERED_LIST,
			// NEWLINE_PARAGRAPH,
		) && possibleAncestor.Children[i].Self.Kind < node.Self.Kind {
			return node.FindAncestor(possibleAncestor.Children[i])
		}
	}

	return possibleAncestor
}

// FIX: Non comparable elements (e.g paragraph) are left out

// NOTE: Perform structuring here
func (node *Node) AddDescendant(newNode *Node) {
	ancestor := newNode.FindAncestor(node)
	ancestor.Children = append(ancestor.Children, newNode)
}

func (node *Node) Display(level int) {
	space := ""
	for i := 0; i <= level; i++ {
		space += "   "
	}

	fmt.Printf("%s", space)
	node.Self.Debug()

	space += "   "

	for _, child := range node.Children {
		fmt.Printf("%s", space)
		child.Display(level + 1)
	}
}

func ParseAST(source string) (*Node, error) {
	sourceNode := NewNode(NewToken(SOURCE_FILE, NewLoc([]int{0, 0}, []int{0, 0})))
	tokens, err := Tokenize(source)
	if err != nil {
		return nil, err
	}

	for _, token := range tokens {
		sourceNode.AddDescendant(NewNode(token))
	}

	return sourceNode, nil
}

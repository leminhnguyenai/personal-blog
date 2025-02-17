package lexer

type Node struct {
	Self     Token
	Values   []*Node
	Children []*Node
}

func NewNode(token Token) *Node {
	return &Node{Self: token}
}

func (node *Node) addChild(newNode *Node) {
	node.Children = append(node.Children, newNode)
}

func (node *Node) addValue(newNode *Node) {
	node.Values = append(node.Values, newNode)
}

// Compare the 2 nodes' priorities
// e.g heading 2 has lower priority than heading 1
func (node *Node) hasLowerPriority(otherNode *Node) bool {
	return node.Self.kind > otherNode.Self.kind
}

// Compare the 2 nodes' indentation size
func (node *Node) indentationDiff(otherNode *Node) int {
	return node.Self.Indentation() - otherNode.Self.Indentation()
}

// Compare the 2 nodes' start vertical position
// If they are the same -> the 1st one is contained within a line
// and the 2nd one is an inline element -> Both are on the same line
func (node *Node) lineDiffStart(otherNode *Node) int {
	return node.Self.loc.start[0] - otherNode.Self.loc.start[0]
}

func (node *Node) lineDiffEnd(otherNode *Node) int {
	return node.Self.loc.start[0] - otherNode.Self.loc.end[0]
}

// Check if the current node is a value of the other node
func (node *Node) isValueOf(otherNode *Node) bool {
	return node.lineDiffStart(otherNode) == 0 &&
		!otherNode.Self.isOneOfKinds(QUOTE) &&
		node.Self.isOneOfKinds(
			PARAGRAPH,
			LINK,
			INLINE_CODE,
		)
}

func (node *Node) isChildOfHeading(otherNode *Node) bool {
	return otherNode.Self.isOneOfKinds(
		HEADING_1,
		HEADING_2,
		HEADING_3,
		HEADING_4,
		HEADING_5,
	) && node.lineDiffStart(otherNode) > 0 &&
		node.hasLowerPriority(otherNode)
}

func (node *Node) isChildOfIndentableToken(otherNode *Node) bool {
	return otherNode.Self.isOneOfKinds(
		DASH,
		NUMBERED_LIST,
		PARAGRAPH,
	) && node.lineDiffStart(otherNode) > 0 &&
		node.indentationDiff(otherNode) > 0
}

func (node *Node) isChildOfQuote(otherNode *Node) bool {
	return otherNode.Self.isOneOfKinds(
		QUOTE,
		CALLOUT_NOTE,
		CALLOUT_IMPORTANT,
		CALLOUT_WARNING,
		CALLOUT_EXAMPLE,
	) &&
		(node.lineDiffStart(otherNode) == 0 || node.lineDiffEnd(otherNode) == 0)
}

// Merge quotes if next to each other
func (node *Node) isQuoteFragment(otherNode *Node) bool {
	return node.Self.isOneOfKinds(QUOTE) &&
		otherNode.Self.isOneOfKinds(
			QUOTE,
			CALLOUT_NOTE,
			CALLOUT_IMPORTANT,
			CALLOUT_WARNING,
			CALLOUT_EXAMPLE,
		) &&
		(node.lineDiffStart(otherNode) == 1 || node.lineDiffEnd(otherNode) == 1) &&
		node.indentationDiff(otherNode) == 0
}

// Find the closest ancestor of the Node using waterfall effect
// The node can either be a value or a child of that ancestor
func (node *Node) findAncestor(possibleAncestor *Node) {
	for i := len(possibleAncestor.Children) - 1; i >= 0; i-- {
		// Comparison for value
		if node.isValueOf(possibleAncestor.Children[i]) {
			possibleAncestor.Children[i].addValue(node)
			return
		}

		// Comparision for Heading token
		if node.isChildOfHeading(possibleAncestor.Children[i]) {
			node.findAncestor(possibleAncestor.Children[i])
			return
		}

		// Comparison for quote
		if node.isQuoteFragment(possibleAncestor.Children[i]) {
			possibleAncestor.Children[i].Self.loc.end = node.Self.loc.end
			return
		}

		if node.isChildOfQuote(possibleAncestor.Children[i]) {
			node.findAncestor(possibleAncestor.Children[i])
			return
		}

		// Comparison for Indentable token
		if node.isChildOfIndentableToken(possibleAncestor.Children[i]) {
			node.findAncestor(possibleAncestor.Children[i])
			return
		}
	}

	possibleAncestor.addChild(node)
}

func (node *Node) Display(str *string, level int) {
	whitespaces := ""
	for i := 0; i < level; i++ {
		whitespaces += " "
	}

	*str += whitespaces + node.Self.Debug() + "\n"

	if len(node.Values) > 0 {
		*str += whitespaces + " " + "values:\n"
	}

	for _, value := range node.Values {
		value.Display(str, level+2)
	}

	if len(node.Children) > 0 {
		*str += whitespaces + " " + "children:\n"
	}

	for _, child := range node.Children {
		child.Display(str, level+2)
	}
}

func ParseAST(source string) (*Node, error) {
	sourceNode := NewNode(NewToken(SOURCE_FILE, NewLoc([]int{-1, -1}, []int{-1, -1})))
	tokens, err := Tokenize(source)
	if err != nil {
		return nil, err
	}

	// Each node will be organized after initialized immediately
	for _, token := range tokens {
		NewNode(token).findAncestor(sourceNode)
	}

	return sourceNode, nil
}

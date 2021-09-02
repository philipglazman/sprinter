package sprinter

import (
	"fmt"
)

// Node represents a web page in an n-ary tree of web pages.
// Contains page links found on the web page, and children of unvisited web pages.
type Node struct {
	// pageLinks is a unique list of links found on a web page.
	pageLinks []string
	// link is the root link of the web page.
	link string
	// child is slice of unvisited web pages.
	child []*Node
}

// NewNode returns a Node.
func NewNode(link string) *Node {
	return &Node{
		link:      link,
		child:     make([]*Node, 0),
		pageLinks: make([]string, 0),
	}
}

// String implements the String interface for Node.
// This pretty prints the root web page visited and the web links found on it.
func (sn *Node) String() string {
	res := fmt.Sprintf("Visited: %s\n", sn.Link())

	for _, page := range sn.pageLinks {
		res += fmt.Sprintf("\t%s\n", page)
	}

	for _, child := range sn.child {
		res += child.String()
	}

	return res
}

// Link returns the root link of the node.
func (sn *Node) Link() string {
	return sn.link
}

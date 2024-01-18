package deptree

type Node struct {
	name     string
	children []*Node
	depth    int
}

func NewNode(name string, children []*Node) *Node {
	return &Node{
		name:     name,
		children: children,
	}
}

func Build() {

}

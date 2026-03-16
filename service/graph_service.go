package service

type GraphNode struct {
	ID       string      `json:"id"`
	Name     string      `json:"name"`
	Children []GraphNode `json:"children"`
}

func BuildGraph(id string, depth int) (*GraphNode, error) {
	person, err := GetPerson(id)
	if err != nil {
		return nil, err
	}

	node := &GraphNode{
		ID:   person.ID,
		Name: person.Name,
	}

	if depth <= 0 {
		return node, nil
	}

	children, err := GetChildren(id)
	if err != nil {
		return node, nil
	}

	for _, c := range children {
		childNode, err := BuildGraph(c.ID, depth-1)
		if err == nil {
			node.Children = append(node.Children, *childNode)
		}
	}

	return node, nil
}
package service

import "family-tree/model"

func GetTree(id string) (*model.TreeNode, error) {
	person, err := GetPerson(id)
	if err != nil {
		return nil, err
	}

	node := &model.TreeNode{Person: person}

	if person.FatherID != "" {
		if father, err := GetPerson(person.FatherID); err == nil {
			node.Parents = append(node.Parents, *father)
		}
	}
	if person.MotherID != "" {
		if mother, err := GetPerson(person.MotherID); err == nil {
			node.Parents = append(node.Parents, *mother)
		}
	}

	if children, err := GetChildren(id); err == nil {
		node.Children = children
	}

	if spouses, err := GetSpouses(id); err == nil {
		node.Spouses = spouses
	}

	return node, nil
}
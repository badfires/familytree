package service

import "family-tree/model"

func BuildFamilyGraph(id string, depth int) (*model.FamilyGraphNode, error) {
	person, err := GetPerson(id)
	if err != nil {
		return nil, err
	}

	personNode := &model.FamilyGraphNode{
		ID:   person.ID,
		Name: person.Name,
		Type: "person",
	}

	if depth <= 0 {
		return personNode, nil
	}

	marriages, err := GetMarriagesByPersonID(id)
	if err != nil || len(marriages) == 0 {
		children, _ := GetChildren(id)
		for _, c := range children {
			childNode, err := BuildFamilyGraph(c.ID, depth-1)
			if err == nil {
				personNode.Children = append(personNode.Children, *childNode)
			}
		}
		return personNode, nil
	}

	for _, m := range marriages {
		mNode := model.FamilyGraphNode{
			ID:   m.ID,
			Type: "marriage",
		}

		if m.HusbandID != "" {
			if h, err := GetPerson(m.HusbandID); err == nil {
				mNode.Spouses = append(mNode.Spouses, model.SimplePersonNode{ID: h.ID, Name: h.Name})
			}
		}
		if m.WifeID != "" {
			if w, err := GetPerson(m.WifeID); err == nil {
				mNode.Spouses = append(mNode.Spouses, model.SimplePersonNode{ID: w.ID, Name: w.Name})
			}
		}

		children, err := GetMarriageChildren(m.ID)
		if err == nil {
			for _, c := range children {
				childNode, err := BuildFamilyGraph(c.ID, depth-1)
				if err == nil {
					mNode.Children = append(mNode.Children, *childNode)
				}
			}
		}

		personNode.Children = append(personNode.Children, mNode)
	}

	return personNode, nil
}
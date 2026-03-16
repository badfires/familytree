package service

import "family-tree/model"

func BuildFamilyView(id string) (*model.FamilyView, error) {
	person, err := GetPerson(id)
	if err != nil {
		return nil, err
	}

	view := &model.FamilyView{
		Center: toViewPerson(person),
	}

	if person.FatherID != "" {
		if p, err := GetPerson(person.FatherID); err == nil {
			view.Parents = append(view.Parents, toViewPerson(p))
		}
	}
	if person.MotherID != "" {
		if p, err := GetPerson(person.MotherID); err == nil {
			view.Parents = append(view.Parents, toViewPerson(p))
		}
	}

	marriages, err := GetMarriagesByPersonID(id)
	if err == nil {
		for _, m := range marriages {
			vm := model.ViewMarriageNode{MarriageID: m.ID}

			if m.HusbandID != "" {
				if p, err := GetPerson(m.HusbandID); err == nil {
					vm.Spouses = append(vm.Spouses, toViewPerson(p))
				}
			}
			if m.WifeID != "" {
				if p, err := GetPerson(m.WifeID); err == nil {
					dup := false
					for _, s := range vm.Spouses {
						if s.ID == p.ID {
							dup = true
							break
						}
					}
					if !dup {
						vm.Spouses = append(vm.Spouses, toViewPerson(p))
					}
				}
			}

			children, err := GetMarriageChildren(m.ID)
			if err == nil {
				for _, c := range children {
					cp := c
					vm.Children = append(vm.Children, toViewPerson(&cp))
				}
			}

			view.Marriages = append(view.Marriages, vm)
		}
	}

	adoption, err := GetAdoption(id)
	if err == nil && adoption != nil {
		av := &model.AdoptionView{
			PersonID: adoption.PersonID,
			Note:     adoption.Note,
		}

		if adoption.FromFatherID != "" {
			if p, err := GetPerson(adoption.FromFatherID); err == nil {
				av.From = append(av.From, toViewPerson(p))
			}
		}
		if adoption.FromMotherID != "" {
			if p, err := GetPerson(adoption.FromMotherID); err == nil {
				av.From = append(av.From, toViewPerson(p))
			}
		}
		if adoption.ToFatherID != "" {
			if p, err := GetPerson(adoption.ToFatherID); err == nil {
				av.To = append(av.To, toViewPerson(p))
			}
		}
		if adoption.ToMotherID != "" {
			if p, err := GetPerson(adoption.ToMotherID); err == nil {
				av.To = append(av.To, toViewPerson(p))
			}
		}

		view.Adoption = av
	}

	return view, nil
}
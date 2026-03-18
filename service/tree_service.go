package service

import (
	"strings"

	"family-tree/model"
)

type treeBuildContext struct {
	visitedPersons    map[string]bool
	renderedAdoptions map[string]bool
	extraLinks        []model.TreeLink
}

func GetTree(id string) (*model.TreeResponse, error) {
	ctx := &treeBuildContext{
		visitedPersons:    map[string]bool{},
		renderedAdoptions: map[string]bool{},
		extraLinks:        make([]model.TreeLink, 0),
	}

	root, err := buildTreeRecursive(id, ctx)
	if err != nil {
		return nil, err
	}

	return &model.TreeResponse{
		Root:       root,
		ExtraLinks: ctx.extraLinks,
	}, nil
}

func buildPersonLabel(p model.ViewPerson) string {
	if strings.TrimSpace(p.Name) != "" {
		return p.Name
	}
	return p.ID
}

func buildTreeRecursive(id string, ctx *treeBuildContext) (*model.TreeNode, error) {
	if strings.TrimSpace(id) == "" {
		return nil, nil
	}
	if ctx.visitedPersons[id] {
		return nil, nil
	}
	ctx.visitedPersons[id] = true

	view, err := BuildFamilyView(id)
	if err != nil {
		return nil, err
	}

	root := &model.TreeNode{
		ID:    view.Center.ID,
		Label: buildPersonLabel(view.Center),
		Title: view.Center.ID,
		Type:  "person",
	}

	coveredPairs := map[string]bool{}

	// 1) 收集所有正式婚姻配偶，用于同排展示
	for _, m := range view.Marriages {
		var spouseID string
		var spouseLabel string

		for _, s := range m.Spouses {
			if s.ID == view.Center.ID {
				continue
			}
			spouseID = s.ID
			spouseLabel = buildPersonLabel(s)

			if !hasSpouse(root.Spouses, s.ID) {
				root.Spouses = append(root.Spouses, model.TreeSpouse{
					ID:    s.ID,
					Label: spouseLabel,
					Title: s.ID,
				})
			}
			break
		}

		pairKey := normalizePair(view.Center.ID, spouseID)
		coveredPairs[pairKey] = true

		family := model.TreeFamily{
			Key:      pairKey,
			SpouseID: spouseID,
			Children: []*model.TreeNode{},
		}

		// 正式婚姻的孩子
		for _, c := range m.Children {
			a, err := GetAdoption(c.ID)
			if err == nil && a != nil {
				if pairContainsParentIDs(pairKey, a.FromFatherID, a.FromMotherID) {
					// 已从这对原父母过继出去，不再挂这里
					continue
				}
			}

			childNode, err := buildTreeRecursive(c.ID, ctx)
			if err != nil || childNode == nil {
				continue
			}
			family.Children = append(family.Children, childNode)
		}

		// 过继到这对夫妻名下的孩子
		adoptedChildren, err := GetAdoptionsByAdoptivePair(view.Center.ID, spouseID)
		if err == nil {
			for _, a := range adoptedChildren {
				if ctx.renderedAdoptions[a.PersonID] {
					continue
				}
				childNode, err := buildTreeRecursive(a.PersonID, ctx)
				if err != nil || childNode == nil {
					continue
				}
				family.Children = append(family.Children, childNode)
				ctx.renderedAdoptions[a.PersonID] = true
				appendAdoptionOriginLinks(ctx, a)
			}
		}

		adoptedChildren2, err := GetAdoptionsByAdoptivePair(spouseID, view.Center.ID)
		if err == nil {
			for _, a := range adoptedChildren2 {
				if ctx.renderedAdoptions[a.PersonID] {
					continue
				}
				childNode, err := buildTreeRecursive(a.PersonID, ctx)
				if err != nil || childNode == nil {
					continue
				}
				family.Children = append(family.Children, childNode)
				ctx.renderedAdoptions[a.PersonID] = true
				appendAdoptionOriginLinks(ctx, a)
			}
		}

		if len(family.Children) > 0 {
			root.Families = append(root.Families, family)
		}
	}

	// 2) 无婚姻记录的过继家庭
	adoptions, err := GetAdoptionsByAdoptiveParent(view.Center.ID)
	if err == nil && len(adoptions) > 0 {
		familyMap := map[string]*model.TreeFamily{}

		for _, a := range adoptions {
			if ctx.renderedAdoptions[a.PersonID] {
				continue
			}

			otherParentID := getOtherAdoptiveParentID(view.Center.ID, a.ToFatherID, a.ToMotherID)
			pairKey := normalizePair(view.Center.ID, otherParentID)

			if coveredPairs[pairKey] {
				continue
			}

			if strings.TrimSpace(otherParentID) != "" && !hasSpouse(root.Spouses, otherParentID) {
				if p, _ := GetPerson(otherParentID); p != nil {
					root.Spouses = append(root.Spouses, model.TreeSpouse{
						ID:    p.ID,
						Label: firstNonEmpty(p.Name, p.ID),
						Title: p.ID,
					})
				} else {
					root.Spouses = append(root.Spouses, model.TreeSpouse{
						ID:    otherParentID,
						Label: otherParentID,
						Title: otherParentID,
					})
				}
			}

			f, ok := familyMap[pairKey]
			if !ok {
				f = &model.TreeFamily{
					Key:      pairKey,
					SpouseID: otherParentID,
					Children: []*model.TreeNode{},
				}
				familyMap[pairKey] = f
			}

			childNode, err := buildTreeRecursive(a.PersonID, ctx)
			if err != nil || childNode == nil {
				continue
			}
			f.Children = append(f.Children, childNode)
			ctx.renderedAdoptions[a.PersonID] = true
			appendAdoptionOriginLinks(ctx, a)
		}

		for _, f := range familyMap {
			if len(f.Children) > 0 {
				root.Families = append(root.Families, *f)
			}
		}
	}

	return root, nil
}

func appendAdoptionOriginLinks(ctx *treeBuildContext, a model.Adoption) {
	if strings.TrimSpace(a.PersonID) == "" {
		return
	}
	if strings.TrimSpace(a.FromFatherID) != "" {
		ctx.extraLinks = append(ctx.extraLinks, model.TreeLink{
			SourceID: a.PersonID,
			TargetID: a.FromFatherID,
			Style:    "dashed",
			Kind:     "adoption_from",
		})
	}
	if strings.TrimSpace(a.FromMotherID) != "" {
		ctx.extraLinks = append(ctx.extraLinks, model.TreeLink{
			SourceID: a.PersonID,
			TargetID: a.FromMotherID,
			Style:    "dashed",
			Kind:     "adoption_from",
		})
	}
}

func normalizePair(a, b string) string {
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)
	if a == "" && b == "" {
		return ""
	}
	if a < b {
		return a + "|" + b
	}
	return b + "|" + a
}

func pairContainsParentIDs(pairKey, fatherID, motherID string) bool {
	return pairKey == normalizePair(fatherID, motherID)
}

func getOtherAdoptiveParentID(centerID, fatherID, motherID string) string {
	centerID = strings.TrimSpace(centerID)
	fatherID = strings.TrimSpace(fatherID)
	motherID = strings.TrimSpace(motherID)

	if fatherID == centerID && motherID != "" {
		return motherID
	}
	if motherID == centerID && fatherID != "" {
		return fatherID
	}
	return ""
}

func hasSpouse(spouses []model.TreeSpouse, id string) bool {
	id = strings.TrimSpace(id)
	for _, s := range spouses {
		if strings.TrimSpace(s.ID) == id {
			return true
		}
	}
	return false
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
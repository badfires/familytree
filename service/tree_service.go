package service

import (
	"strings"

	"family-tree/model"
)

type treeBuildContext struct {
	renderedAdoptions map[string]bool
	extraLinks        []model.TreeLink
}

func GetTree(id string) (*model.TreeResponse, error) {
	ctx := &treeBuildContext{
		renderedAdoptions: map[string]bool{},
		extraLinks:        make([]model.TreeLink, 0),
	}

	root, err := buildTreeRecursive(id, ctx, map[string]bool{})
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

func cloneVisitPath(src map[string]bool) map[string]bool {
	dst := make(map[string]bool, len(src)+1)
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func buildTreeRecursive(id string, ctx *treeBuildContext, visitPath map[string]bool) (*model.TreeNode, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, nil
	}

	if visitPath[id] {
		return nil, nil
	}

	currentPath := cloneVisitPath(visitPath)
	currentPath[id] = true

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

	// 1) 正式婚姻：收集 spouses + 为每个 spouse 建 family
	for _, m := range view.Marriages {
		var spouseID string
		for _, s := range m.Spouses {
			if s.ID == view.Center.ID {
				continue
			}
			spouseID = s.ID
			if !hasSpouse(root.Spouses, s.ID) {
				root.Spouses = append(root.Spouses, model.TreeSpouse{
					ID:    s.ID,
					Label: buildPersonLabel(s),
					Title: s.ID,
				})
			}
			break
		}

		pairKey := normalizePair(view.Center.ID, spouseID)
		coveredPairs[pairKey] = true

		family := model.TreeFamily{
			Key:        pairKey,
			SpouseID:   spouseID,
			FamilyType: "marriage",
			Children:   []*model.TreeNode{},
		}

		// 婚生孩子
		for _, c := range m.Children {
			a, err := GetAdoption(c.ID)
			if err == nil && a != nil && pairContainsParentIDs(pairKey, a.FromFatherID, a.FromMotherID) {
				// 已从这对父母过继走，不在原婚姻下画实线
				continue
			}

			childNode, err := buildTreeRecursive(c.ID, ctx, currentPath)
			if err != nil || childNode == nil {
				continue
			}
			family.Children = append(family.Children, childNode)
		}

		// 过继到这对配偶名下的孩子
		adopted1, err := GetAdoptionsByAdoptivePair(view.Center.ID, spouseID)
		if err == nil {
			for _, a := range adopted1 {
				if ctx.renderedAdoptions[a.PersonID] {
					continue
				}
				childNode, err := buildTreeRecursive(a.PersonID, ctx, currentPath)
				if err != nil || childNode == nil {
					continue
				}
				family.Children = append(family.Children, childNode)
				ctx.renderedAdoptions[a.PersonID] = true
				appendAdoptionOriginLinks(ctx, a)
			}
		}

		adopted2, err := GetAdoptionsByAdoptivePair(spouseID, view.Center.ID)
		if err == nil {
			for _, a := range adopted2 {
				if ctx.renderedAdoptions[a.PersonID] {
					continue
				}
				childNode, err := buildTreeRecursive(a.PersonID, ctx, currentPath)
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

	// 2) 没婚姻记录的过继家庭
	adoptions, err := GetAdoptionsByAdoptiveParent(view.Center.ID)
	if err == nil && len(adoptions) > 0 {
		familyMap := map[string]*model.TreeFamily{}

		for _, a := range adoptions {
			if ctx.renderedAdoptions[a.PersonID] {
				continue
			}

			otherParentID := getOtherAdoptiveParentID(view.Center.ID, a.ToFatherID, a.ToMotherID)
			pairKey := normalizePair(view.Center.ID, otherParentID)

			// 正式婚姻已经处理过，则跳过
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
				familyType := "single_parent"
				if strings.TrimSpace(otherParentID) != "" {
					familyType = "adoption"
				}
				f = &model.TreeFamily{
					Key:        pairKey,
					SpouseID:   otherParentID,
					FamilyType: familyType,
					Children:   []*model.TreeNode{},
				}
				familyMap[pairKey] = f
			}

			childNode, err := buildTreeRecursive(a.PersonID, ctx, currentPath)
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
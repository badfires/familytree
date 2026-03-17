package service

import (
	"family-tree/model"
	"strings"
)

func GetTree(id string) (*model.TreeNode, error) {
    return buildTreeRecursive(id, map[string]bool{})
}

func buildPersonLabel(p model.ViewPerson) string {
	if p.Name != "" && p.ID != "" {
		return p.Name + " (" + p.ID + ")"
	}
	if p.Name != "" {
		return p.Name
	}
	return p.ID
}

func buildMarriageLabel(centerID string, m model.ViewMarriageNode) string {
	var spouseNames []string
	for _, s := range m.Spouses {
		if s.ID == centerID {
			continue
		}
		if s.Name != "" {
			spouseNames = append(spouseNames, s.Name)
		} else if s.ID != "" {
			spouseNames = append(spouseNames, s.ID)
		}
	}

	if len(spouseNames) == 0 {
		
		return "配偶"
	}

	return "配偶：" + strings.Join(spouseNames, "、")
}
func buildTreeRecursive(id string, visited map[string]bool) (*model.TreeNode, error) {
    // 防止死循环（非常重要）
    if visited[id] {
        return nil, nil
    }
    visited[id] = true

    view, err := BuildFamilyView(id)
    if err != nil {
        return nil, err
    }

    root := &model.TreeNode{
        ID:    view.Center.ID,
        Label: buildPersonLabel(view.Center),
        Type:  "person",
    }

    for _, m := range view.Marriages {
        marriageNode := &model.TreeNode{
            ID:    m.MarriageID,
            Label: buildMarriageLabel(view.Center.ID, m),
            Type:  "marriage",
        }

        // 子女递归
        for _, c := range m.Children {
            childNode, err := buildTreeRecursive(c.ID, visited)
            if err != nil {
                continue
            }
            if childNode != nil {
                marriageNode.Children = append(marriageNode.Children, childNode)
            }
        }

        root.Children = append(root.Children, marriageNode)
    }

    return root, nil
}
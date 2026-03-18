package model

type TreeSpouse struct {
	ID    string `json:"id,omitempty"`
	Label string `json:"label"`
	Title string `json:"title,omitempty"`
}

type TreeFamily struct {
	Key        string      `json:"key"`
	SpouseID   string      `json:"spouse_id,omitempty"`
	FamilyType string      `json:"family_type,omitempty"` // marriage / adoption / single_parent
	Children   []*TreeNode `json:"children,omitempty"`
}

type TreeNode struct {
	ID       string       `json:"id,omitempty"`
	Label    string       `json:"label"`
	Type     string       `json:"type"`
	Title    string       `json:"title,omitempty"`
	Spouses  []TreeSpouse `json:"spouses,omitempty"`
	Families []TreeFamily `json:"families,omitempty"`
}

type TreeLink struct {
	SourceID string `json:"source_id"`
	TargetID string `json:"target_id"`
	Style    string `json:"style"`
	Kind     string `json:"kind"`
}

type TreeResponse struct {
	Root       *TreeNode  `json:"root"`
	ExtraLinks []TreeLink `json:"extra_links,omitempty"`
}
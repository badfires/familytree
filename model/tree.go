package model

type TreeNode struct {
	ID       string      `json:"id,omitempty"`
	Label    string      `json:"label"`
	Type     string      `json:"type"`
	Children []*TreeNode `json:"children,omitempty"`
}
package model

type TreeNode struct {
	Person   *Person  `json:"person"`
	Parents  []Person `json:"parents"`
	Spouses  []Person `json:"spouses"`
	Children []Person `json:"children"`
}
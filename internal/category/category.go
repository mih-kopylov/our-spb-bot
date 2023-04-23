package category

import (
	_ "embed"
	"github.com/joomcode/errorx"
	"strings"
)

//go:embed defaultCategories.yaml
var DefaultCategoriesText []byte

var (
	Errors                 = errorx.NewNamespace("Category")
	ErrMalformedCategories = Errors.NewType("MalformedCategories")
)

type UserCategory struct {
	Id      int64
	Message string
}

type UserCategoryTreeNode struct {
	Id       string
	Name     string
	Category *UserCategory
	Parent   *UserCategoryTreeNode
	Children []*UserCategoryTreeNode
}

func (n *UserCategoryTreeNode) GetFullName() string {
	var names []string
	node := n
	for {
		names = append([]string{node.Name}, names...)
		if node.Parent != nil && node.Parent.Parent != nil {
			node = node.Parent
		} else {
			break
		}
	}
	return strings.Join(names, " / ")
}

func (n *UserCategoryTreeNode) FindNodeById(id string) *UserCategoryTreeNode {
	if n.Id == id {
		return n
	}
	for _, child := range n.Children {
		result := child.FindNodeById(id)
		if result != nil {
			return result
		}
	}
	return nil
}

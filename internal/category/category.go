package category

import (
	"crypto/md5"
	_ "embed"
	"encoding/base64"
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
	Name     string
	Category *UserCategory
	Parent   *UserCategoryTreeNode
	Children []*UserCategoryTreeNode
}

func (n *UserCategoryTreeNode) Id() string {
	rawId := n.GetFullName()
	if rawId == "" {
		return ""
	}
	hash := md5.Sum([]byte(rawId))
	result := base64.URLEncoding.EncodeToString(hash[:])
	return result
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
	if n.Id() == id {
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

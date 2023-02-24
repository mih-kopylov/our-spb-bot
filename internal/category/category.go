package category

import (
	_ "embed"
	"github.com/google/uuid"
	"github.com/joomcode/errorx"
	"gopkg.in/yaml.v3"
	"strings"
)

//go:embed categories.yaml
var categoriesText []byte

type UserCategories map[string]map[string]UserCategory

type UserCategory struct {
	Id      int
	Message string
}

type UserCategoryTreeNode struct {
	Id       string
	Name     string
	Message  string
	Category *UserCategory
	Children []*UserCategoryTreeNode
	Parent   *UserCategoryTreeNode
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

func getUserCategories() (UserCategories, error) {
	var result UserCategories
	err := yaml.Unmarshal(categoriesText, &result)
	if err != nil {
		return nil, errorx.EnhanceStackTrace(err, "failed to unmarshall user categories file")
	}

	return result, nil
}

func CreateUserCategoryTree() (*UserCategoryTreeNode, error) {
	categories, err := getUserCategories()
	if err != nil {
		return nil, err
	}

	rootNode := UserCategoryTreeNode{
		Id:   uuid.NewString(),
		Name: "Корень",
	}

	for groupName, group := range categories {
		groupNode := UserCategoryTreeNode{
			Id:       uuid.NewString(),
			Name:     groupName,
			Message:  "",
			Category: nil,
			Children: []*UserCategoryTreeNode{},
			Parent:   &rootNode,
		}
		rootNode.Children = append(rootNode.Children, &groupNode)
		for categoryName, userCategory := range group {
			categoryNode := UserCategoryTreeNode{
				Id:       uuid.NewString(),
				Name:     categoryName,
				Message:  userCategory.Message,
				Category: &userCategory,
				Children: nil,
				Parent:   &groupNode,
			}
			groupNode.Children = append(groupNode.Children, &categoryNode)
		}
	}

	return &rootNode, nil
}

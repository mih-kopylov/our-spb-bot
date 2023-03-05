package category

import (
	_ "embed"
	"github.com/goioc/di"
	"github.com/joomcode/errorx"
	"github.com/lithammer/shortuuid/v4"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
	"strings"
)

//go:embed categories.yaml
var categoriesText []byte

const (
	BeanId = "Categories"
)

func RegisterBean() {
	tree := lo.Must(createUserCategoryTree())
	_ = lo.Must(di.RegisterBeanInstance(BeanId, tree))
}

type userCategories map[string]map[string]UserCategory

type UserCategory struct {
	Id      int64
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

func getUserCategories() (userCategories, error) {
	var result userCategories
	err := yaml.Unmarshal(categoriesText, &result)
	if err != nil {
		return nil, errorx.EnhanceStackTrace(err, "failed to unmarshall user categories file")
	}

	return result, nil
}

func createUserCategoryTree() (*UserCategoryTreeNode, error) {
	categories, err := getUserCategories()
	if err != nil {
		return nil, err
	}

	rootNode := UserCategoryTreeNode{
		Id:   "",
		Name: "Корень",
	}

	for groupName, group := range categories {
		groupNode := UserCategoryTreeNode{
			Id:       shortuuid.New(),
			Name:     groupName,
			Message:  "",
			Category: nil,
			Children: []*UserCategoryTreeNode{},
			Parent:   &rootNode,
		}
		rootNode.Children = append(rootNode.Children, &groupNode)
		for categoryName, userCategory := range group {
			userCategory := userCategory
			categoryNode := UserCategoryTreeNode{
				Id:       shortuuid.New(),
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

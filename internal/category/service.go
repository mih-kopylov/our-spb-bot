package category

import (
	"strconv"

	"github.com/joomcode/errorx"
	"gopkg.in/yaml.v3"
)

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

func (s Service) ParseCategoriesTree(value string) (*UserCategoryTreeNode, error) {
	return createUserCategoryTree(value)
}

func createUserCategoryTree(categoriesString string) (*UserCategoryTreeNode, error) {
	var categoriesDocumentNode yaml.Node
	err := yaml.Unmarshal([]byte(categoriesString), &categoriesDocumentNode)
	if err != nil {
		return nil, errorx.EnhanceStackTrace(err, "failed to unmarshall user categories")
	}

	if categoriesDocumentNode.Kind != yaml.DocumentNode {
		return nil, ErrMalformedCategories.New("root node is expected to be document but was %v, position:%v-%v", categoriesDocumentNode.Kind, categoriesDocumentNode.Line, categoriesDocumentNode.Column)
	}

	if len(categoriesDocumentNode.Content) != 1 {
		return nil, ErrMalformedCategories.New("a single element is expected in yaml document, but was %v, position:%v-%v", len(categoriesDocumentNode.Content), categoriesDocumentNode.Line, categoriesDocumentNode.Column)
	}

	categoriesNode := categoriesDocumentNode.Content[0]
	if categoriesNode.Kind != yaml.MappingNode {
		return nil, ErrMalformedCategories.New("type of yaml document is expected to be a mapping node, but was %v, position:%v-%v", categoriesNode.Kind, categoriesNode.Line, categoriesNode.Column)
	}

	rootNode := &UserCategoryTreeNode{
		Name: "",
	}

	err = parseMapNode(categoriesNode, rootNode)
	if err != nil {
		return nil, err
	}

	return rootNode, nil
}

func parseMapNode(yamlMapNode *yaml.Node, treeNode *UserCategoryTreeNode) error {
	if yamlMapNode.Kind != yaml.MappingNode {
		return ErrMalformedCategories.New("map node expected, but was %v, position:%v-%v", yamlMapNode.Kind, yamlMapNode.Line, yamlMapNode.Column)
	}

	if len(yamlMapNode.Content)%2 != 0 {
		return ErrMalformedCategories.New("map node is expected to have even number of content elemenets, but was odd, position:%v-%v", yamlMapNode.Line, yamlMapNode.Column)
	}

	for index := 0; index < len(yamlMapNode.Content); index += 2 {
		keyNode := yamlMapNode.Content[index]
		if keyNode.Kind != yaml.ScalarNode {
			return ErrMalformedCategories.New("key node is expected to be scalar, but was %v, position:%v-%v", keyNode.Kind, keyNode.Line, keyNode.Column)
		}

		valueNode := yamlMapNode.Content[index+1]
		childNode, err := parseChildTreeNode(valueNode, keyNode.Value, treeNode)
		if err != nil {
			return err
		}
		treeNode.Children = append(treeNode.Children, childNode)
	}

	return nil
}

func parseChildTreeNode(yamlNode *yaml.Node, name string, parent *UserCategoryTreeNode) (*UserCategoryTreeNode, error) {
	if yamlNode.Kind != yaml.MappingNode {
		return nil, ErrMalformedCategories.New("map node expected, but was %v, position:%v-%v", yamlNode.Kind, yamlNode.Line, yamlNode.Column)
	}

	if len(yamlNode.Content) == 4 && yamlNode.Content[0].Value == "id" && yamlNode.Content[2].Value == "message" {
		idString := yamlNode.Content[1].Value
		id, err := strconv.ParseInt(idString, 10, 64)
		if err != nil {
			return nil, ErrMalformedCategories.Wrap(err, "failed to parse category id: %v", idString)
		}

		message := yamlNode.Content[3].Value

		return &UserCategoryTreeNode{
			Name:     name,
			Category: &UserCategory{Id: id, Message: message},
			Parent:   parent,
			Children: nil,
		}, nil
	}

	result := &UserCategoryTreeNode{
		Name:     name,
		Category: nil,
		Parent:   parent,
		Children: nil,
	}
	err := parseMapNode(yamlNode, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

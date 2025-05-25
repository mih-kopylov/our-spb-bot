package category

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateUserCategoryTree(t *testing.T) {
	tests := []struct {
		name          string
		sourceFile    string
		expected      *UserCategoryTreeNode
		errorExpected bool
	}{
		{
			name:          "empty",
			sourceFile:    "testdata/empty.yaml",
			expected:      nil,
			errorExpected: true,
		},
		{
			name:       "single",
			sourceFile: "testdata/single.yaml",
			expected: &UserCategoryTreeNode{
				Name:     "",
				Category: nil,
				Parent:   nil,
				Children: []*UserCategoryTreeNode{{
					Name: "Category 1",
					Category: &UserCategory{
						Id:      1,
						Message: "message 1",
					},
					Parent:   nil,
					Children: nil,
				}},
			},
			errorExpected: false,
		},
		{
			name:       "multiple",
			sourceFile: "testdata/multiple.yaml",
			expected: &UserCategoryTreeNode{
				Name:     "",
				Category: nil,
				Parent:   nil,
				Children: []*UserCategoryTreeNode{{
					Name: "Category 1",
					Category: &UserCategory{
						Id:      1,
						Message: "message 1",
					},
					Parent:   nil,
					Children: nil,
				}, {
					Name: "Category 2",
					Category: &UserCategory{
						Id:      2,
						Message: "message 2",
					},
					Parent:   nil,
					Children: nil,
				}},
			},
			errorExpected: false,
		},
		{
			name:       "group",
			sourceFile: "testdata/group.yaml",
			expected: &UserCategoryTreeNode{
				Name:     "",
				Category: nil,
				Parent:   nil,
				Children: []*UserCategoryTreeNode{{
					Name:     "Group 1",
					Category: nil,
					Parent:   nil,
					Children: []*UserCategoryTreeNode{{
						Name: "Category 1",
						Category: &UserCategory{
							Id:      1,
							Message: "message 1",
						},
						Parent:   nil,
						Children: nil,
					}},
				}},
			},
			errorExpected: false,
		},
		{
			name:       "nestedGroup",
			sourceFile: "testdata/nestedGroup.yaml",
			expected: &UserCategoryTreeNode{
				Name:     "",
				Category: nil,
				Parent:   nil,
				Children: []*UserCategoryTreeNode{{
					Name:     "Group 1",
					Category: nil,
					Parent:   nil,
					Children: []*UserCategoryTreeNode{{
						Name:     "Group 11",
						Category: nil,
						Parent:   nil,
						Children: []*UserCategoryTreeNode{{
							Name: "Category 1",
							Category: &UserCategory{
								Id:      1,
								Message: "message 1",
							},
							Parent:   nil,
							Children: nil,
						}},
					}},
				}},
			},
			errorExpected: false,
		},
		{
			name:       "combined",
			sourceFile: "testdata/combined.yaml",
			expected: &UserCategoryTreeNode{
				Name:     "",
				Category: nil,
				Parent:   nil,
				Children: []*UserCategoryTreeNode{{
					Name: "Category 1",
					Category: &UserCategory{
						Id:      1,
						Message: "message 1",
					},
					Parent:   nil,
					Children: nil,
				}, {
					Name:     "Group 1",
					Category: nil,
					Parent:   nil,
					Children: []*UserCategoryTreeNode{{
						Name: "Category 2",
						Category: &UserCategory{
							Id:      2,
							Message: "message 2",
						},
						Parent:   nil,
						Children: nil,
					}},
				}},
			},
			errorExpected: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bytes, err := os.ReadFile(test.sourceFile)
			if !assert.NoError(t, err) {
				return
			}

			actual, err := createUserCategoryTree(string(bytes))
			if test.errorExpected {
				assert.Error(t, err)
			} else {
				assert.True(t, test.expected.equals(actual))
			}
		})
	}
}

func (n *UserCategoryTreeNode) equals(other *UserCategoryTreeNode) bool {
	result := n.Name == other.Name

	result = result && compareStructPointers(n.Category, other.Category, func(left *UserCategory, right *UserCategory) bool {
		return *left == *right
	})

	result = result && len(n.Children) == len(other.Children)

	for i, child := range n.Children {
		result = result && child.equals(other.Children[i])
	}

	return result
}

func compareStructPointers[T any](left *T, right *T, compareWith func(left *T, right *T) bool) bool {
	if left != nil && right != nil {
		return compareWith(left, right)
	}

	return left == nil && right == nil
}

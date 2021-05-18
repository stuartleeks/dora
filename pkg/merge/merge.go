package merge

import (
	"fmt"

	"github.com/bradford-hamilton/dora/pkg/ast"
)

func MergeJSON(baseDocument ast.RootNode, mergeDocument ast.RootNode) (*ast.RootNode, error) {

	result := baseDocument

	newContent, err := mergeValueContent(result.RootValue.Content, mergeDocument.RootValue.Content, "$")
	if err != nil {
		return nil, err
	}
	result.RootValue.Content = newContent
	return &result, nil
}

func mergeValueContent(baseValue ast.ValueContent, mergeValue ast.ValueContent, currentPath string) (ast.ValueContent, error) {

	switch baseContent := (baseValue).(type) {
	case ast.Object:
		switch mergeContent := mergeValue.(type) {
		case ast.Object:
			for _, mergeChild := range mergeContent.Children {
				baseChild := getChildByKey(baseContent, mergeChild.Key.Value)
				if baseChild == nil {
					lastChildIndex := len(baseContent.Children) - 1
					if baseContent.Children[lastChildIndex].HasCommaSeparator {
						baseContent.SuffixStructure = append(stripWhiteSpace(baseContent.SuffixStructure), mergeContent.SuffixStructure...)
					} else {
						// Add in comma
						baseContent.Children[lastChildIndex].HasCommaSeparator = true
						baseContent.Children[lastChildIndex].Value.SuffixStructure = stripWhiteSpace(baseContent.Children[lastChildIndex].Value.SuffixStructure)
						if mergeChild.HasCommaSeparator {
							baseContent.SuffixStructure = append(stripWhiteSpace(baseContent.SuffixStructure), mergeContent.SuffixStructure...)
						}
					}
					baseContent.Children = append(baseContent.Children, mergeChild)
				} else {
					// TODO - handle merging object properties
				}
			}
			return baseContent, nil
		default:
			return nil, fmt.Errorf("mis-matched types at %q. base type: %T, merge type: %T", currentPath, baseContent, mergeContent)
		}
	default:
		return nil, fmt.Errorf("unhandled type at %q. base type: %T", currentPath, baseContent)
	}
}

func getChildByKey(object ast.Object, key string) *ast.Property {
	for _, child := range object.Children {
		if child.Key.Value == key {
			return &child
		}
	}
	return nil
}

func stripWhiteSpace(structuralItems []ast.StructuralItem) []ast.StructuralItem {
	var lastNonWhitespaceIndex int
	for lastNonWhitespaceIndex := len(structuralItems) - 1; lastNonWhitespaceIndex >= 0; lastNonWhitespaceIndex-- {
		if structuralItems[lastNonWhitespaceIndex].ItemType != ast.WhitespaceStructuralItemType {
			break
		}
	}
	return structuralItems[0:lastNonWhitespaceIndex]
}

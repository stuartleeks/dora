package dora

import (
	"errors"
	"fmt"

	"github.com/bradford-hamilton/dora/pkg/ast"
	"github.com/bradford-hamilton/dora/pkg/danger"
)

var (
	// ErrNoDollarSignRoot is used for telling the user the very first character must be a `$`
	ErrNoDollarSignRoot = errors.New(
		"Incorrect syntax, query must start with `$` representing the root object or array",
	)
	// ErrWrongObjectRootSelector is used for telling the user their JSON root is an object and the selector found was not a `.`
	ErrWrongObjectRootSelector = errors.New(
		"Incorrect syntax. Your root JSON type is an object. Therefore, path queries must" +
			"begin by selecting a `key` from your root object. Ex: `$.keyOnRootObject` or `$[\"keyOnRootObject\"]`",
	)
	// ErrWrongArrayRootSelector is used for telling the user their JSON root is an array and the selector found was not a `[`
	ErrWrongArrayRootSelector = errors.New(
		"Incorrect syntax. Your root JSON type is an array. Therefore, path queries must" +
			"begin by selecting an item by index on the root array. Ex: `$[0]` or `$[1]`",
	)
)

// prepAndExecQuery prepares and executes a passed in query
func (c *Client) prepAndExecQuery(query string) error {
	if err := c.prepareQuery(query, c.tree.Type); err != nil {
		return err
	}
	if err := c.executeQuery(); err != nil {
		return err
	}
	return nil
}

// prepareQuery validates the query root, sets the query on the client struct, and parses the query.
func (c *Client) prepareQuery(query string, rootNodeType ast.RootNodeType) error {
	if err := validateQueryRoot(query, c.tree.Type); err != nil {
		return err
	}
	c.setQuery(danger.StringToBytes(query))
	if err := c.parseQuery(); err != nil {
		return err
	}
	return nil
}

func (c *Client) setQuery(query []byte) {
	c.query = query
}

// parseQuery is pretty straight forward. Scan the query into tokens, set the the tokns
// to the `parsedQuery` field on the client.
func (c *Client) parseQuery() error {
	tokens, err := scanQueryTokens(c.query)
	if err != nil {
		return err
	}
	c.parsedQuery = tokens
	return nil
}

// Get takes a dora query, prepares and validates it, executes the query, and returns the result or an error.
func (c *Client) get(query string) (string, error) {
	if err := c.prepAndExecQuery(query); err != nil {
		return "", err
	}
	return c.result, nil
}

// executeQuery is called after the JSON and the query are parsed into their respective
// tokens. We then iterate over the query tokens, and traverse our tree attempting to
// find the result the user is looking for.
func (c *Client) executeQuery() error {
	rootVal := *c.tree.RootValue
	obj, _ := rootVal.(ast.Object)
	arr, ok := rootVal.(ast.Array)
	currentType := ast.ObjectType
	if ok {
		currentType = ast.ArrayType
	}
	parsedQueryLen := len(c.parsedQuery)

	for i := 0; i < parsedQueryLen; i++ {
		// If i == parsedQueryLen-1, we are on the final iteration
		if i == parsedQueryLen-1 {
			c.setFinalValue(currentType, i, obj, arr)
		}

		// If the query token we're on is asking for an object
		if c.parsedQuery[i].accessType == ObjectAccess {
			if currentType != ast.ObjectType {
				return errors.New("incorrect syntax, your query asked for an object but found array")
			}
			var found bool

			for _, v := range obj.Children {
				if v.Key.Value == c.parsedQuery[i].key {
					found = true
					o, astObj := v.Value.(ast.Object)
					if astObj {
						obj = o
						currentType = ast.ObjectType
						break
					}
					a, astArr := v.Value.(ast.Array)
					if astArr {
						arr = a
						currentType = ast.ArrayType
						break
					}
				}
			}
			if !found {
				return fmt.Errorf("Sorry, could not find a key with that value. Key: %s", c.parsedQuery[i].key)
			}
		} else { // If the query token we're on is asking for an array
			if currentType != ast.ArrayType {
				return errors.New("incorrect syntax, your query asked for an array but found object")
			}
			qt := c.parsedQuery[i]
			val := arr.Children[qt.index]

			switch v := val.(type) {
			case ast.Object:
				obj = v
				currentType = ast.ObjectType
				break
			case ast.Array:
				arr = v
				currentType = ast.ArrayType
				break
			case ast.Literal:
				// If we're on the final value, return it
				if i == parsedQueryLen-1 {
					c.setResultFromValue(v)
				} else {
					return errors.New("Sorry, it looks like your query isn't quite right")
				}
			}
		}
	}

	return nil
}

// setFinalValue is called when we are on the final queryToken. It handles narrowing down what
// needs to be returned and sets the result to the Client
func (c *Client) setFinalValue(currentType ast.Type, index int, obj ast.Object, arr ast.Array) {
	if currentType == ast.ObjectType {
		r := c.parsedQuery[index].key
		for _, v := range obj.Children {
			if r == v.Key.Value {
				c.setResultFromValue(v.Value)
				break
			}
		}
		return
	}
	ind := c.parsedQuery[index].index
	c.setResultFromValue(arr.Children[ind])
}

// setResultFromValue switches on an ast.Value type and assigns the appropriate result to the client
func (c *Client) setResultFromValue(value ast.Value) {
	c.resultValue = value
	switch val := value.(type) {
	case ast.Literal:
		c.result = val.String()
	case ast.Object:
		c.result = string(c.input[val.Start:val.End])
	case ast.Array:
		c.result = val.String()
	}
}

// validateQueryRoot handles some very simple validation around the root of the query
func validateQueryRoot(query string, rootNodeType ast.RootNodeType) error {
	if query[0] != '$' {
		return ErrNoDollarSignRoot
	}

	// The query root after the `$` must be a `.` if the rootNodeType is an object
	validObjQueryRoot := query[1] == '.'
	if rootNodeType == ast.ObjectRoot && !validObjQueryRoot {
		return ErrWrongObjectRootSelector
	}

	// The query root after the `$` must be a `[` if the rootNodeType is an array
	validArrayQueryRoot := query[1] == '['
	if rootNodeType == ast.ArrayRoot && !validArrayQueryRoot {
		return ErrWrongArrayRootSelector
	}

	return nil
}

func errSelectorSytax(operator string) error {
	return fmt.Errorf(
		"Error parsing query, expected either a `.` for selections on an object or a `[` for selections on an array. Got: %s",
		operator,
	)
}

// Package dora TODO: package docs
package dora

import (
	"strconv"

	"github.com/bradford-hamilton/dora/pkg/ast"
	"github.com/bradford-hamilton/dora/pkg/lexer"
	"github.com/bradford-hamilton/dora/pkg/parser"
)

// Client represents a dora client. The client holds things like a copy of the input, the tree (the
// parsed AST representation built with Go types), the user's query & parsed version of the query, and
// a query result. Client exposes public methods which access this underlying data.
type Client struct {
	input       []byte
	tree        *ast.RootNode
	query       []byte
	parsedQuery []queryToken
	result      string
	resultValue ast.Value
}

// NewFromString takes a string, creates a lexer, creates a parser from the lexer,
// and parses the json into an AST. Methods on the Client give access to private
// data like the AST held inside.
func NewFromString(jsonStr string) (*Client, error) {
	l := lexer.New(jsonStr)
	p := parser.New(l)
	tree, err := p.ParseJSON()
	if err != nil {
		return nil, err
	}
	return &Client{tree: &tree, input: l.Input}, nil
}

// NewFromBytes takes a slice of bytes, converts it to a string, then returns `NewFromString`, passing in the JSON string.
func NewFromBytes(bytes []byte) (*Client, error) {
	return NewFromString(string(bytes))
}

// GetString wraps a call to `get` and returns the result as a string
func (c *Client) GetString(query string) (string, error) {
	result, err := c.get(query)
	if err != nil {
		return "", err
	}
	return result, nil
}

// GetBool wraps a call to `get` and returns the result as a bool
func (c *Client) GetBool(query string) (bool, error) {
	res, err := c.get(query)
	if err != nil {
		return false, err
	}
	s, err := strconv.ParseBool(res)
	if err != nil {
		return false, err
	}
	return s, nil
}

// GetFloat64 wraps a call to `get` and returns the result as a float64 (JSONs only number type)
func (c *Client) GetFloat64(query string) (float64, error) {
	res, err := c.get(query)
	if err != nil {
		return 0.0, err
	}
	f, err := strconv.ParseFloat(res, 64)
	if err != nil {
		return 0.0, err
	}
	return f, nil
}

// GetObject wraps a call to `get` and returns the result as an interface{}
func (c *Client) GetObject(query string) (interface{}, error) {
	if err := c.prepAndExecQuery(query); err != nil {
		return nil, err
	}
	return c.resultValue.GoType(), nil
}

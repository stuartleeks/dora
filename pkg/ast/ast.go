// Package ast TODO: package docs
package ast

import (
	"fmt"
	"strconv"
)

// These are the available root node types. In JSON it will either be an
// object or an array at the base.
const (
	ObjectRoot RootNodeType = iota
	ArrayRoot
)

// RootNodeType is a type alias for an int
type RootNodeType int

// RootNode is what starts every parsed AST. There is a `Type` field so that
// you can ask which root node type starts the tree.
type RootNode struct {
	RootValue *Value
	Type      RootNodeType
}

// Available ast value types
const (
	ObjectType Type = iota
	ArrayType
	LiteralType
	PropertyType
	IdentifierType
)

// Type is a type alias for int. Represents a values type.
type Type int

// Value will eventually have some methods that all Values must implement. For now
// it represents any JSON value (object | array | boolean | string | number | null)
type Value interface {
	String() string
	GoType() interface{}
}

// Object represents a JSON object. It holds a slice of Property as its children,
// a Type ("Object"), and start & end code points for displaying.
type Object struct {
	Type      Type
	Children  []Property
	Start     int
	End       int
	sourceBuf *[]byte
}

func NewObject(sourceBuf *[]byte) Object {
	return Object{
		Type:      ObjectType,
		sourceBuf: sourceBuf,
	}
}

func (o Object) String() string {
	return string((*o.sourceBuf)[o.Start:o.End])
}
func (o Object) GoType() interface{} {
	result := map[string]interface{}{}
	for _, property := range o.Children {
		result[property.Key.String()] = property.Value.GoType()
	}
	return result
}

var _ Value = Object{}

// Array represents a JSON array It holds a slice of Value as its children,
// a Type ("Array"), and start & end code points for displaying.
type Array struct {
	Type      Type
	Children  []Value
	Start     int
	End       int
	sourceBuf *[]byte
}

func NewArray(sourceBuf *[]byte) Array {
	return Array{
		Type:      ArrayType,
		sourceBuf: sourceBuf,
	}
}
func (a Array) String() string {
	return string((*a.sourceBuf)[a.Start:a.End])
}
func (a Array) GoType() interface{} {
	result := make([]interface{}, len(a.Children))
	for i, child := range a.Children {
		result[i] = child.GoType()
	}
	return result
}

var _ Value = Array{}

// Literal represents a JSON literal value. It holds a Type ("Literal") and the actual value.
type Literal struct {
	Type  Type
	Value interface{}
}

func (l Literal) String() string {
	switch lit := l.Value.(type) {
	case string:
		return lit
	case float64:
		return fmt.Sprintf("%f", lit)
	case int:
		return strconv.Itoa(lit)
	case bool:
		return fmt.Sprintf("%v", lit)
	case nil:
		return "null"
	default:
		return fmt.Sprintf("%v", lit)
	}
}
func (l Literal) GoType() interface{} {
	return l.Value
}

var _ Value = Literal{}

// Property holds a Type ("Property") as well as a `Key` and `Value`. The Key is an Identifier
// and the value is any Value.
type Property struct {
	Type  Type
	Key   Identifier
	Value Value
}

func (p Property) String() string {
	return fmt.Sprintf("%s=%s", p.Key.String(), p.Value.String())
}
func (p Property) GoType() interface{} {
	return nil // TODO - revisit this
}

var _ Value = Property{}

// Identifier represents a JSON object property key
type Identifier struct {
	Type  Type
	Value string // "key1"
}

func (i Identifier) String() string {
	return i.Value
}
func (i Identifier) GoType() interface{} {
	return nil // TODO - revisit this
}

var _ Value = Identifier{}

// state is a type alias for int and used to create the available value states below
type state int

// Available states for each type used in parsing
const (
	// Object states
	ObjStart state = iota
	ObjOpen
	ObjProperty
	ObjComma

	// Property states
	PropertyStart
	PropertyKey
	PropertyColon

	// Array states
	ArrayStart
	ArrayOpen
	ArrayValue
	ArrayComma

	// String states
	StringStart
	StringQuoteOrChar
	Escape

	// Number states
	NumberStart
	NumberMinus
	NumberZero
	NumberDigit
	NumberPoint
	NumberDigitFraction
	NumberExp
	NumberExpDigitOrSign
)

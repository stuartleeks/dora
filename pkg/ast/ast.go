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
	ArrayItemType
	LiteralType
	PropertyType
	IdentifierType
)

const (
	StringLiteralValueType LiteralValueType = iota
	NumberLiteralValueType
	NullLiteralValueType
	BooleanLiteralValueType
)

// Type is a type alias for int. Represents a node's type.
type Type int

// LiteralValueType is a type alias for int. Represents the type of the value in a Literal node
type LiteralValueType int

// StructuralItemType is a type alias for int. Represents the type of the structural item
type StructuralItemType int

const (
	WhitespaceStructuralItemType StructuralItemType = iota
	LineCommentStructuralItemType
	BlockCommentStructuralItemType
)

type StructuralItem struct {
	ItemType StructuralItemType
	Value    string
}

// Object represents a JSON object. It holds a slice of Property as its children,
// a Type ("Object"), and start & end code points for displaying.
type Object struct {
	Type            Type
	Children        []Property
	Start           int
	End             int
	SuffixStructure []StructuralItem
	sourceBuf       *[]byte
}

var _ ValueContent = Object{}

func NewObject(sourceBuf *[]byte) Object {
	return Object{
		Type:      ObjectType,
		sourceBuf: sourceBuf,
	}
}

func (o Object) String() string {
	return string((*o.sourceBuf)[o.Start:o.End])
}

// Array represents a JSON array It holds a slice of Value as its children,
// a Type ("Array"), and start & end code points for displaying.
type Array struct {
	Type            Type
	PrefixStructure []StructuralItem
	Children        []ArrayItem
	SuffixStructure []StructuralItem
	Start           int
	End             int
	sourceBuf       *[]byte
}

var _ ValueContent = Array{}

func NewArray(sourceBuf *[]byte) Array {
	return Array{
		Type:      ArrayType,
		sourceBuf: sourceBuf,
	}
}

func (a Array) String() string {
	return string((*a.sourceBuf)[a.Start:a.End])
}

// Array holds a Type ("ArrayItem") as well as a `Value` and whether there is a comma after the item
type ArrayItem struct {
	Type               Type
	PrefixStructure    []StructuralItem
	Value              ValueContent
	PostValueStructure []StructuralItem
	HasCommaSeparator  bool
}

var _ ValueContent = ArrayItem{}

func (ai ArrayItem) String() string {
	return ai.Value.String()
}

// Literal represents a JSON literal value. It holds a Type ("Literal") and the actual value.
type Literal struct {
	Type              Type
	ValueType         LiteralValueType
	Value             interface{}
	Delimiter         string // Delimiter is set for string values
	OriginalRendering string // Allows preservig numeric formatting from source documents
}

var _ ValueContent = Literal{}

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

// Property holds a Type ("Property") as well as a `Key` and `Value`. The Key is an Identifier
// and the value is any Value.
type Property struct {
	Type              Type
	Key               Identifier
	Value             Value
	HasCommaSeparator bool
}

// Identifier represents a JSON object property key
type Identifier struct {
	Type            Type
	PrefixStructure []StructuralItem
	Value           string // "key1"
	SuffixStructure []StructuralItem
	Delimiter       string
}

type Value struct {
	PrefixStructure []StructuralItem
	Content         ValueContent
	SuffixStructure []StructuralItem
}

var _ ValueContent = Value{}

func (v Value) String() string {
	return v.Content.String()
}

// ValueContent will eventually have some methods that all Values must implement. For now
// it represents any JSON value (object | array | boolean | string | number | null)
type ValueContent interface {
	String() string
}

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
	PropertyValue

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

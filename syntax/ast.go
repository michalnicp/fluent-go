package syntax

type Resource struct {
	Body []Entry `json:"body"`
}

func (a Resource) MarshalJSON() ([]byte, error) {
	type alias Resource
	tmp := struct {
		Type string `json:"type"`
		alias
	}{
		Type:  "Resource",
		alias: alias(a),
	}
	return marshal(tmp)
}

type Entry interface {
	Entry()
}

func (a Message) Entry()         {}
func (a Term) Entry()            {}
func (a Comment) Entry()         {}
func (a GroupComment) Entry()    {}
func (a ResourceComment) Entry() {}
func (a Junk) Entry()            {}

type Junk struct {
	Annotations []Annotation `json:"annotations"`
	Content     string       `json:"content"`
}

func (a Junk) MarshalJSON() ([]byte, error) {
	type alias Junk
	tmp := struct {
		Type string `json:"type"`
		alias
	}{
		Type:  "Junk",
		alias: alias(a),
	}
	return marshal(tmp)
}

// TODO: Implement. See https://github.com/projectfluent/fluent/pull/40
type Annotation struct{}

func (a Annotation) MarshalJSON() ([]byte, error) {
	type alias Annotation
	tmp := struct {
		Type string `json:"type"`
		alias
	}{
		Type:  "Annotation",
		alias: alias(a),
	}
	return marshal(tmp)
}

type Message struct {
	ID         Identifier  `json:"id"`
	Value      *Pattern    `json:"value"`
	Attributes []Attribute `json:"attributes"`
	Comment    *Comment    `json:"comment"`
}

func (a Message) MarshalJSON() ([]byte, error) {
	type alias Message
	tmp := struct {
		Type string `json:"type"`
		alias
	}{
		Type:  "Message",
		alias: alias(a),
	}
	return marshal(tmp)
}

type Term struct {
	ID         Identifier  `json:"id"`
	Value      Pattern     `json:"value"`
	Attributes []Attribute `json:"attributes"`
	Comment    *Comment    `json:"comment"`
}

func (a Term) MarshalJSON() ([]byte, error) {
	type alias Term
	tmp := struct {
		Type string `json:"type"`
		alias
	}{
		Type:  "Term",
		alias: alias(a),
	}
	return marshal(tmp)
}

type Pattern struct {
	Elements []PatternElement `json:"elements"`
}

func (a Pattern) MarshalJSON() ([]byte, error) {
	type alias Pattern
	tmp := struct {
		Type string `json:"type"`
		alias
	}{
		Type:  "Pattern",
		alias: alias(a),
	}
	return marshal(tmp)
}

type PatternElement interface {
	PatternElement()
}

func (a TextElement) PatternElement() {}
func (a Placeable) PatternElement()   {}

type Attribute struct {
	ID    Identifier `json:"id"`
	Value Pattern    `json:"value"`
}

func (a Attribute) MarshalJSON() ([]byte, error) {
	type alias Attribute
	tmp := struct {
		Type string `json:"type"`
		alias
	}{
		Type:  "Attribute",
		alias: alias(a),
	}
	return marshal(tmp)
}

type Identifier struct {
	Name string `json:"name"`
}

func (a Identifier) MarshalJSON() ([]byte, error) {
	type alias Identifier
	tmp := struct {
		Type string `json:"type"`
		alias
	}{
		Type:  "Identifier",
		alias: alias(a),
	}
	return marshal(tmp)
}

type Variant struct {
	Key     VariantKey `json:"key"`
	Value   Pattern    `json:"value"`
	Default bool       `json:"default"`
}

func (a Variant) MarshalJSON() ([]byte, error) {
	type alias Variant
	tmp := struct {
		Type string `json:"type"`
		alias
	}{
		Type:  "Variant",
		alias: alias(a),
	}
	return marshal(tmp)
}

type VariantKey interface {
	VariantKey()
}

func (a Identifier) VariantKey()    {}
func (a NumberLiteral) VariantKey() {}

type CommentLine interface {
	CommentLine()
}

func (a Comment) CommentLine()         {}
func (a GroupComment) CommentLine()    {}
func (a ResourceComment) CommentLine() {}

type Comment struct {
	Content string `json:"content"`
}

func (a Comment) MarshalJSON() ([]byte, error) {
	type alias Comment
	tmp := struct {
		Type string `json:"type"`
		alias
	}{
		Type:  "Comment",
		alias: alias(a),
	}
	return marshal(tmp)
}

type GroupComment struct {
	Content string `json:"content"`
}

func (a GroupComment) MarshalJSON() ([]byte, error) {
	type alias GroupComment
	tmp := struct {
		Type string `json:"type"`
		alias
	}{
		Type:  "GroupComment",
		alias: alias(a),
	}
	return marshal(tmp)
}

type ResourceComment struct {
	Content string `json:"content"`
}

func (a ResourceComment) MarshalJSON() ([]byte, error) {
	type alias ResourceComment
	tmp := struct {
		Type string `json:"type"`
		alias
	}{
		Type:  "ResourceComment",
		alias: alias(a),
	}
	return marshal(tmp)
}

type TextElement struct {
	Value string `json:"value"`
}

func (a TextElement) MarshalJSON() ([]byte, error) {
	type alias TextElement
	tmp := struct {
		Type string `json:"type"`
		alias
	}{
		Type:  "TextElement",
		alias: alias(a),
	}
	return marshal(tmp)
}

type InlineExpression interface {
	InlineExpression()
}

func (a StringLiteral) InlineExpression()     {}
func (a NumberLiteral) InlineExpression()     {}
func (a FunctionReference) InlineExpression() {}
func (a MessageReference) InlineExpression()  {}
func (a TermReference) InlineExpression()     {}
func (a VariableReference) InlineExpression() {}
func (a Placeable) InlineExpression()         {}

type StringLiteral struct {
	Value string `json:"value"`
}

func (a StringLiteral) MarshalJSON() ([]byte, error) {
	type alias StringLiteral
	tmp := struct {
		Type string `json:"type"`
		alias
	}{
		Type:  "StringLiteral",
		alias: alias(a),
	}
	return marshal(tmp)
}

type NumberLiteral struct {
	Value string `json:"value"`
}

func (a NumberLiteral) MarshalJSON() ([]byte, error) {
	type alias NumberLiteral
	tmp := struct {
		Type string `json:"type"`
		alias
	}{
		Type:  "NumberLiteral",
		alias: alias(a),
	}
	return marshal(tmp)
}

type FunctionReference struct {
	ID        Identifier    `json:"id"`
	Arguments CallArguments `json:"arguments"`
}

func (a FunctionReference) MarshalJSON() ([]byte, error) {
	type alias FunctionReference
	tmp := struct {
		Type string `json:"type"`
		alias
	}{
		Type:  "FunctionReference",
		alias: alias(a),
	}
	return marshal(tmp)
}

type MessageReference struct {
	ID        Identifier  `json:"id"`
	Attribute *Identifier `json:"attribute"`
}

func (a MessageReference) MarshalJSON() ([]byte, error) {
	type alias MessageReference
	tmp := struct {
		Type string `json:"type"`
		alias
	}{
		Type:  "MessageReference",
		alias: alias(a),
	}
	return marshal(tmp)
}

type TermReference struct {
	ID        Identifier     `json:"id"`
	Attribute *Identifier    `json:"attribute"`
	Arguments *CallArguments `json:"arguments"`
}

func (a TermReference) MarshalJSON() ([]byte, error) {
	type alias TermReference
	tmp := struct {
		Type string `json:"type"`
		alias
	}{
		Type:  "TermReference",
		alias: alias(a),
	}
	return marshal(tmp)
}

type VariableReference struct {
	ID Identifier `json:"id"`
}

func (a VariableReference) MarshalJSON() ([]byte, error) {
	type alias VariableReference
	tmp := struct {
		Type string `json:"type"`
		alias
	}{
		Type:  "VariableReference",
		alias: alias(a),
	}
	return marshal(tmp)
}

type Placeable struct {
	Expr Expression `json:"expression"`
}

func (a Placeable) MarshalJSON() ([]byte, error) {
	type alias Placeable
	tmp := struct {
		Type string `json:"type"`
		alias
	}{
		Type:  "Placeable",
		alias: alias(a),
	}
	return marshal(tmp)
}

// Expression can be an InlineExpression or SelectExpression.
type Expression interface {
	Expression()
}

func (a StringLiteral) Expression()     {}
func (a NumberLiteral) Expression()     {}
func (a FunctionReference) Expression() {}
func (a MessageReference) Expression()  {}
func (a TermReference) Expression()     {}
func (a VariableReference) Expression() {}
func (a Placeable) Expression()         {}
func (a SelectExpression) Expression()  {}

type SelectExpression struct {
	Selector InlineExpression `json:"selector"`
	Variants []Variant        `json:"variants"`
}

func (a SelectExpression) MarshalJSON() ([]byte, error) {
	type alias SelectExpression
	tmp := struct {
		Type string `json:"type"`
		alias
	}{
		Type:  "SelectExpression",
		alias: alias(a),
	}
	return marshal(tmp)
}

type CallArguments struct {
	Positional []InlineExpression `json:"positional"`
	Named      []NamedArgument    `json:"named"`
}

func (a CallArguments) MarshalJSON() ([]byte, error) {
	type alias CallArguments
	tmp := struct {
		Type string `json:"type"`
		alias
	}{
		Type:  "CallArguments",
		alias: alias(a),
	}
	return marshal(tmp)
}

type NamedArgument struct {
	Name  Identifier       `json:"name"`
	Value InlineExpression `json:"value"`
}

func (a NamedArgument) MarshalJSON() ([]byte, error) {
	type alias NamedArgument
	tmp := struct {
		Type string `json:"type"`
		alias
	}{
		Type:  "NamedArgument",
		alias: alias(a),
	}
	return marshal(tmp)
}

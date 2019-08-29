package syntax

import (
	"fmt"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"
)

var eof = rune(0)

func Parse(input []byte) (Resource, error) {
	return newParser(input).parse()
}

type parser struct {
	input []byte
	pos   int
	ch    rune
	w     int
	line  int
	col   int
}

func newParser(input []byte) *parser {
	p := parser{
		input: input,
		pos:   0,
		line:  1,
	}
	p.next()

	return &p
}

// next advances the parser by one rune. Updates line and column on the parser.
func (p *parser) next() {
	if p.ch == '\n' {
		p.line++
		p.col = 0
	}
	p.col++

	if p.pos >= len(p.input) {
		p.ch = eof
		return
	}

	// TODO: check for utf8 errors
	p.pos += p.w
	p.ch, p.w = utf8.DecodeRune(p.input[p.pos:])
	if p.ch == utf8.RuneError && p.w == 1 {
		// invalid utf8 encoding
		fmt.Fprintln(os.Stderr, "invalid utf8 encoding")
	}

	return
}

func (p *parser) peek() rune {
	if p.pos > len(p.input)-1 {
		return eof
	}
	ch, _ := utf8.DecodeRune(p.input[p.pos+p.w:])
	return ch
}

func (p *parser) peekn(n int) string {
	var runes []rune

	pos := p.pos + p.w
	for i := 0; i < n; i++ {
		r, size := utf8.DecodeRune(p.input[pos:])
		if size > 0 {
			pos += size
			runes = append(runes, r)
		}
	}

	return string(runes)
}

func (p *parser) errorf(format string, a ...interface{}) error {
	return p.error(fmt.Sprintf(format, a...))
}

func (p *parser) error(message string) error {
	return newParseError(p.line, p.col, p.pos, message)
}

func (p *parser) skipWhitespace() {
	for p.ch == ' ' || p.ch == '\t' || p.ch == '\n' {
		p.next()
	}
}

func (p *parser) isEOL() bool {
	switch {
	case p.ch == '\n':
		return true
	case p.ch == '\r' && p.peek() == '\n':
		return true
	default:
		return false
	}
}

func (p *parser) skipEOL() bool {
	switch {
	case p.ch == '\n':
		p.next()
		return true
	case p.ch == '\r' && p.peek() == '\n':
		p.next()
		p.next()
		return true
	default:
		return false
	}
}

func (p *parser) skipBlankInline() int {
	start := p.pos
	for p.ch == ' ' {
		p.next()
	}
	return p.pos - start
}

func (p *parser) skipBlankBlock() int {
	var count int
	for {
		start, ch := p.pos, p.ch
		p.skipBlankInline()
		if !p.skipEOL() {
			p.col = 1
			p.pos, p.ch = start, ch
			break
		}
		count++
	}
	return count
}

func (p *parser) skipBlank() {
	for {
		p.skipBlankInline()
		if !p.skipEOL() {
			break
		}
	}
}

func (p *parser) skipToNextEntryStart() {
	for p.ch != eof {
		newline := false
		if p.pos == 0 || p.input[p.pos-1] == '\n' {
			newline = true
		}
		if newline && (isLetter(p.ch) || p.ch == '-' || p.ch == '#') {
			break
		}
		p.next()
	}
}

func (p *parser) skipUnicodeEscapeSequence() error {
	var need int
	switch p.ch {
	case 'u':
		need = 4
	case 'U':
		need = 6
	}
	p.next()

	taken := 0
	for p.ch != eof {
		if taken >= need {
			break
		}
		if !isHex(p.ch) {
			break
		}
		taken++
		p.next()
	}
	if taken < need {
		return p.error("invalid unicode escape sequence")
	}
	return nil
}

func (p *parser) skipDigits() int {
	count := 0
	for isDigit(p.ch) {
		p.next()
		count++
	}
	return count
}

func (p *parser) debug(n int) {
	min := p.pos - n
	if min < 0 {
		min = 0
	}

	max := p.pos + n
	if max > len(p.input)-1 {
		max = len(p.input) - 1
	}

	fmt.Println(string(p.input[min:max]))
}

// parse parses the root level resource.
func (p *parser) parse() (Resource, error) {
	var errors []error

	p.skipBlankBlock()

	entries := make([]Entry, 0)
	var lastComment *Comment

	for p.pos < len(p.input) {
		start := p.pos

		entry, err := p.parseEntry()
		if err != nil {
			errors = append(errors, err)

			p.skipToNextEntryStart()
			content := string(p.input[start:p.pos])
			entry = Junk{
				Content:     content,
				Annotations: make([]Annotation, 0),
			}
		}

		blankLines := p.skipBlankBlock()
		if comment, ok := entry.(Comment); ok && blankLines == 0 {
			lastComment = &comment
			continue
		}

		if lastComment != nil {
			switch v := entry.(type) {
			case Message:
				v.Comment = lastComment
				entry = v
			case Term:
				v.Comment = lastComment
				entry = v
			default:
				entries = append(entries, *lastComment)
			}
			lastComment = nil
		}

		entries = append(entries, entry)
	}

	if lastComment != nil {
		entries = append(entries, *lastComment)
		lastComment = nil
	}

	resource := Resource{
		Body: entries,
	}

	var err error
	if len(errors) > 0 {
		err = &ParseErrors{
			input:  p.input,
			errors: errors,
		}
	}

	return resource, err
}

func (p *parser) parseEntry() (Entry, error) {
	switch p.ch {
	case '#':
		return p.parseComment()
	case '-':
		return p.parseTerm()
	default:
		return p.parseMessage()
	}
}

// parseCommentLevel returns to
func (p *parser) parseCommentLevel() int {
	level := 0
	for p.ch != eof {
		if p.ch != '#' {
			break
		}
		level++
		p.next()
	}
	return level
}

func (p *parser) parseCommentLine() string {
	start := p.pos
	for p.ch != eof {
		if p.isEOL() {
			break
		}
		p.next()
	}
	line := string(p.input[start:p.pos])
	return line
}

func (p *parser) parseComment() (Entry, error) {
	var lines []string

	lastLevel := 0
	for p.ch != eof {
		pos, ch := p.pos, p.ch
		level := p.parseCommentLevel()
		if level == 0 {
			break
		}
		if lastLevel != 0 && level != lastLevel {
			p.pos, p.ch = pos, ch
			break
		}
		lastLevel = level

		var line string
		if !p.isEOL() {
			if p.ch != ' ' {
				return Comment{}, p.error(fmt.Sprintf("expected %q, found %q", ' ', p.ch))
			}
			p.next() // skip ' '
			line = p.parseCommentLine()
		}

		lines = append(lines, line)
		p.skipEOL()
	}

	content := strings.Join(lines, "\n")

	switch lastLevel {
	case 1:
		return Comment{Content: content}, nil
	case 2:
		return GroupComment{Content: content}, nil
	case 3:
		return ResourceComment{Content: content}, nil
	default:
		panic("shouldn't happen")
	}
}

func (p *parser) parseMessage() (Message, error) {
	id, err := p.parseIdentifier()
	if err != nil {
		return Message{}, err
	}

	p.skipBlankInline()

	if p.ch != '=' {
		return Message{}, p.error(fmt.Sprintf("expected %q, got %q", '=', p.ch))
	}
	p.next()

	pattern, err := p.parsePattern()
	if err != nil {
		return Message{}, err
	}

	p.skipBlankBlock()

	attributes, err := p.parseAttributes()
	if err != nil {
		return Message{}, err
	}

	if len(pattern.Elements) == 0 && len(attributes) == 0 {
		return Message{}, p.error("expected message field")
	}

	message := Message{
		ID:         id,
		Attributes: attributes,
	}

	if len(pattern.Elements) != 0 {
		message.Value = &pattern
	}

	return message, nil
}

func (p *parser) parsePattern() (Pattern, error) {
	var elements []PatternElement

	var (
		block        = false // inline or block text/placeable
		lastNonBlank = -1    // last non blank element index
		commonIndent = 0     // track comment indent for detentation
	)

	p.skipBlankInline()

	if p.skipEOL() {
		p.skipBlankBlock()
		block = true
	}

	for p.ch != eof {
		if p.skipEOL() {
			block = true
		}

		if p.ch == '{' {
			element, err := p.parsePlaceable()
			if err != nil {
				return Pattern{}, err
			}
			lastNonBlank = len(elements)
			elements = append(elements, element)
			continue
		}

		if block {
			start, ch := p.pos, p.ch
			indent := p.skipBlankInline()
			if indent == 0 && !p.isEOL() {
				break
			}
			if p.ch == '[' || p.ch == '*' || p.ch == '.' {
				break
			}

			if indent > 0 &&
				(indent < commonIndent || commonIndent == 0) {
				commonIndent = indent
			}
			p.pos, p.ch = start, ch
		}

		element, err := p.parseTextElement()
		if err != nil {
			return Pattern{}, err
		}

		if element.Value != "" {
			lastNonBlank = len(elements)
		}

		elements = append(elements, element)
	}

	if lastNonBlank < 0 {

		// TODO: Should this be an error?
		return Pattern{}, nil
	}

	// dedent common indent, remove trailing whitespace, and join adjacent text elements
	var processed []PatternElement

	// TODO: trailing newlines are being included.
	indent := strings.Repeat(" ", commonIndent)
	var buf []string
	for _, element := range elements {
		if text, ok := element.(TextElement); ok {
			s := strings.TrimPrefix(text.Value, indent)
			buf = append(buf, s)
			continue
		}
		if len(buf) > 0 {
			text := TextElement{
				Value: strings.Join(buf, "\n"),
			}
			processed = append(processed, text)
			buf = nil
		}
		processed = append(processed, element)
	}
	if len(buf) > 0 {
		value := strings.TrimRightFunc(strings.Join(buf, "\n"), unicode.IsSpace)
		if value != "" {
			text := TextElement{
				Value: value,
			}
			processed = append(processed, text)
		}
	}

	pattern := Pattern{
		Elements: processed,
	}
	return pattern, nil
}

func (p *parser) parsePlaceable() (Placeable, error) {
	p.next() // skip '{'

	p.skipBlank()

	expr, err := p.parseExpression()
	if err != nil {
		return Placeable{}, err
	}

	p.skipBlankInline()

	if p.ch != '}' {
		return Placeable{}, p.error("missing closing '}'")
	}
	p.next()

	placeable := Placeable{
		Expr: expr,
	}

	return placeable, nil
}

// parses a text element up to but not including a newline.
// TODO: consider changing this to just return a line of text
func (p *parser) parseTextElement() (TextElement, error) {
	start := p.pos

loop:
	for p.ch != eof {
		if p.isEOL() {
			break
		}
		switch p.ch {
		case '{':
			break loop
		case '}':
			return TextElement{}, p.error("unbalanced closing '}'")
		default:
		}
		p.next()
	}

	value := string(p.input[start:p.pos])
	text := TextElement{
		Value: value,
	}

	return text, nil
}

func (p *parser) parseExpression() (Expression, error) {
	selector, err := p.parseInlineExpression()
	if err != nil {
		return nil, err
	}

	p.skipBlank()

	if p.ch != '-' || p.peek() != '>' {
		if ref, ok := selector.(TermReference); ok {
			if ref.Attribute != nil {
				return nil, p.error("term attribute as placeable")
			}
		}
		return selector.(Expression), nil
	}

	if ref, ok := selector.(MessageReference); ok {
		if ref.Attribute == nil {
			return nil, p.error("message reference as selector")
		}
		return nil, p.error("message attribute as selector")
	}
	if ref, ok := selector.(TermReference); ok {
		if ref.Attribute == nil {
			return nil, p.error("term attribute used as placeable")
		}
	}

	p.next() // skip '-'
	p.next() // skip '>'

	p.skipBlankInline()
	if !p.skipEOL() {
		return nil, p.error("expected eol")
	}
	p.skipBlank()

	variants, err := p.parseVariants()
	if err != nil {
		return nil, p.errorf("parse variants: %v", err)
	}

	selectExp := SelectExpression{
		Selector: selector,
		Variants: variants,
	}

	return selectExp, nil
}

func (p *parser) parseLiteral() (InlineExpression, error) {
	if isDigit(p.ch) {
		return p.parseNumberLiteral()
	}
	if p.ch == '"' {
		return p.parseStringLiteral()
	}
	return nil, p.error("expected literal") // E0014
}

func (p *parser) parseInlineExpression() (InlineExpression, error) {
	switch {
	case p.ch == '"':
		return p.parseStringLiteral()
	case isDigit(p.ch):
		return p.parseNumberLiteral()
	case p.ch == '-':
		ch := p.peek()
		if isLetter(ch) {
			p.next()

			id, err := p.parseIdentifier()
			if err != nil {
				return nil, err
			}

			var attr *Identifier
			if p.ch == '.' {
				p.next()
				id, err := p.parseIdentifier()
				if err != nil {
					return nil, err
				}
				attr = &id
			}

			var arguments *CallArguments
			if p.ch == '(' {
				args, err := p.parseCallArguments()
				if err != nil {
					return nil, err
				}
				arguments = &args
			}

			ref := TermReference{
				ID:        id,
				Attribute: attr,
				Arguments: arguments,
			}

			return ref, nil
		}
		return p.parseNumberLiteral()
	case p.ch == '$':
		p.next()
		id, err := p.parseIdentifier()
		if err != nil {
			return nil, err
		}
		ref := VariableReference{
			ID: id,
		}
		return ref, nil
	case isLetter(p.ch): // identifier start
		id, err := p.parseIdentifier()
		if err != nil {
			return nil, err
		}

		if p.ch == '(' { // it's a function
			for _, ch := range id.Name {
				if !(isUppercase(ch) || isDigit(ch) || ch == '_' || ch == '-') {
					return nil, p.error("forbidden callee") // E0008
				}
			}
			arguments, err := p.parseCallArguments()
			if err != nil {
				return nil, err
			}
			ref := FunctionReference{
				ID:        id,
				Arguments: arguments,
			}
			return ref, nil
		}

		var attr *Identifier
		if p.ch == '.' {
			p.next()
			id, err := p.parseIdentifier()
			if err != nil {
				return nil, err
			}
			attr = &id
		}

		ref := MessageReference{
			ID:        id,
			Attribute: attr,
		}
		return ref, nil
	case p.ch == '{':
		return p.parsePlaceable()
	default:
		return nil, p.error("expected inline expression") // E0028
	}
}

func (p *parser) parseCallArguments() (CallArguments, error) {
	p.next() // skip '('

	positional := make([]InlineExpression, 0)
	named := make([]NamedArgument, 0)
	var argumentNames []string

	p.skipBlank()

	for p.ch != eof {
		if p.ch == ')' {
			break
		}

		exp, err := p.parseInlineExpression()
		if err != nil {
			return CallArguments{}, err
		}

		p.skipBlank()

		if p.ch == ':' { // named argument
			ref, ok := exp.(MessageReference)
			if !ok || ref.Attribute != nil {
				return CallArguments{}, p.error("argument name must be simple identifier") // E0009
			}

			p.next() // skip ':'
			p.skipBlank()

			value, err := p.parseLiteral()
			if err != nil {
				return CallArguments{}, err
			}

			if containsString(argumentNames, ref.ID.Name) {
				return CallArguments{}, p.error("named arguments must be unique")
			}

			arg := NamedArgument{
				Name:  ref.ID,
				Value: value,
			}

			named = append(named, arg)
			argumentNames = append(argumentNames, arg.Name.Name)
		} else if len(argumentNames) > 0 {
			return CallArguments{}, p.error("positional argument follows names") // E0021
		} else {
			positional = append(positional, exp)
		}

		p.skipBlank()

		if p.ch == ',' {
			p.next()
			p.skipBlank()
			continue
		}

		break
	}

	if p.ch != ')' {
		return CallArguments{}, p.error("expected ')'")
	}
	p.next()

	args := CallArguments{
		Positional: positional,
		Named:      named,
	}

	return args, nil
}

func (p *parser) parseStringLiteral() (StringLiteral, error) {
	p.next() // skip '"'

	start := p.pos
loop:
	for p.ch != eof {
		if p.isEOL() {
			return StringLiteral{}, p.error("unexpected eol")
		}
		switch p.ch {
		case '\\': // escape special characters
			p.next()
			switch p.ch {
			case '\\', '"':
				p.next()
			case 'u', 'U':
				if err := p.skipUnicodeEscapeSequence(); err != nil {
					return StringLiteral{}, err
				}
			default:
				return StringLiteral{}, p.error("invalid escape sequence")
			}
		case '"':
			break loop
		default:
			p.next()
		}
	}

	value := string(p.input[start:p.pos])
	lit := StringLiteral{
		Value: value,
	}

	p.next() // skip closing '"'

	return lit, nil
}

func (p *parser) parseNumberLiteral() (NumberLiteral, error) {
	start := p.pos
	if p.ch == '-' {
		p.next()
	}

	if p.skipDigits() == 0 {
		return NumberLiteral{}, p.error("expected digit")
	}

	if p.ch == '.' {
		p.next()
		if p.skipDigits() == 0 {
			return NumberLiteral{}, p.error("expected digit")
		}
	}

	value := string(p.input[start:p.pos])
	lit := NumberLiteral{
		Value: value,
	}
	return lit, nil
}

func (p *parser) parseAttributes() ([]Attribute, error) {
	attributes := make([]Attribute, 0)

	for p.ch != eof {
		start := p.pos

		p.skipBlankInline()

		if p.ch != '.' {
			p.pos = start
			break
		}

		attr, err := p.parseAttribute()
		if err != nil {
			return nil, err
		}
		attributes = append(attributes, attr)
	}

	return attributes, nil
}

func (p *parser) parseAttribute() (Attribute, error) {
	p.next() // skip '.'

	id, err := p.parseIdentifier()
	if err != nil {
		return Attribute{}, err
	}

	p.skipBlankInline()

	if p.ch != '=' {
		return Attribute{}, p.error("unexpected character")
	}
	p.next()

	p.skipBlankInline()

	pattern, err := p.parsePattern()
	if err != nil {
		return Attribute{}, err
	}

	attr := Attribute{
		ID:    id,
		Value: pattern,
	}

	return attr, nil
}

func (p *parser) parseTerm() (Term, error) {
	if p.ch != '-' {
		return Term{}, p.error("expected '-'")
	}
	p.next()

	id, err := p.parseIdentifier()
	if err != nil {
		return Term{}, p.error("expected identifier")
	}
	p.skipBlankInline()

	if p.ch != '=' {
		return Term{}, p.error("expected '='")
	}
	p.next()

	p.skipBlankInline()

	value, err := p.parsePattern()
	if err != nil {
		return Term{}, err
	}

	p.skipBlankBlock()

	attributes, err := p.parseAttributes()
	if err != nil {
		return Term{}, err
	}

	term := Term{
		ID:         id,
		Value:      value,
		Attributes: attributes,
	}

	return term, nil
}

func (p *parser) parseIdentifier() (Identifier, error) {
	start := p.pos

	if !isLetter(p.ch) {
		return Identifier{}, p.error("expected identifier")
	}

	for p.ch != eof {
		p.next()
		if !(isLetter(p.ch) || isDigit(p.ch) || p.ch == '_' || p.ch == '-') {
			break
		}
	}

	id := Identifier{
		Name: string(p.input[start:p.pos]),
	}

	return id, nil
}

func (p *parser) parseVariants() ([]Variant, error) {
	variants := make([]Variant, 0)

	defaultVariant := false
	for p.ch == '*' || p.ch == '[' {
		if p.ch == '*' {
			defaultVariant = true
			p.next()
		}

		key, err := p.parseVariantKey()
		if err != nil {
			return nil, p.errorf("parse variant key: %v", err)
		}

		value, err := p.parsePattern()
		if err != nil {
			return nil, p.errorf("parse pattern: %v", err)
		}

		variant := Variant{
			Key:     key,
			Value:   value,
			Default: defaultVariant,
		}
		variants = append(variants, variant)
		p.skipBlank()
	}

	if !defaultVariant {
		return nil, p.error("missing default variant")
	}

	return variants, nil
}

func (p *parser) parseVariantKey() (VariantKey, error) {
	if p.ch != '[' {
		return nil, p.error("expected '['")
	}
	p.next()
	p.skipBlank()

	var key VariantKey
	var err error
	if isDigit(p.ch) {
		key, err = p.parseNumberLiteral()
	} else {
		key, err = p.parseIdentifier()
	}
	if err != nil {
		return nil, err
	}

	p.skipBlank()
	if p.ch != ']' {
		return nil, p.error("expected ']'")
	}
	p.next()

	return key, nil
}

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n'
}

func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

func isLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isHex(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9')
}

func isUppercase(ch rune) bool {
	return ch >= 'A' && ch <= 'Z'
}

func isTextChar(ch rune) bool {
	return !isSpecialTextChar(ch) && ch != '\n'
}

func isSpecialTextChar(ch rune) bool {
	return ch == '{' || ch == '}'
}

func containsString(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

func containsRune(a []rune, x rune) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

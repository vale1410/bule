package parser

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)


func Tokens(text string) (RuleElements []Token, err error) {
	p := parser{
		lex: lex(text),
	}

	for token := range p.lex.tokens {
		tokens = append(tokens,token)
	}
	for token := range p.lex.tokens {

		switch token.kind {
		case tokenEOF:
			p.result = sb.String()
			return
		case tokenError:
			p.errItem = &token
			fmt.Println("test", token.value)
			sb.WriteString("ERROR" + token.value)
			return
		case tokenDot:
			sb.WriteString("DOT{.}")
		default:
			sb.WriteString(printToken(token.kind) + "{" + token.value + "}")
			sb.WriteString("-")
		}
	}

	if p.errItem != nil {
		return tokens, fmt.Errorf("error processing the following %q", p.errItem.value)
	}
	return tokens, nil
}

func DebugString(text string) (string, error) {
	p := parser{
		lex: lex(text),
	}

	p.parse()



	return p.result, nil
}

type parser struct {
	result  string
	lex     *lexer
	errItem *Token
}


func (p *parser) parse() {
	sb := strings.Builder{}

	for token := range p.lex.tokens {

		switch token.kind {
		case tokenEOF:
			p.result = sb.String()
			return
		case tokenError:
			p.errItem = &token
			fmt.Println("test", token.value)
			sb.WriteString("ERROR" + token.value)
			return
		case tokenDot:
			sb.WriteString("DOT{.}")
		default:
			sb.WriteString(printToken(token.kind) + "{" + token.value + "}")
			sb.WriteString("-")
		}
	}
}

func parseExpression(tokens []Token) (tree SyntaxTree, pos int){

	switch token {
	case tokenAtomName:
	}

}

func parseAtomExpression(tokens []Token) (tree SyntaxTree, pos int){

	switch token {
	case tokenAtomName:
	}

}

type SyntaxTree struct {
	token Token
	children []SyntaxTree
}

type tokenKind int



const (
	tokenEOF tokenKind = iota
	tokenError
	tokenAtomName
	tokenAtomBracketLeft
	tokenAtomBracketRight
	tokenNegation
	tokenTermComma
	tokenRuleComma
	tokenColon
	tokenTermExpression
	tokenConstraint
	tokenEquivalence
	tokenImplication
	tokenDot
	tokenDoubleDot

	tokenCurlyBracketLeft
	tokenCurlyBracketRight
)


func printToken(token tokenKind) (s string) {
	switch token {
	case tokenAtomName:
		s = "ATOM"
	case tokenAtomBracketLeft:
		s = "PL"
	case tokenAtomBracketRight:
		s = "PR"
	case tokenNegation:
		s = "NEGATION"
	case tokenTermComma:
		s = "TERMCOMMA"
	case tokenRuleComma:
		s = "RULECOMMA"
	case tokenColon:
		s = "COLON"
	case tokenTermExpression:
		s = "TERM"
	case tokenConstraint:
		s = "CONSTRAINT"
	case tokenEquivalence:
		s = "EQUIVALENCE"
	case tokenImplication:
		s = "IMPLICATION"
	case tokenDot:
		s = "DOT"
	case tokenDoubleDot:
		s = "DOUBLEDOT"
	default:
		panic("not implemented tokentype")
	}
	return
}

// Token is accumulated while lexing the provided input, and emitted over a
// channel to the parser.
type Token struct {
	position int

	// kind signals how we've classified the data we have accumulated while
	// scanning the string.
	kind tokenKind

	// value is the segment of data we've accumulated.
	value string
}

const eof = -1

// stateFn is a function that is specific to a state within the string.
type stateFn func(*lexer) stateFn

// lex creates a lexer and starts scanning the provided input.
func lex(input string) *lexer {
	l := &lexer{
		input:  input,
		state:  lexRuleElement,
		tokens: make(chan Token, 1),
	}

	go l.scan()

	return l
}

// lexer is created to manage an individual scanning/parsing operation.
type lexer struct {
	input    string     // we'll store the string being parsed
	start    int        // the position we started scanning
	position int        // the current position of our scan
	width    int        // we'll be using runes which can be double byte
	state    stateFn    // the current state function
	tokens   chan Token // the channel we'll use to communicate between the lexer and the parser
}

// emit sends a Token over the channel so the parser can collect and manage
// each segment.
func (l *lexer) emit(k tokenKind) {
	accumulation := l.input[l.start:l.position]

	i := Token{
		position: l.start,
		kind:     k,
		value:    accumulation,
	}

	l.tokens <- i

	l.ignore() // reset our scanner now that we've dispatched a segment
}

// nextItem pulls an Token from the lexer's result channel.
func (l *lexer) nextItem() Token {
	return <-l.tokens
}

// ignore resets the start position to the current scan position effectively
// ignoring any input.
func (l *lexer) ignore() {
	l.start = l.position
}

// next advances the lexer state to the next rune.
func (l *lexer) next() (r rune) {
	if l.position >= len(l.input) {
		l.width = 0
		return eof
	}

	r, l.width = utf8.DecodeRuneInString(l.input[l.position:])
	l.position += l.width
	return r
}

// backup allows us to step back one run1e which is helpful when you've crossed
// a boundary from one state to another.
func (l *lexer) backup() {
	l.position = l.position - 1
}

// scan will step through the provided text and execute state functions as
// state changes are observed in the provided input.
func (l *lexer) scan() {
	for fn := lexRuleElement; fn != nil; {
		fn = fn(l)
	}
	close(l.tokens)
}

func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	msg := fmt.Sprintf(format, args...)
	l.tokens <- Token{
		kind:  tokenError,
		value: msg,
	}

	return nil
}

// lexEOF emits the accumulated data classified by the provided tokenKind and
// signals that we've reached the end of our lexing by returning `nil` instead
// of a state function.
func (l *lexer) lexEOF(k tokenKind) stateFn {

	//	l.backup()
	if l.start > l.position {
		l.ignore()
	}

	l.emit(k)
	l.emit(tokenEOF)
	return nil
}
func lexRuleElement(l *lexer) (fn stateFn) {

	r := l.next()
	switch {
	case r == eof:
		l.emit(tokenEOF)
		fn = nil
	case r == '.':
		l.emit(tokenDot)
		fn = lexRuleElement
	case r == '=':
		if l.next() == '>' {
			l.emit(tokenImplication)
			return lexRuleElement
		} else {
			l.emit(tokenError)
			return l.errorf("%s", "This should be an implication!")
		}
	case r == '<':
		if l.next() == '=' && l.next() == '>' {
			l.emit(tokenEquivalence)
			return lexRuleElement
		} else {
			l.emit(tokenError)
			return l.errorf("%s", "This should be an equivalence!")
		}
	case r == ',':
		l.emit(tokenRuleComma)
		fn = lexRuleElement
	case r == ':':
		l.emit(tokenColon)
		fn = lexRuleElement
	case r == '~':
		l.emit(tokenNegation)
		return lexRuleElement
	case unicode.IsLower(r) || r == '#':
		l.backup()
		fn = lexAtom
	case unicode.IsNumber(r) || unicode.IsUpper(r) || r == '(' || r == ')':
		l.backup()
		fn = lexConstraint
	}
	return fn
}

// lexText scans what is expected to be text.
func lexAtom(l *lexer) stateFn {
	for {
		r := l.next()
		switch {
		case r == eof:
			return l.lexEOF(tokenAtomName)
		case r == ',':
			l.backup()
			l.emit(tokenAtomName)
			l.next()
			l.emit(tokenRuleComma)
			return lexRuleElement
		case r == '[':
			l.backup()
			l.emit(tokenAtomName)
			l.next()
			l.emit(tokenAtomBracketLeft)
			l.next()
			return lexTerm
		case unicode.IsDigit(r) || unicode.IsLetter(r) || r == '_' || r == '#' :
			continue
		default:
			l.emit(tokenError)
			return l.errorf("What is this?")
		}
	}
}

func lexConstraint(l *lexer) stateFn {
	for {
		r := l.next()
		switch {
		case r == eof:
			return l.errorf("%s", "Constraint lexing should not end here.?")
		case r == '.':
			l.backup()
			l.emit(tokenConstraint)
			return lexRuleElement
		case r == ',':
			l.backup()
			l.emit(tokenConstraint)
			return lexRuleElement
		case r == ':':
			l.backup()
			l.emit(tokenConstraint)
			l.next()
			l.emit(tokenColon)
			return lexRuleElement
		}
	}
}

func lexTerm(l *lexer) stateFn {
	for {
		r := l.next()
		switch {
		case r == eof:
			return l.errorf("%s", "Term lexing should not end here!")
		case r == '.':
			l.backup()
			l.emit(tokenTermExpression)
			l.next()
			rr := l.next()
			if rr != '.' {
				return l.errorf("Double dot in Term expression missing!")
			}
			l.emit(tokenDoubleDot)
			return lexTerm
		case r == ',':
			l.backup()
			l.emit(tokenTermExpression)
			l.next()
			l.emit(tokenTermComma)
			return lexTerm
		case r == ']':
			l.backup()
			l.emit(tokenTermExpression)
			l.next()
			l.emit(tokenAtomBracketRight)
			return lexRuleElement
		}
	}
}

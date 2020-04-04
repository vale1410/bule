package lib

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"
)

func ParseProgram(fileName string) Program {
	// open a file or stream
	var scanner *bufio.Scanner
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("File not found. waiting for program in Stdin. Finish with ctrl-D")
		//scanner = bufio.NewScanner(os.Stdin)
		return Program{}
	} else {
		defer file.Close()
		scanner = bufio.NewScanner(file)
	}
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return ParseProgramFromStrings(lines)
}

func ParseProgramFromStrings(lines []string) (p Program ) {

	p.PredicateToTuples = make(map[Predicate][][]int)
	p.GroundFacts = make(map[Predicate]bool)
	p.Constants = make(map[string]int)
	p.PredicateGroundTuple = map[string]bool{}

	acc := ""
	for row, s := range lines {
		s := strings.TrimSpace(s)
		if pos := strings.Index(s, "%"); pos >= 0 {
			s = s[:pos]
		}

		s = strings.Replace(s, " ", "", -1)
		if s == "" {
			continue
		}

		if strings.HasPrefix(s, "#const") {
			s = strings.TrimPrefix(s, "#const")
			s = strings.TrimSuffix(s, ".")
			def := strings.Split(s, "=")
			asserts(len(def) == 2, s)
			term,_ := assign(Term(def[1]), p.Constants)
			asserts(groundMathExpression(term.String()), "is not ground:" + term.String())
			p.Constants[def[0]] = evaluateTermExpression(term.String())
			continue
		}

		acc += s
		if !strings.HasSuffix(s, ".") {
			continue
		}

		rule, err := parseRule(acc)
		if err != nil {
			fmt.Printf("Parsing Error in Row %v: %v\n", row, err)
			os.Exit(1)
		}
		p.Rules = append(p.Rules, rule)
		acc = ""
	}
	return p
}

func lexRule(text string) (ts Tokens) {
	lex := lex(text)
	for token := range lex.tokens {
		ts = append(ts, token)
	}
	return ts
}

// <Constraint>, <Literals> <-> head(1,2,3).
// <Constraint>, <Literals> -> head(1,2,3).
// <Constraint>, <Literals> -> head(1,2,3)?
// <ClauseDisjunction>.
// <ClauseDisjunction>?

func parseRule(text string) (rule Rule, err error) {

	// TODO: error handling
	tokens := lexRule(text)
	if tokens[len(tokens)-1].kind == tokenEOF {
		tokens = tokens[:len(tokens)-1]
	}
	rule.initialTokens = tokens
	restTokens := rule.parseEquivalenceImplicationHead()
	rule.parseRestIntoRuleElements(restTokens)
	return rule, err

}

func (rule *Rule) parseEquivalenceImplicationHead() (rest Tokens) {

	splitEquivalences := map[tokenKind]bool{tokenEquivalence: true, tokenImplication: true}
	rest, sep, head := splitIntoTwo(rule.initialTokens, splitEquivalences)
	//fmt.Println("rest:", rest.String(), "head:", head.String())
	switch sep {
	case tokenEmpty:
		rule.Typ = ruleTypeDisjunction
		return
	case tokenEquivalence:
		rule.Head = parseLiteral(head)
		rule.Typ = ruleTypeEquivalence
	case tokenImplication:
		rule.Head = parseLiteral(head)
		rule.Typ = ruleTypeImplication
	}
	return rest
}

func (rule *Rule) parseRestIntoRuleElements(tokens Tokens) {

	splitRuleElementsSeparators := map[tokenKind]bool{tokenDot: true, tokenQuestionsmark: true, tokenRuleComma: true}
	rest := splitTokens(tokens, splitRuleElementsSeparators)
	for _, sep := range rest {
		assert(len(sep.tokens) > 0)
		asserts(sep.separator.kind != tokenEmpty, sep.tokens.String(), tokens.String())

		//parse for Generators
		splitGenerator := map[tokenKind]bool{tokenColon: true}
		generator := splitTokens(sep.tokens, splitGenerator)
		if len(generator) == 1 {
			if checkIfLiteral(generator[0].tokens) {
				rule.Literals = append(rule.Literals, parseLiteral(generator[0].tokens))
			} else {
				rule.Constraints = append(rule.Constraints, parseConstraint(generator[0].tokens))
			}
		} else {
			rule.Generators = append(rule.Generators, parseGenerators(generator))
		}
	}
}

// First element is a Literal
func parseGenerators(generators []SepToken) (generator Generator) {

	for i, genElement := range generators {
		if i == 0 {
			assert(checkIfLiteral(genElement.tokens))
			generator.Head = parseLiteral(genElement.tokens)
		} else {
			if checkIfLiteral(genElement.tokens) {
				generator.Literals = append(generator.Literals, parseLiteral(genElement.tokens))
			} else {
				generator.Constraints = append(generator.Constraints, parseConstraint(genElement.tokens))
			}
		}
	}
	return
}

func checkIfLiteral(tokens Tokens) bool {
	asserts(len(tokens) > 0, "Tokens must have elements: ", tokens.String())
	return tokens[0].kind == tokenAtomName || tokens[0].kind == tokenNegation
}

// assuming it is not a constraint
// ~a4gDH[123,a*b,432-43#mod2,(123*32)-#lg(123)]
func parseLiteral(tokens Tokens) (literal Literal) {
	//fmt.Println("DEBUG", tokens.String())
	if len(tokens) == 0 {
		return
	}

	if tokens[0].kind == tokenNegation {
		literal.Neg = true
		tokens = tokens[1:]
	}

	asserts(tokens[0].kind == tokenAtomName, "Atom Structure", tokens.Debug())
	literal.Name = Predicate(tokens[0].value)

	terms := make([]Term, 0, len(tokens))
	var acc string
	for _, tok := range tokens[1:] {
		//if tokenTermMap[tok.kind] {
		//	terms = append(terms, acc)
		//	continue
		//}
		switch tok.kind {
		case tokenTermExpression,tokenDoubleDot:
			acc += tok.value
		case tokenTermComma:
			terms = append(terms, Term(acc))
			acc = ""
		case tokenAtomBracketRight:
			terms = append(terms, Term(acc))
		case tokenAtomBracketLeft:
		default:
			asserts(false, "Atom Structure:", tok.value, " ", tokens.Debug())
		}
	}
	literal.Terms = terms
	return
}

// z.B.: A*3v<=v5-2*R/7#mod3.
func parseConstraint(tokens Tokens) (constraint Constraint) {
	asserts(len(tokens) > 0, "tokens not empty!")
	left, sep, right := splitIntoTwo(tokens, tokenComparisonMap())
	asserts(len(left) > 0, "must contain left and right side")
	asserts(len(right) > 0, "must contain left and right side")
	constraint.Comparison = sep
	constraint.LeftTerm = Term(left.String())
	constraint.RightTerm = Term(right.String())
	return
}

type Tokens []Token

func (ts Tokens) Debug() string {

	sb := strings.Builder{}
	for _, token := range ts {
		switch token.kind {
		case tokenEOF:
		case tokenError:
			sb.WriteString("ERROR" + token.value)
		default:
			sb.WriteString(printToken(token.kind) + "{" + token.value + "}")
			sb.WriteString("-")
		}
	}
	return sb.String()
}

func (ts Tokens) String() string {
	sb := strings.Builder{}
	for _, token := range ts {
		sb.WriteString(token.value)
	}
	return sb.String()
}

func replaceBrackets(tokens []Token) (res []Token, err error) {

	openBrackets := 0
	for _, token := range tokens {
		switch token.kind {
		case tokenAtomBracketLeft:
			//if openBrackets > 0 {
			//	token.kind = tokenTermBracketLeft
			//}
			openBrackets++
		case tokenAtomBracketRight:
			openBrackets--
			//if openBrackets > 1 {
			//	token.kind = tokenTermBracketRight
			//}
		case tokenRuleComma:
			//if openBrackets > 1 {
			//	token.kind = tokenTermBracketRight
			//}
		}

		res = append(res, token)
		if openBrackets < 0 {
			err = errors.New(fmt.Sprintf("Wrong number of open and closing brackets!"+
				"Parsing problem with rule tokens %v \n ", res))
			return res, err
		}
	}
	return
}

func splitIntoTwo(tokens []Token, kinds map[tokenKind]bool) (left Tokens, sep tokenKind, right Tokens) {
	res := splitTokens(tokens, kinds)
	switch len(res) {
	case 0:
		assert(false)
	case 1:
		sep = res[0].separator.kind
		left = res[0].tokens
	case 2:
		sep = res[0].separator.kind
		left = res[0].tokens
		right = res[1].tokens
	default:
		asserts(false, fmt.Sprintf("More than 2 occurences Seperators. "+
			"Parsing problem with rule tokens %v with kinds %v \n ", tokens, kinds))
	}
	return left, sep, right
}

func splitTokens(tokens []Token, separator map[tokenKind]bool) (res []SepToken) {
	var acc []Token
	for _, token := range tokens {
		if separator[token.kind] {
			res = append(res, SepToken{acc, token})
			acc = []Token{}
		} else {
			acc = append(acc, token)
		}
	}
	if len(acc) > 0 {
		res = append(res, SepToken{acc, Token{}})
	}
	return
}

type SepToken struct {
	tokens    Tokens
	separator Token
}

type tokenKind int

const (
	tokenEmpty tokenKind = iota
	tokenEOF
	tokenError
	tokenAtomName          // [a-z][a-zA-Z0-9_]*
	tokenAtomBracketLeft   // [
	tokenAtomBracketRight  // ]
	tokenNegation          // ~
	tokenTermComma         // ,
	tokenRuleComma         // ,
	tokenColon             // :
	tokenEquivalence       // <->
	tokenImplication       // ->
	tokenDot               // .
	tokenQuestionsmark     // ?
	tokenDoubleDot         // ..

	tokenTermExpression
	//tokenTermModulo     // #mod
	//tokenTermLogarithm  // #log
	//tokenTermBracketLeft     // (
	//tokenTermBracketRight    // )
	//tokenTermVariable        // [A-Z][a-zA-Z0-9_]*
	//tokenTermConstant        // [a-z][a-zA-Z0-9_]*
	//tokenTermMultiplication  // *
	//tokenTermAddition        // +
	//tokenTermSubtraction     // -
	//tokenTermExponent        // **

	//tokenTermDivide          // /
	//tokenTermNumber          // 0-9

	//tokenConstraint
	tokenComparisonLT  // >
	tokenComparisonGT  // <
	tokenComparisonLE  // <=
	tokenComparisonGE  // >=
	tokenComparisonEQ  // ==
	tokenComparisonNQ  // !=

	//tokenBracketLeft   // (
	//tokenBracketRight  // )
	//tokenComma   // ,
	//tokenCurlyBracketLeft // {
	//tokenCurlyBracketRight // }
)

//func tokenTermMap() map[tokenKind]bool {
//	return map[tokenKind]bool{
//		tokenTermBracketLeft: true,
//		tokenTermBracketRight: true,
//		tokenTermVariable: true,
//		tokenTermConstant: true,
//		tokenTermMultiplication: true,
//		tokenTermAddition: true,
//		tokenTermSubtraction: true,
//		tokenTermExponent: true,
//		tokenTermModulo: true,
//		tokenTermDivide: true,
//		tokenTermLogarithmDown: true,
//		tokenTermNumber: true,
//	}
//}

func tokenComparisonMap() map[tokenKind]bool {
	return map[tokenKind]bool{
		tokenComparisonLT: true,
		tokenComparisonGT: true,
		tokenComparisonLE: true,
		tokenComparisonGE: true,
		tokenComparisonEQ: true,
		tokenComparisonNQ: true,
	}
}

func printToken(kind tokenKind) (s string) {
	switch kind {
	case tokenAtomName:
		s = "ATOM"
	case tokenAtomBracketLeft:
		s = "AtomBL"
	case tokenAtomBracketRight:
		s = "AtomBR"
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
	//case tokenConstraint:
	//	s = "CONSTRAINT"
	case tokenEquivalence:
		s = "EQUIVALENCE"
	case tokenImplication:
		s = "IMPLICATION"
	case tokenDot:
		s = "DOT"
	case tokenDoubleDot:
		s = "DOUBLEDOT"
	case tokenComparisonLT:
		s = "LT"
	case tokenComparisonGT:
		s = "GT"
	case tokenComparisonLE:
		s = "LE"
	case tokenComparisonGE:
		s = "GE"
	case tokenComparisonEQ:
		s = "EQ"
	case tokenComparisonNQ:
		s = "QN"
	default:
		asserts(false, "not implemented tokentype:", fmt.Sprintf("%+v", kind))
	}
	return
}

// Token is accumulated while lexing the provided input, and emitted over a
// channel to the parser.
type Token struct {
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
		kind:  k,
		value: accumulation,
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
	case r == '-':
		if l.next() == '>' {
			l.emit(tokenImplication)
			return lexRuleElement
		} else {
			l.backup()
			fn = lexConstraintLeft
		}
	case r == '<':
		if l.next() == '-' && l.next() == '>' {
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
		fn = lexRuleElement
	case unicode.IsLower(r) || r == '#':
		l.backup()
		fn = lexAtom
	case unicode.IsNumber(r) || unicode.IsUpper(r) || r == '(' || r == ')':
		l.backup()
		fn = lexConstraintLeft
	default:
		l.emit(tokenError)
		return l.errorf("What is this?")
	}
	return fn
}

func lexAtom(l *lexer) stateFn {

	r := l.next()
	asserts(unicode.IsLower(r) || r == '#', "Not correct Atom parsing:", l.input)

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
			return lexTermInAtom
		case unicode.IsDigit(r) || unicode.IsLetter(r) || r == '_':
			continue
		default:
			l.emit(tokenError)
			return l.errorf("What is this?")
		}
	}
}

func lexConstraintLeft(l *lexer) stateFn {
	for {
		r := l.next()
		switch {
		case r == eof:
			return l.errorf("%s", "Constraint lexing should not end here.?")
		//case r == ':', r == '.': // Global Variable!
		//	l.backup()
		//	l.emit(tokenTermExpression)
		//	return lexRuleElement(l)
		case r == '!':
			l.backup()
			l.emit(tokenTermExpression)
			l.next()
			rr := l.next()
			if rr == '=' {
				l.emit(tokenComparisonNQ)
				return lexConstraintRight(l)
			} else {
				return l.errorf("%s", "Constraint lexing should not end here.?")
			}
		case r == '<':
			l.backup()
			l.emit(tokenTermExpression)
			l.next()
			rr := l.next()
			if rr == '=' {
				l.emit(tokenComparisonLE)
				return lexConstraintRight(l)
			} else {
				l.backup()
				l.emit(tokenComparisonLT)
				return lexConstraintRight(l)
			}
		case r == '>':
			l.backup()
			l.emit(tokenTermExpression)
			l.next()
			rr := l.next()
			if rr == '=' {
				l.emit(tokenComparisonGE)
				return lexConstraintRight(l)
			} else {
				l.backup()
				l.emit(tokenComparisonGT)
				return lexConstraintRight(l)
			}
		case r == '=':
			l.backup()
			l.emit(tokenTermExpression)
			l.next()
			rr := l.next()
			if rr == '=' {
				l.emit(tokenComparisonEQ)
				return lexConstraintRight(l)
			} else {
				return l.errorf("%s", "Constraint lexing should not end here.?")
			}
		//case r == '#': TODO SPECIAL
		//	return lexSpecialFn(l, lexConstraintLeft)
		case isTermExpressionRune(r):
			continue
		default:
			return l.errorf("%v%v.", "Lexing Problem. What is this? ", string(r))
		}
	}
}

// TODO SPECIAL
//func lexSpecialFn(l *lexer, fn stateFn) stateFn {
//	r1 := l.next()
//	r2 := l.next()
//	r3 := l.next()
//	if r1 == 'l' && r2 == 'o' && r3 == 'g' {
//		l.emit(tokenTermLogarithm)
//		return fn
//	} else if r1 == 'm' && r2 == 'o' && r3 == 'd' {
//		l.emit(tokenTermModulo)
//		return fn
//	} else {
//		return l.errorf("Wrong special function")
//	}
//}

func lexConstraintRight(l *lexer) stateFn {
	for {
		r := l.next()
		switch {
		case r == eof:
			return l.errorf("%s", "Constraint lexing should not end here.?")
		case isTermExpressionFinish(r):
			l.backup()
			l.emit(tokenTermExpression)
			return lexRuleElement
		//case r == '#': TODO SPECIAL
		//	l.backup()
		//	l.emit(tokenTermExpression)
		//	l.next()
		//	return lexSpecialFn(l, lexConstraintRight)
		case isTermExpressionRune(r):
			continue
		default:
			return l.errorf("rune:\"%v\". %v", string(r), "Unrecognised TermExpression.?")
		}
	}
}

func isTermExpressionFinish(r rune) bool {
	return strings.ContainsRune(",:.", r)
}

func isTermExpressionRune(r rune) bool {
	return r == '_' || unicode.IsNumber(r) || unicode.IsLetter(r) || strings.ContainsRune("#()*/-+", r)
}

func lexTermInAtom(l *lexer) stateFn {
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
			return lexTermInAtom
		case r == ',':
			l.backup()
			l.emit(tokenTermExpression)
			l.next()
			l.emit(tokenTermComma)
			return lexTermInAtom
		case r == ']':
			l.backup()
			l.emit(tokenTermExpression)
			l.next()
			l.emit(tokenAtomBracketRight)
			return lexRuleElement
		case isTermExpressionRune(r):
			continue
		default:
			return l.errorf("rune:\"%v\". %v", string(r), "Unrecognised TermExpression.?")
		}
	}
}
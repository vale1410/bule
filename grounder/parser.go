package grounder

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"
)

func ParseProgram(fileNames []string) (Program, error) {
	Debug(2, "opening files:", fileNames)
	var lines []string
	where := make([]LineNumberInfo, 0)
	i := 0
	for _, fileName := range fileNames {
		var scanner *bufio.Scanner
		file, err := os.Open(fileName)
		if err != nil {
			return Program{}, err
		}
		scanner = bufio.NewScanner(file)
		line := 1
		for scanner.Scan() {
			where = append(where, LineNumberInfo{fileName, line})
			lines = append(lines, scanner.Text())
			i++
			line++
		}
		file.Close()
	}
	return ParseProgramFromStrings(lines, where)
}

func ParseProgramFromStrings(lines []string, where []LineNumberInfo) (p Program, err error) {

	p.PredicateToTuples = make(map[Predicate][][]string)
	p.FinishCollectingFacts = make(map[Predicate]bool)
	p.Constants = make(map[string]string)
	p.PredicateTupleMap = map[string]bool{}

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
			defs := strings.Split(s, ",")
			for _, def := range defs {
				def := strings.Split(def, "=")
				asserts(len(def) == 2, s)
				term, _, err := assign(Term(def[1]), p.Constants)
				if err != nil {
					return p, err
				}
				asserts(groundMathExpression(term.String()), "is not ground:"+term.String())
				// Just checking that it's a ground expression ! We do not allow
				// non ground stuff in #const expressions.
				_, err = evaluateTermExpression(term.String())
				if err != nil {
					return p, err
				}
				p.Constants[def[0]] = term.String()
			}
			continue
		}

		acc += s

		if !(strings.HasSuffix(s, ".") || strings.HasSuffix(s, "?")) {
			continue
		}

		// Split into mulitple rules if we use single dot somewhere. E.g. x[1].y[1..3].
		scanner := bufio.NewScanner(strings.NewReader(acc))
		scanner.Split(splitBuleRule)
		var rulesString []string
		for scanner.Scan() {
			rulesString = append(rulesString, scanner.Text())
		}

		for _, tmp := range rulesString {
			rule, err := parseRule(tmp)
			if err != nil {
				return p, fmt.Errorf("Parsing Error in row %v in file %v:\n %w\n in rule: %s", where[row].line, where[row].fileName, err, tmp)
			}
			rule.LineNumber = where[row]
			p.Rules = append(p.Rules, rule)
		}
		acc = ""
	}
	return p, nil
}

func splitBuleRule(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	// Scan until space, marking end of word.
	for width, i := 0, 0; i < len(data); i += width {
		var r rune
		r, width = utf8.DecodeRune(data[i:])
		if r == '.' {
			if len(data) == i+width { // Last element in data
				return i + width, data[:i+width], nil
			} else {
				r2, width2 := utf8.DecodeRune(data[i+width:])
				if r2 == '.' {
					width += width2
				} else {
					return i + width, data[:i+width], nil
				}
			}
		}
	}
	// If we're at EOF, we have a final, non-empty, non-terminated word. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

func lexRule(text string) (ts Tokens) {
	lex := lex(text)
	for token := range lex.tokens {
		ts = append(ts, token)
	}
	return ts
}

// EBNF:
// Rule = [ Quantification ',' ] [ Guard '::' ] Definition ('.' | '?' )
// Generator = (Fact | Constraint ) ( ',' Guard | )
// Constraint = Expression Comparison Expression
// Definition =  ( Fact | Implication )
// Implication = [Conjunction '->'] Disjunction
// Conjunction = [Guard ':' ] Literal [ '&' Conjunction ]
// Disjunction = [Guard ':' ] Literal [ '|' Disjunction ]
// Fact = [ '~' ] Predicate [ '[' Terms ']' ]
// Literal = [ '~' ] Predicate [ '(' Terms ')' ]
// Terms = Term [ ',' Terms ]
// Predicate = starts with lower case letter
// Variable = starts with upper case letter
// TermConstant = [_a-z][a-zA-Z0-9_]starts with lower case letter
// Number = [0-9]+
// Expression = (Expression) | Expression BinaryOperation Expression | Number | Variable |
// BinaryOperation = '+' | '*' | '/' | '-' | '#mod'
// UnaryOperation = '+' | '-'
// TODO FOR THE FUTURE
// PseudoBoolean = Number '{' Linear '}' Number
// Linear = [Guard ':' ] [Coefficient '*'] Literal [ ';' Linear ]
func parseRule(text string) (rule Rule, err error) {

	tokens := lexRule(text)
	last := tokens[len(tokens)-1]
	if last.kind == tokenError {
		return rule, fmt.Errorf("Lexing Error in %v \n in rule %v", last.value, text)
	}
	if last.kind == tokenEOF {
		tokens = tokens[:len(tokens)-1]
	}
	rule.initialTokens = tokens

	splitDoubleColon := map[tokenKind]bool{tokenDoubleColon: true}
	left, sep, right := splitIntoTwo(tokens, splitDoubleColon)
	switch sep {
	case tokenEmpty:
		err := rule.parseImplication(left)
		if err != nil {
			return rule, err
		}
	case tokenDoubleColon:
		rule.Generators, rule.Constraints, err = parseGuard(left)
		if err != nil {
			return rule, err
		}
		err = rule.parseImplication(right)
		if err != nil {
			return rule, err
		}
	default:
		return rule, fmt.Errorf("Problems parsing rule %v.", rule.Debug())
	}
	return rule, err
}

func parseGuard(tokens Tokens) (generators []Literal, constraints []Constraint, err error) {

	splitRuleElementsSeparators := map[tokenKind]bool{token2RuleComma: true}
	elements := splitTokens(tokens, splitRuleElementsSeparators)
	if len(elements) == 0 || len(elements[0].tokens) == 0 {
		return generators, constraints, fmt.Errorf("Could not parse guard: %v  ", tokens.String())
	}
	for _, element := range elements {
		if !(element.separator.kind == tokenEmpty || element.separator.kind == token2RuleComma) {
			return generators, constraints, fmt.Errorf("Wrong separator %v in guard: %v  ", element.separator.value, tokens.String())
		}
		if checkIfLiteral(element.tokens) {
			lit, err := parseLiteral(element.tokens)
			if err != nil {
				return generators, constraints, fmt.Errorf("Could not parse literal: %v in guard: %v  ", element.tokens, tokens.String())
			}
			if lit.Fact == false {
				return generators, constraints, fmt.Errorf("literal %v in guard %v is not a fact. Please use brackets [...] instead of (...)", element.tokens.String(), tokens.String())
			}
			generators = append(generators, lit)
		} else {
			constraint, err := parseConstraint(element.tokens)
			if err != nil {
				return generators, constraints, fmt.Errorf("Could not parse constraint %v in guard %v", element.tokens.String(), tokens.String())
			}
			constraints = append(constraints, constraint)
		}
	}
	return generators, constraints, nil
}

func (rule *Rule) parseImplication(tokens Tokens) (err error) {

	splitDoubleColon := map[tokenKind]bool{tokenImplication: true}
	left, sep, right := splitIntoTwo(tokens, splitDoubleColon)
	switch sep {
	case tokenEmpty:
		err := rule.parseDisjunction(left)
		if err != nil {
			return err
		}
	case tokenImplication:
		err = rule.parseConjunction(left)
		if err != nil {
			return err
		}
		err = rule.parseDisjunction(right)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("problems parsing rule %v.", tokens)
	}
	return err
}

func (rule *Rule) parseConjunction(tokens Tokens) (err error) {

	splitSet := map[tokenKind]bool{
		tokenAmpersand: true,
	}
	elements := splitTokens(tokens, splitSet)
	var iterator Iterator
	for i, element := range elements {
		if i == len(elements)-1 && element.separator.kind != tokenEmpty {
			return fmt.Errorf("element %v: %v in left side of implication %v has problems ", i, element.tokens.String(), tokens.String())
		} else if element.separator.kind != tokenAmpersand {
			return fmt.Errorf("element %v: %v in left side of implication %v has wrong seperator, should be &", i, element.tokens.String(), tokens.String())
		}
		iterator, err = parseIterator(element.tokens)
		if err != nil {
			return err
		}
		iterator.Head = iterator.Head.createNegatedLiteral()
		rule.Iterators = append(rule.Iterators, iterator)
	}
	return
}

func (rule *Rule) parseDisjunction(tokens Tokens) (err error) {

	splitSet := map[tokenKind]bool{
		tokenPipe:          true,
		tokenQuestionsmark: true,
		tokenDot:           true,
	}
	elements := splitTokens(tokens, splitSet)
	var iterator Iterator
	for i, element := range elements {
		if i == len(elements)-1 {
			if !(element.separator.kind == tokenQuestionsmark || element.separator.kind == tokenDot) {
				return fmt.Errorf("rule should end with either dot or questions mark but ended with %+v", printToken(element.separator.kind))
			}
		} else if element.separator.kind != tokenPipe {
			return fmt.Errorf("element %v: %v has wrong seperator, should be | in disjunction %v", i, element.tokens.String(), tokens.String())
		}
		iterator, err = parseIterator(element.tokens)
		if err != nil {
			return err
		}
		if element.separator.kind == tokenQuestionsmark {
			rule.IsQuestionMark = true
		}
		rule.Iterators = append(rule.Iterators, iterator)
	}
	return
}

// Last element is a Literal
// <Fact[]>, .. ,  : literal
func parseIterator(tokens Tokens) (iterator Iterator, err error) {

	splitColon := map[tokenKind]bool{tokenColon: true}
	left, sep, right := splitIntoTwo(tokens, splitColon)
	switch sep {
	case tokenEmpty:
		iterator.Head, err = parseLiteral(left)
		if err != nil {
			return
		}
	case tokenColon:
		iterator.Conditionals, iterator.Constraints, err = parseGuard(left)
		if err != nil {
			return
		}
		iterator.Head, err = parseLiteral(right)
		if err != nil {
			return
		}
	default:
		return iterator, fmt.Errorf("cannot parse iterator %v", tokens.String())
	}
	return
}

func checkIfLiteral(tokens Tokens) bool {
	asserts(len(tokens) > 0, "Tokens must have elements: ", tokens.String())
	return tokens[0].kind == token2AtomName || tokens[0].kind == tokenNegation
}

// assuming it is not a constraint
// ~a4gDH[123,a*b,432-43#mod2,(123*32)-#lg(123)]
func parseLiteral(tokens Tokens) (literal Literal, err error) {
	if len(tokens) == 0 {
		return literal, fmt.Errorf("Literal definition is empty")
	}

	if tokens[0].kind == tokenNegation {
		literal.Neg = true
		tokens = tokens[1:]
	}

	if tokens[0].kind != token2AtomName {
		return literal, fmt.Errorf("predicate %v in literal %v definition is incorrect: ", tokens[0], tokens)
	}

	literal.Name = Predicate(tokens[0].value)

	terms := make([]Term, 0, len(tokens))
	var acc string
	for _, tok := range tokens[1:] {
		switch tok.kind {
		case token2TermExpression, tokenDoubleDot:
			acc += tok.value
		case token2TermComma:
			terms = append(terms, Term(acc))
			acc = ""
		case tokenAtomParanthesisRight:
			if literal.Fact {
				return literal, fmt.Errorf("closing paranthesis does not match fact type: %v", tokens)
			}
			terms = append(terms, Term(acc))
		case tokenAtomBracketRight:
			if !literal.Fact {
				return literal, fmt.Errorf("closing brackets does not match search literal: %v", tokens)
			}
			terms = append(terms, Term(acc))
		case tokenAtomBracketLeft:
			literal.Fact = true
		case tokenAtomParanthesisLeft:
			literal.Fact = false
		default:
			return literal, fmt.Errorf("problem parsing literal: %v", tokens)
		}
	}
	literal.Terms = terms
	return
}

// z.B.: A*3v<=v5-2*R/7#mod3.
func parseConstraint(tokens Tokens) (constraint Constraint, err error) {
	if len(tokens) == 0 {
		return constraint, fmt.Errorf("constraint definition is empty")
	}
	left, sep, right := splitIntoTwo(tokens, tokenComparisonMap())
	if len(left) == 0 || len(right) == 0 {
		return constraint, fmt.Errorf("constraint definition is not formed correctly %v", tokens.String())
	}
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
			//	token.kind = token2TermBracketLeft
			//}
			openBrackets++
		case tokenAtomBracketRight:
			openBrackets--
			//if openBrackets > 1 {
			//	token.kind = token2TermBracketRight
			//}
		case token2RuleComma:
			//if openBrackets > 1 {
			//	token.kind = token2TermBracketRight
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
		fmt.Println(res)
		asserts(false, fmt.Sprintf("More than 2 occurences Seperators. "+
			"Parsing problem with rule tokens %v with kinds %v \n ", tokens, kinds))
	}
	return left, sep, right
}

// Splits token slice into:
// t0,t0,t0, sep, t1,t1,.. sep,sep tntntn sep
// into [[t0,t0,t0], sep], [[t1,t1,.. ],sep],[[],sep], [[tntntn], sep]
// - tokens are empty slice if first element is sep, or two sep in sequence
// - sep is tokenEmpty when last sep is not existing
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
	token2AtomName            // [a-z][a-zA-Z0-9_]*
	tokenAtomParanthesisLeft  // (
	tokenAtomParanthesisRight // )
	tokenAtomBracketLeft      // [
	tokenAtomBracketRight     // ]
	tokenNegation             // ~
	tokenPipe                 // |
	tokenAmpersand            // &
	token2TermComma           // ,
	token2RuleComma           // ,
	tokenColon                // :
	tokenEquivalence          // <->
	tokenSimpleImplication    // ->
	tokenImplication          // =>
	tokenDoubleColon          // ::
	tokenDot                  // .
	tokenQuestionsmark        // ?
	tokenDoubleDot            // ..

	token2TermExpression

	tokenComparisonLT // >
	tokenComparisonGT // <
	tokenComparisonLE // <=
	tokenComparisonGE // >=
	tokenComparisonEQ // ==
	tokenComparisonNQ // !=

	///  // This technical set is important for phase one parsing!
	///  tokenIdentifier// ,
	///  tokenBracketLeft   // (
	///  tokenBracketRight  // )
	///  tokenComma   // ,
	///  //tokenCurlyBracketLeft // {
	///  //tokenCurlyBracketRight // }

	///  tokenTermBracketLeft   // (
	///  tokenTermBracketRight  // )
	///  tokenTermVariable      // [A-Z][a-zA-Z0-9_]*
	///  tokenTermConstant      // [a-z][a-zA-Z0-9_]*
	///  tokenTermMultiplication // *
	///  tokenTermAddition       // +
	///  tokenTermSubtraction    // -
	///  tokenTermExponent       // ** or  ^
	///  tokenTermDivide         // /
	///  tokenTermNumber         // 0-9

	///  tokenTermModulo         // #mod
	///  tokenTermLogarithm      // #log
)

//func tokenTermMap() map[tokenKind]bool {
//	return map[tokenKind]bool{
//		token2TermBracketLeft: true,
//		token2TermBracketRight: true,
//		token2TermVariable: true,
//		token2TermConstant: true,
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
	case token2AtomName:
		s = "ATOM"
	case tokenAtomBracketLeft:
		s = "AtomBL"
	case tokenAtomBracketRight:
		s = "AtomBR"
	case tokenNegation:
		s = "NEGATION"
	case tokenAmpersand:
		s = "AMPERSAND"
	case tokenPipe:
		s = "PIPE"
	case token2TermComma:
		s = "TERMCOMMA"
	case token2RuleComma:
		s = "RULECOMMA"
	case tokenColon:
		s = "COLON"
	case token2TermExpression:
		s = "TERM"
	case tokenEquivalence:
		s = "EQUIVALENCE"
	case tokenDoubleColon:
		s = "DOUBLECOLON"
	case tokenSimpleImplication:
		s = "SIMPLEIMPLICATION"
	case tokenImplication:
		s = "IMPLICATION"
	case tokenQuestionsmark:
		s = "QUESTIONMARK"
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
	case tokenAtomParanthesisLeft:
		s = "ATOMPL"
	case tokenAtomParanthesisRight:
		s = "ATOMPR"
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
	input            string // we'll store the string being parsed
	start            int    // the position we started scanning
	position         int    // the current position of our scan
	width            int    // we'll be using runes which can be double byte
	paranthesisStack int    //  number of open paranthesis  ( ( are closed by ) )
	//bracketStack int  // number of open brackets [ [  are closed via ] ] ]
	state  stateFn    // the current state function
	tokens chan Token // the channel we'll use to communicate between the lexer and the parser
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
	case r == '?':
		l.emit(tokenQuestionsmark)
		fn = lexRuleElement
	case r == '.':
		l.emit(tokenDot)
		fn = lexRuleElement
	case r == '=':
		if l.next() == '>' {
			l.emit(tokenImplication)
			return lexRuleElement
		} else {
			l.emit(tokenError)
			return l.errorf("%s", "This should be an implication")
		}
	case r == '-':
		if l.next() == '>' {
			l.emit(tokenSimpleImplication)
			fn = lexRuleElement
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
	case r == '&':
		l.emit(tokenAmpersand)
		fn = lexRuleElement
	case r == '|':
		l.emit(tokenPipe)
		fn = lexRuleElement
	case r == ',':
		l.emit(token2RuleComma)
		fn = lexRuleElement
	case r == ':':
		if l.next() == ':' {
			l.emit(tokenDoubleColon)
			fn = lexRuleElement
		} else {
			l.backup()
			l.emit(tokenColon)
			fn = lexRuleElement
		}
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
			return l.lexEOF(token2AtomName)
		case r == '=':
			l.backup()
			l.emit(token2AtomName)
			return lexRuleElement
		case r == '.':
			l.backup()
			l.emit(token2AtomName)
			l.next()
			l.emit(tokenDot)
			return lexRuleElement
		case r == '?':
			l.backup()
			l.emit(token2AtomName)
			l.next()
			l.emit(tokenQuestionsmark)
			return lexRuleElement
		case r == '&':
			l.backup()
			l.emit(token2AtomName)
			l.next()
			l.emit(tokenAmpersand)
			return lexRuleElement
		case r == '|':
			l.backup()
			l.emit(token2AtomName)
			l.next()
			l.emit(tokenPipe)
			return lexRuleElement
		case r == ',':
			l.backup()
			l.emit(token2AtomName)
			l.next()
			l.emit(token2RuleComma)
			return lexRuleElement
		case r == '(':
			l.backup()
			l.emit(token2AtomName)
			l.next()
			l.emit(tokenAtomParanthesisLeft)
			//l.next()
			return lexTermInAtom
		case r == '[':
			l.backup()
			l.emit(token2AtomName)
			l.next()
			l.emit(tokenAtomBracketLeft)
			//l.next()
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
		//	l.emit(token2TermExpression)
		//	return lexRuleElement(l)
		case r == '!':
			l.backup()
			l.emit(token2TermExpression)
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
			l.emit(token2TermExpression)
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
			l.emit(token2TermExpression)
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
			l.emit(token2TermExpression)
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

func lexConstraintRight(l *lexer) stateFn {
	for {
		r := l.next()
		switch {
		case r == eof:
			return l.errorf("%s", "Constraint lexing should not end here.")
			//		case r == '-':
			//			if l.next() == '>' {
			//				l.backup()
			//				l.backup()
			//				l.emit(token2TermExpression)
			//				return lexRuleElement
			//			} else {
			//				l.backup()
			//			}
		case r == '.':
			l.next()
			rr := l.next()
			if rr != '.' {
				continue
			} else {
				l.backup()
				l.emit(token2TermExpression)
				return lexRuleElement
			}
		case isTermExpressionFinish(r):
			l.backup()
			l.emit(token2TermExpression)
			return lexRuleElement
		case isTermExpressionRune(r):
			continue
		default:
			return l.errorf("rune:\"%v\". %v", string(r), "Unrecognised TermExpression.?")
		}
	}
}

func isTermExpressionFinish(r rune) bool {
	return strings.ContainsRune(",:", r)
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
			l.emit(token2TermExpression)
			l.next()
			rr := l.next()
			if rr != '.' {
				return l.errorf("The second of the double dot in Term expression is missing!")
			}
			l.emit(tokenDoubleDot)
			return lexTermInAtom
		case r == ',':
			l.backup()
			l.emit(token2TermExpression)
			l.next()
			l.emit(token2TermComma)
			return lexTermInAtom
		case r == ']':
			l.backup()
			l.emit(token2TermExpression)
			l.next()
			l.emit(tokenAtomBracketRight)
			return lexRuleElement
		//case r == ')' && l.paranthesisStack == 0 : ERROR
		case r == '(':
			l.paranthesisStack++
			continue
		case r == ')' && l.paranthesisStack >= 1:
			l.paranthesisStack--
			continue
		case r == ')' && l.paranthesisStack == 0:
			l.backup()
			l.emit(token2TermExpression)
			l.next()
			l.emit(tokenAtomParanthesisRight)
			return lexRuleElement
		case isTermExpressionRune(r):
			continue
		default:
			return l.errorf("rune:\"%v\". %v", string(r), "Unrecognised TermExpression.?")
		}
	}
}

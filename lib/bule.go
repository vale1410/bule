package lib

import (
	"fmt"
	"github.com/Knetic/govaluate"
	"github.com/scylladb/go-set/strset"
	"strconv"
	"strings"
	"unicode"
)

func (p *Program) ConstraintSimplification() (bool, error) {

	finalChanged := true

	i := 0
	for {
		i++
		changed, err := p.TransformConstraintsToInstantiation()
		if err != nil {
			return true, fmt.Errorf("Constraint simplification, iteration %v. \n %w", i, err)
		}
		if !changed {
			//Debug(2, "Remove clauses with contradictions, e.g.  (1==2) or (1!=1),  and remove true constraints, e.g.  (1>2, 1==1).")
			finalChanged, err = p.CleanRulesFromGroundBoolExpression()
			if err != nil {
				return true, fmt.Errorf("Remove of clauses failed %v. \n %w", i, err)
			}
			break
		}
	}
	return finalChanged || i > 1, nil
}

func (p *Program) ExpandGroundRanges() (changed bool, err error) {
	check := func(term Term) bool {
		interval := strings.Split(string(term), "..")
		return len(interval) == 2 && groundMathExpression(interval[0]) && groundMathExpression(interval[1])
	}
	transform := func(term Term) (newTerms []Term, err error) {
		interval := strings.Split(string(term), "..")
		i1, err := evaluateTermExpression(interval[0])
		if err != nil {
			return newTerms, err
		}
		i2, err := evaluateTermExpression(interval[1])
		if err != nil {
			return newTerms, err
		}
		for _, newValue := range makeSet(i1, i2) {
			newTerms = append(newTerms, Term(strconv.Itoa(newValue)))
		}
		return
	}
	return p.TermExpansion(check, transform)
}

// This resolves facts with clauses.
func (p *Program) InstantiateExplicitNonGroundLiterals() (changed bool, err error) {
	// Find rule with non-ground literal that is going to be rolled out
	check := func(r Rule) bool {
		for _, lit := range r.Literals {
			if p.Explicit[lit.Name] && !lit.FreeVars().IsEmpty() {
				return true
			}
		}
		return false
	}

	transform := func(rule Rule) (generatedRules []Rule, err error) {

		var litNG Literal // First non-Ground literal
		var i int
		for i, litNG = range rule.Literals {
			if p.Explicit[litNG.Name] && !litNG.FreeVars().IsEmpty() {
				break
			}
		}

		for _, tuple := range p.findFilteredTuples(litNG) {
			newRule := rule.Copy()
			for j, val := range tuple {
				newRule.Literals[i].Terms[j] = Term(strconv.Itoa(val))
				newConstraint := Constraint{
					LeftTerm:   litNG.Terms[j],
					RightTerm:  Term(strconv.Itoa(val)),
					Comparison: tokenComparisonEQ,
				}
				newRule.Constraints = append(newRule.Constraints, newConstraint)
			}
			generatedRules = append(generatedRules, newRule)
		}
		return generatedRules, err
	}
	return p.RuleExpansion(check, transform)
}

// This resolves facts with clauses.
func (p *Program) InstantiateNonGroundLiterals() (changed bool, err error) {
	// Find rule with non-ground literal that is going to be rolled out
	check := func(r Rule) bool {
		for _, lit := range r.Literals {
			if !lit.FreeVars().IsEmpty() {
				return true
			}
		}
		return false
	}

	transform := func(rule Rule) (generatedRules []Rule, err error) {

		var litNG Literal // First non-Ground literal
		var i int
		for i, litNG = range rule.Literals {
			if !litNG.FreeVars().IsEmpty() {
				break
			}
		}

		for _, tuple := range p.findFilteredTuples(litNG) {
			newRule := rule.Copy()
			for j, val := range tuple {
				newRule.Literals[i].Terms[j] = Term(strconv.Itoa(val))
				newConstraint := Constraint{
					LeftTerm:   litNG.Terms[j],
					RightTerm:  Term(strconv.Itoa(val)),
					Comparison: tokenComparisonEQ,
				}
				newRule.Constraints = append(newRule.Constraints, newConstraint)
			}
			generatedRules = append(generatedRules, newRule)
		}
		return generatedRules, err
	}
	return p.RuleExpansion(check, transform)
}

// given a literal p(X,4,2,Y), a simple and quick way to find all tuples that fulfil this!
func (p *Program) findFilteredTuples(literal Literal) [][]int {
	positions, values := literal.findGroundTerms()
	filteredTuples := make([][]int, 0, len(p.PredicateToTuples[literal.Name]))
	for _, tuple := range p.PredicateToTuples[literal.Name] {
		good := true
		for i, p := range positions {
			if tuple[p] != values[i] {
				good = false
				break
			}
		}
		if good {
			filteredTuples = append(filteredTuples, tuple)
		}
	}
	return filteredTuples
}

// This resolves facts with clauses.
func (p *Program) InstantiateAndRemoveFactFromIterator() (changed bool, err error) {
	// Find iterator with fact that we can replace!
	check := func(r Rule) bool {
		for _, iter := range r.Iterators {
			for _, lit := range iter.Conditionals {
				if p.FinishCollectingFacts[lit.Name] && lit.Neg == false {
					return true
				}
			}
		}
		return false
	}

	transform := func(rule Rule) (Rule, error) {

		var fact Literal
		var iter Iterator
		var i int
		var j int
		found := false
		for i, iter = range rule.Iterators {
			for j, fact = range iter.Conditionals {
				if p.FinishCollectingFacts[fact.Name] && fact.Neg == false {
					found = true
					break
				}
			}
			if found {
				break
			}
		}

		newIterators := make([]Iterator, i)
		copy(newIterators, rule.Iterators[:i])

		newIterator := rule.Iterators[i].Copy()
		newIterator.Conditionals = append(newIterator.Conditionals[:j],
			newIterator.Conditionals[j+1:]...)

		for _, tuples := range p.findFilteredTuples(fact) {
			tmpIterator := newIterator.Copy()
			for k, val := range tuples {
				newConstraint := Constraint{
					LeftTerm:   fact.Terms[k],
					RightTerm:  Term(strconv.Itoa(val)),
					Comparison: tokenComparisonEQ,
				}
				tmpIterator.Constraints = append(tmpIterator.Constraints, newConstraint)
			}
			newIterators = append(newIterators, tmpIterator)
		}
		newIterators = append(newIterators, rule.Iterators[i+1:]...)
		rule.Iterators = newIterators
		return rule, nil
	}
	return p.RuleTransformation(check, transform)
}

// This resolves facts with clauses.
func (p *Program) InstantiateAndRemoveFactFromGenerator() (changed bool, err error) {
	// Find rule with fact
	check := func(r Rule) bool {
		for _, lit := range r.Generators {
			if p.FinishCollectingFacts[lit.Name] && lit.Neg == false {
				return true
			}
		}
		return false
	}

	transform := func(rule Rule) (generatedRules []Rule, err error) {

		var fact Literal
		var i int
		for i, fact = range rule.Generators {
			if p.FinishCollectingFacts[fact.Name] && fact.Neg == false {
				break
			}
		}
		rule.Generators = append(rule.Generators[:i], rule.Generators[i+1:]...)

		for _, tuple := range p.findFilteredTuples(fact) {
			newRule := rule.Copy()
			newRule.Parent = &rule
			for j, val := range tuple {
				newConstraint := Constraint{
					LeftTerm:   fact.Terms[j],
					RightTerm:  Term(strconv.Itoa(val)),
					Comparison: tokenComparisonEQ,
				}
				newRule.Constraints = append(newRule.Constraints, newConstraint)
			}
			generatedRules = append(generatedRules, newRule)
		}
		return
	}
	return p.RuleExpansion(check, transform)
}

func (p *Program) InsertLiteralTuple(lit Literal) error {
	groundTerms, err := evaluateExpressionTuples(lit.Terms)
	if err != nil {
		return err
	}
	if !p.PredicateTupleMap[lit.String()] &&
		!p.PredicateTupleMap[lit.createNegatedLiteral().String()] {
		p.PredicateToTuples[lit.Name] = append(p.PredicateToTuples[lit.Name], groundTerms)
	}
	p.PredicateTupleMap[lit.String()] = true
	p.PredicateTupleMap[lit.createNegatedLiteral().String()] = true
	return nil
}

func (p *Program) RemoveNegatedGroundGenerator() (changed bool, err error) {
	removeIfTrue := func(literal Literal) bool {
		if p.FinishCollectingFacts[literal.Name] &&
			literal.IsGround() &&
			literal.Neg == true &&
			p.PredicateTupleMap[literal.String()] == false {
			return true
		}
		return false
	}

	for i, r := range p.Rules {
		for j, g := range r.Generators {
			if removeIfTrue(g) {
				changed = true
				p.Rules[i].Generators = append(p.Rules[i].Generators[:j],
					p.Rules[i].Generators[j+1:]...)
				break
			}
		}
	}

	return
}

func (p *Program) RemoveRulesWithNegatedGroundGenerator() (changed bool, err error) {
	removeIfTrue := func(r Rule) bool {
		for _, literal := range r.Generators {
			if p.FinishCollectingFacts[literal.Name] &&
				literal.IsGround() &&
				literal.Neg == true &&
				p.PredicateTupleMap[literal.createNegatedLiteral().String()] == true {
				return true
			}
		}
		return false
	}
	return p.RemoveRules(removeIfTrue)
}

// // A fact is fully collected if it does not occur as a head in any rule.
// func (p *Program) DetermineIfFactIsFullyCollected() (changed bool, err error) {
// 	inHead := make(map[Predicate]bool,0)
// 	p.FinishCollectingFacts[lit.Name] = true
//
// }

func (p *Program) CollectGroundFacts() (changed bool, err error) {

	check := func(r Rule) bool {
		return len(r.Literals) == 1 &&
			r.Literals[0].Fact &&
			len(r.Iterators) == 0 &&
			len(r.Constraints) == 0 &&
			len(r.Generators) == 0 &&
			r.FreeVars().IsEmpty()
	}
	transform := func(rule Rule) (empty []Rule, err error) {
		lit := rule.Literals[0]
		res, err := evaluateExpressionTuples(lit.Terms)
		if err != nil {
			return empty, RuleError{
				rule,
				"Collect Ground Facts Problem",
				err,
			}
		}
		if !p.PredicateTupleMap[lit.String()] {
			p.PredicateToTuples[lit.Name] = append(p.PredicateToTuples[lit.Name], res)
			p.PredicateTupleMap[lit.String()] = true
			p.FinishCollectingFacts[lit.Name] = true
		}
		return
	}
	return p.RuleExpansion(check, transform)
}

// Checks if constraint is of the form X==<math>, or <math>==X
// It also does very simple equation solving for equations with one variable, like X-3+1==<math> .
func (constraint Constraint) IsInstantiation() (is bool, variable string, value int, err error) {
	if constraint.Comparison != tokenComparisonEQ {
		return false, "", 0, nil
	}

	freeVars := constraint.FreeVars()
	if freeVars.Size() != 1 {
		return false, "", 0, nil
	}
	freeVar := freeVars.Pop()
	mathExpression := ""
	varExpression := ""

	if constraint.LeftTerm.FreeVars().IsEmpty() {
		mathExpression = constraint.LeftTerm.String()
		varExpression = constraint.RightTerm.String()
	} else if constraint.RightTerm.FreeVars().IsEmpty() {
		asserts(constraint.RightTerm.FreeVars().IsEmpty(), "Must be math expression: "+constraint.String())
		mathExpression = constraint.RightTerm.String()
		varExpression = constraint.LeftTerm.String()
	}

	if !strings.HasPrefix(varExpression, freeVar) {
		return false, "", 0, nil
	}

	remainingExpression := strings.TrimPrefix(varExpression, freeVar)
	asserts(Term(remainingExpression).FreeVars().IsEmpty(), "Must be math expression: "+remainingExpression)
	if remainingExpression == "" {
		val, err := evaluateTermExpression(mathExpression)
		return true, freeVar, val, err
	}

	if strings.HasPrefix(remainingExpression, "+") {
		tmp := strings.TrimPrefix(remainingExpression, "+")
		val, err := evaluateTermExpression(mathExpression + "-(" + tmp + ")")
		return true, freeVar, val, err
	}

	if strings.HasPrefix(remainingExpression, "-") {
		tmp := strings.TrimPrefix(remainingExpression, "-")
		val, err := evaluateTermExpression(mathExpression + "+(" + tmp + ")")
		return true, freeVar, val, err
	}

	return false, "", 0, nil
}

// Remove Rules with false constraint
// Remove true constraints from Rule
// This is essentially Unit Propagation on Constraint Instantiation
func (p *Program) CleanRulesFromGroundBoolExpression() (bool, error) {

	check := func(r Rule) bool {
		for _, cons := range r.Constraints {
			re, _ := cons.GroundBoolExpression()
			if re {
				return true
			}
		}
		return false
	}

	transform := func(rule Rule) (result []Rule, err error) {
		newRule := rule
		newRule.Constraints = []Constraint{}
		for _, cons := range rule.Constraints {
			isGround, boolResult := cons.GroundBoolExpression()
			if isGround {
				if boolResult {
					result = []Rule{newRule}
				} else {
					result = []Rule{}
					break
				}
			} else {
				newRule.Constraints = append(rule.Constraints, cons)
			}
		}
		return
	}
	return p.RuleExpansion(check, transform)
}

// Remove Iterator if with false constraint
// Remove true constraints from Iterator
func (p *Program) CleanIteratorFromGroundBoolExpressions() (bool, error) {

	check := func(r Rule) bool {
		for _, iter := range r.Iterators {
			for _, cons := range iter.Constraints {
				re, _ := cons.GroundBoolExpression()
				if re {
					return true
				}
			}
		}
		return false
	}

	transform := func(rule Rule) ([]Rule, error) {
		newRule := rule.Copy()
		newRule.Iterators = []Iterator{}
		for _, iter := range rule.Iterators {
			newIterator := iter.Copy()
			isGood := true
			var newConstraints []Constraint
			for _, cons := range iter.Constraints {
				isGround, boolResult := cons.GroundBoolExpression()
				if isGround {
					if !boolResult {
						isGood = false
						break
					}
				} else {
					newConstraints = append(newConstraints, cons)
				}
			}
			if isGood {
				newIterator.Constraints = newConstraints
				newRule.Iterators = append(newRule.Iterators, newIterator)
			}
		}
		return []Rule{newRule}, nil
	}
	return p.RuleExpansion(check, transform)
}

func (p *Program) ConvertHeadOnlyIteratorsToLiterals() (bool, error) {

	check := func(r Rule) bool {
		for _, iterator := range r.Iterators {
			if len(iterator.Constraints) == 0 && len(iterator.Conditionals) == 0 {
				// This should be moved to be a fact!
				return true
			}
		}
		return false
	}

	transform := func(rule Rule) (Rule, error) {
		var it int
		for it = range rule.Iterators {
			if len(rule.Iterators[it].Constraints) == 0 && len(rule.Iterators[it].Conditionals) == 0 {
				break
			}
		}
		rule.Literals = append(rule.Literals, rule.Iterators[it].Head)
		rule.Iterators = append(rule.Iterators[:it], rule.Iterators[it+1:]...)
		return rule, nil
	}
	return p.RuleTransformation(check, transform)
}

// for each Constraint X==<Value>
// Rewrite all Terms with X <- <Value>
// Only within Iterators
func (p *Program) TransformConstraintsToInstantiationIterator() (bool, error) {

	check := func(r Rule) bool {
		for _, iterator := range r.Iterators {
			for _, cons := range iterator.Constraints {
				is, _, _, _ := cons.IsInstantiation()
				if is {
					return true
				}
			}
		}
		return false
	}

	transform := func(rule Rule) (Rule, error) {
		var it int
		var i int
		var cons Constraint
		var is bool
		var variable string
		var value int
		var err error
		for it = range rule.Iterators {
			for i, cons = range rule.Iterators[it].Constraints {
				is, variable, value, err = cons.IsInstantiation()
				if err != nil {
					return rule, RuleError{rule, "Transform Constraint Problem", err}
				}
				if is {
					break
				}
			}
			if is {
				break
			}
		}
		rule.Iterators[it].Constraints = append(rule.Iterators[it].Constraints[:i],
			rule.Iterators[it].Constraints[i+1:]...)
		if !IsMarkedAsFree(variable) {
			assignment := map[string]int{variable: value}
			_, err := rule.Iterators[it].Simplify(assignment)
			if err != nil {
				return rule, err
			}
		}

		//// Transform iterator into literal if empty!
		if len(rule.Iterators[it].Constraints) == 0 && len(rule.Iterators[it].Conditionals) == 0 {
			rule.Literals = append(rule.Literals, rule.Iterators[it].Head)
			rule.Iterators = append(rule.Iterators[:it], rule.Iterators[it+1:]...)
		}
		return rule, err
	}
	return p.RuleTransformation(check, transform)
}

// for each Constraint X==<Value>
// Rewrite all Terms with X <- <Value>
func (p *Program) TransformConstraintsToInstantiation() (bool, error) {

	check := func(r Rule) bool {
		for _, cons := range r.Constraints {
			is, _, _, _ := cons.IsInstantiation()
			if is {
				return true
			}
		}
		return false
	}

	transform := func(rule Rule) (Rule, error) {
		var i int
		var cons Constraint
		var is bool
		var variable string
		var value int
		var err error
		for i, cons = range rule.Constraints {
			is, variable, value, err = cons.IsInstantiation()
			if err != nil {
				return rule, RuleError{rule, "Transform Constraint Problem", err}
			}
			if is {
				break
			}
		}
		rule.Constraints = append(rule.Constraints[:i], rule.Constraints[i+1:]...)
		if !IsMarkedAsFree(variable) {
			assignment := map[string]int{variable: value}
			_, err := rule.Simplify(assignment)
			if err != nil {
				return rule, err
			}
		}
		return rule, err
	}
	return p.RuleTransformation(check, transform)
}

func (p *Program) ReplaceConstantsAndMathFunctions() {

	transform := func(term Term) (Term, bool, error) {
		out := strings.ReplaceAll(string(term), "#mod", "%")
		return Term(out), out != string(term), nil
	}

	for i := range p.Rules {
		TermTranslation(p.Rules[i], transform)
		p.Rules[i].Simplify(p.Constants)
	}
}

func (iterator *Iterator) Simplify(assignment map[string]int) (bool, error) {

	transform := func(term Term) (Term, bool, error) {
		return assign(term, assignment)
	}

	return TermTranslation(iterator, transform)
}

func (rule *Rule) Simplify(assignment map[string]int) (bool, error) {

	transform := func(term Term) (Term, bool, error) {
		return assign(term, assignment)
	}

	return TermTranslation(rule, transform)
}

// Check if rule contains #exist or #forall literal!
// At this version of BULE it has to have exactly 2 literals, be ground, and have question mark!
// Then take literal and add tuple and add to quantification level

func (p *Program) CollectExplicitTupleDefinitions() (bool, error) {
	p.forallQ = make(map[int][]Literal, 0)
	p.existQ = make(map[int][]Literal, 0)
	p.Explicit = make(map[Predicate]bool, 0)

	check := func(r Rule) bool {
		if r.IsQuestionMark == true {
			return true
		}
		//for _, l := range r.Literals {
		//	if l.Name == "#exist" || l.Name == "#forall" {
		//		return true
		//	}
		//}
		return false
	}

	transform := func(rule Rule) ([]Rule, error) {
		//if !rule.IsQuestionMark {
		//	return []Rule{}, RuleError{
		//		R:       rule,
		//		Message: "Must have questions mark!",
		//		Err:     nil,
		//	}
		//}
		if !rule.IsGround() {
			return []Rule{}, RuleError{
				R:       rule,
				Message: "Must be ground when explicit collection takes place!",
				Err:     nil,
			}
		}
		if len(rule.Literals) != 2 &&
			(rule.Literals[0].Name == "#exist" || rule.Literals[0].Name == "#forall") {
			return []Rule{}, RuleError{
				R:       rule,
				Message: "The clause must have exactly 2 literals (first is quantifier of exist or forall, second literal)!",
				Err:     nil,
			}
		}
		err := p.InsertLiteralTuple(rule.Literals[1])
		if err != nil {
			err = LiteralError{
				L:       rule.Literals[1],
				R:       rule,
				Message: fmt.Sprintf("Could not insert tuple into db. %v", err),
			}
		}

		p.Explicit[rule.Literals[1].Name] = true
		quantification := rule.Literals[0]
		if len(quantification.Terms) != 1 {
			return []Rule{}, LiteralError{
				L:       quantification,
				R:       rule,
				Message: fmt.Sprintf("Wrong arity %v, should be 1",len(quantification.Terms) ),
			}
		}
		val, err := evaluateTermExpression(quantification.Terms[0].String())
		if err != nil {
			return []Rule{}, LiteralError{
				L:       quantification,
				R:       rule,
				Message: fmt.Sprintf("Cant evaluate, not ground", quantification.Terms[0]),
			}
		}
		switch quantification.Name {
		case "#forall":
			p.forallQ[val] = append(p.forallQ[val], rule.Literals[1])
		case "#exist":
			p.existQ[val] = append(p.existQ[val], rule.Literals[1])
		default:
			err = fmt.Errorf("First literal in clause must be #forall or #exist. %v", rule)
		}

		return []Rule{}, err
	}

	return p.RuleExpansion(check, transform)
}

func (p *Program) CollectGroundTuples() (bool,error) {

	for _, r := range p.Rules {
		for _, literal := range r.Literals {
			if literal.IsGround() {
				err := p.InsertLiteralTuple(literal)
				if err != nil {
					return true,LiteralError{
						L:       literal,
						R:       r,
						Message: fmt.Sprintf("%v", err),
					}
				}
			}
		}
	}
	return true,nil
}

func (p *Program) RemoveRulesWithGenerators() (bool, error) {
	removeIfTrue := func(rule Rule) bool {
		if len(rule.Generators) > 0 {
			return true
		}
		return false
	}
	return p.RemoveRules(removeIfTrue)
}

func (p *Program) RemoveLiteralsWithEmptyIterators() (bool, error) {
	removeIfTrue := func(rule Rule) bool {
		if len(rule.Iterators) > 0 {
			return true
		}
		return false
	}
	return p.RemoveRules(removeIfTrue)
}

func (p *Program) RemoveClausesWithExplicitLiteralAndTuplesThatDontExist() (bool, error) {
	removeIfTrue := func(rule Rule) bool {
		for _, lit := range rule.Literals {
			if p.Explicit[lit.Name] && lit.FreeVars().IsEmpty() {
				if !p.PredicateTupleMap[lit.String()] {
					return true
				}
			}
		}
		return false
	}
	return p.RemoveRules(removeIfTrue)
}

func (p *Program) RemoveClausesWithTuplesThatDontExist() (bool, error) {
	removeIfTrue := func(rule Rule) bool {
		for _, lit := range rule.Literals {
			if lit.FreeVars().IsEmpty() {
				if !p.PredicateTupleMap[lit.String()] {
					return true
				}
			}
		}
		return false
	}
	return p.RemoveRules(removeIfTrue)
}

func (p *Program) ExtractQuantors() {

	p.forallQ = make(map[int][]Literal, 0)
	p.existQ = make(map[int][]Literal, 0)

	checkA := func(r Rule) bool {

		return len(r.Literals) > 0 && r.Literals[0].Name == "#forall"
	}

	transformA := func(rule Rule) (remove []Rule, err error) {
		lit := rule.Literals[0]
		asserts(len(lit.Terms) == 1, "Wrong arity for forall")
		val, err := strconv.Atoi(string(lit.Terms[0].String()))
		asserte(err)
		p.forallQ[val] = append(p.forallQ[val], rule.Literals[1:]...)
		return
	}

	checkE := func(r Rule) bool {
		return len(r.Literals) > 0 && r.Literals[0].Name == "#exist"
	}

	transformE := func(rule Rule) (remove []Rule, err error) {
		lit := rule.Literals[0]
		asserts(len(lit.Terms) == 1, "Wrong arity for exist")
		val, err := strconv.Atoi(string(lit.Terms[0].String()))
		asserte(err)
		p.existQ[val] = append(p.existQ[val], rule.Literals[1:]...)
		return
	}

	p.RuleExpansion(checkA, transformA)
	p.RuleExpansion(checkE, transformE)
	return
}

// only works on disjunctions
func (rule *Rule) FreeVars() *strset.Set {
	set := strset.New()
	for _, a := range rule.Literals {
		set.Merge(a.FreeVars())
	}
	return set
}

func (literal Literal) FreeVars() *strset.Set {
	set := strset.New()
	for _, t := range literal.Terms {
		set.Merge(t.FreeVars())
	}
	return set
}

func (constraint Constraint) FreeVars() *strset.Set {
	set := constraint.LeftTerm.FreeVars()
	set.Merge(constraint.RightTerm.FreeVars())
	return set
}

func (term Term) FreeVars() *strset.Set {
	s := term.String()
	f := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c) && c != '_'
	}
	variables := strings.FieldsFunc(s, f)
	set := strset.New()
	for _, x := range variables {
		if !number(x) {
			set.Add(x)
		}
	}
	return set
}

func (constraint *Constraint) GroundBoolExpression() (isGround bool, result bool) {
	isGround = groundMathExpression(string(constraint.LeftTerm)) && groundMathExpression(string(constraint.RightTerm))
	if !isGround {
		return
	}
	result = evaluateBoolExpression(constraint.String())
	return
}

// Makes a deep copy and creates a new Literal
func (literal Literal) assign(assignment map[string]int) (newLiteral Literal, err error) {
	newLiteral = literal.Copy()
	for i, term := range literal.Terms {
		newLiteral.Terms[i], _, err = assign(term, assignment)
	}
	return newLiteral, nil
}

//returns true if term has been changed
func assign(term Term, assignment map[string]int) (Term, bool, error) {
	output := term.String()
	for Const, Val := range assignment {
		// TODO currenlty variables need to be prefix free, i.e. X, Xa, will make problems :\
		// Use Term parser for getting proper FreeVariables. Or the FreeVariables function
		output = strings.ReplaceAll(output, Const, strconv.Itoa(Val))
	}
	if groundMathExpression(output) {
		val, err := evaluateTermExpression(output)
		if err != nil {
			return Term(output), false, err
		}
		output = strconv.Itoa(val)
	}
	return Term(output), term.String() != output, nil
}

func number(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func groundMathExpression(s string) bool {
	//	r, _ := regexp.MatchString("[0-9+*/%]+", s)
	return "" == strings.Trim(string(s), "0123456789+*%-/()")
}

// Evaluates a ground math expression, needs to path mathExpression
func evaluateBoolExpression(termComparison string) bool {
	//	termComparison = strings.ReplaceAll(termComparison, "#mod", "%")
	expression, err := govaluate.NewEvaluableExpression(termComparison)
	assertx(err, termComparison)
	result, err := expression.Evaluate(nil)
	assertx(err, termComparison)
	return result.(bool)
}

// Evaluates a ground math expression, needs to path mathExpression
func evaluateTermExpression(termExpression string) (int, error) {
	//	termExpression = strings.ReplaceAll(termExpression, "#mod", "%")
	expression, err := govaluate.NewEvaluableExpression(termExpression)
	if err != nil {
		return 0, fmt.Errorf("problem in term expression %v: %w", termExpression, err)
	}
	result, err := expression.Evaluate(nil)
	if err != nil {
		return 0, fmt.Errorf("problem in term expression %v: %w", termExpression, err)
	}
	return int(result.(float64)), nil
}

// Evaluates a ground math expression, needs to pass mathExpression
func evaluateExpressionTuples(terms []Term) (result []int, err error) {
	for _, t := range terms {
		val, err := evaluateTermExpression(string(t))
		if err != nil {
			return result, err
		}
		result = append(result, val)
	}
	return
}

func (literal *Literal) findGroundTerms() (positions []int, values []int) {
	for i, t := range literal.Terms {
		if v, ok := strconv.Atoi(t.String()); ok == nil {
			positions = append(positions, i)
			values = append(values, v)
		}
	}
	return
}

func (literal *Literal) createNegatedLiteral() Literal {
	a := literal.Copy()
	a.Neg = !a.Neg
	return a
}

func (constraint *Constraint) createNegatedConstraint() Constraint {
	negatedConstraint := constraint.Copy()
	switch constraint.Comparison {
	case tokenComparisonLT:
		constraint.Comparison = tokenComparisonGE
	case tokenComparisonGT:
		constraint.Comparison = tokenComparisonLE
	case tokenComparisonEQ:
		constraint.Comparison = tokenComparisonNQ
	case tokenComparisonGE:
		constraint.Comparison = tokenComparisonLT
	case tokenComparisonLE:
		constraint.Comparison = tokenComparisonGT
	case tokenComparisonNQ:
		constraint.Comparison = tokenComparisonEQ
	}
	return negatedConstraint
}

func ComparisonString(tokenComparison tokenKind) (s string) {
	switch tokenComparison {
	case tokenComparisonLT:
		s = "<"
	case tokenComparisonGT:
		s = ">"
	case tokenComparisonEQ:
		s = "=="
	case tokenComparisonGE:
		s = ">="
	case tokenComparisonLE:
		s = "<="
	case tokenComparisonNQ:
		s = "!="
	}
	return
}

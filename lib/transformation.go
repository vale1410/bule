package lib

import (
	"fmt"
	"github.com/scylladb/go-set/strset"
)

// Check that there are no unbound variables
// All variables that only occur in exactly one term are not bound by others and
// must  be marked as free and unbound (i.e. starting with underscore _)
func (p *Program) CheckUnboundVariables() error {
	for _, rule := range p.Rules {
		countVars := make(map[string]int, 0)
		for _, term := range rule.AllTerms() {
			for _, v := range term.FreeVars().List() {
				if !IsMarkedAsFree(v) {
					if c, ok := countVars[v]; ok {
						countVars[v] = c + 1
					} else {
						countVars[v] = 1
					}
				}
			}
		}

		// collect free vars in head of iterators,
		// they can occur alone but due to unrolling the generator
		// they do occur multiple times. Example. q(X,D):d[D]. % X occurs alone but is rolled out
		inHead := strset.New()
		for _, g := range rule.Iterators {
			for i := range g.Head.Terms {
				inHead = strset.Union(inHead, g.Head.Terms[i].FreeVars())
			}
		}

		for v, c := range countVars {
			if !inHead.Has(v) && c < 2 {
				return RuleError{
					rule,
					fmt.Sprintf("The variables %s is not marked as free (starting with underscore _).", v),
					nil,
				}
			}
		}
	}
	return nil
}

// Remove all rules where check is true.
func (p *Program) RemoveRules(ifTrueRemove func(r Rule) bool) (changed bool) {
	var newRules []Rule
	for _, rule := range p.Rules {
		if !ifTrueRemove(rule) {
			newRules = append(newRules, rule)
		}
	}
	p.Rules = newRules
	return
}

// goes through all rules and expands expands if check is true.
// Note that this does not expand the generated rules. (i.e. it does not run until fixpoint)
func (p *Program) RuleExpansion(check func(r Rule) bool, expand func(Rule) ([]Rule, error)) (changed bool, err error) {
	var newRules []Rule
	for _, rule := range p.Rules {
		if check(rule) {
			changed = true
			tmpRules, err := expand(rule)
			if err != nil {
				return changed, RuleError{
					R:       rule,
					Message: fmt.Sprintf("Rule Expansion: %v", err),
				}
			}
			newRules = append(newRules, tmpRules...)
		} else {
			newRules = append(newRules, rule)
		}
	}
	p.Rules = newRules
	return
}

// goes through all rules and translates if check is true.
// Singleton version of RuleExpansion
func (p *Program) RuleTransformation(check func(r Rule) bool,
	transformation func(Rule) (Rule, error)) (changed bool, err error) {
	return p.RuleExpansion(check, func(r Rule) ([]Rule, error) {
		rn, err := transformation(r)
		return []Rule{rn}, err
	})
	//for row, rule := range p.Rules {
	//	if check(rule) {
	//		changed = true
	//		p.Rules[row], err = transformation(rule)
	//		if err != nil {
	//			return changed, RuleError{
	//				R:       rule,
	//				Message: fmt.Sprintf("Rule Transformation %v",err),
	//			}
	//		}
	//	}
	//}
	return
}

// Check terms in literals and expands the rules according to the *first* term found according to it's expansion.
func (p *Program) TermExpansion(check func(r Term) bool, expand func(Term) ([]Term,error)) (changed bool, err error) {

	checkRule := func(r Rule) bool {
		for _, t := range r.AllTerms() {
			if check(*t) {
				return true
			}
		}
		return false
	}

	// Rule is completely replaced!
	expandRule := func(r Rule) (newRules []Rule, err error) {
		workingRule := r.Copy()
		workingRule.Parent = &r
		for _, t := range workingRule.AllTerms() {
			if check(*t) {
				terms, err := expand(*t)
				if err != nil {
					return newRules, err
				}
				for _, newTerm := range terms {
					*t = newTerm
					newRule := workingRule.Copy()
					newRules = append(newRules, newRule)
				}
			}
		}
		return
	}
	return p.RuleExpansion(checkRule, expandRule)
}

func (rule *Rule) TermTranslation(transform func(Term) (Term, bool, error)) (changed bool, err error) {
	var ok bool
	for _, term := range rule.AllTerms() {
		*term, ok, err = transform(*term)
		changed = ok || changed
		if err != nil {
			return changed, err
		}
	}
	return
}

func (rule *Rule) AllTerms() (terms []*Term) {
	for _, l := range rule.AllLiterals() {
		for i := range l.Terms {
			terms = append(terms, &l.Terms[i])
		}
	}
	for i := range rule.Constraints {
		terms = append(terms, &rule.Constraints[i].LeftTerm)
		terms = append(terms, &rule.Constraints[i].RightTerm)
	}
	for _, g := range rule.Iterators {
		for j := range g.Constraints {
			terms = append(terms, &g.Constraints[j].LeftTerm)
			terms = append(terms, &g.Constraints[j].RightTerm)
		}
	}
	return
}

func (rule *Rule) AllLiterals() (lits []*Literal) {

	for i := range rule.Generators {
		lits = append(lits, &rule.Generators[i])
	}

	for i := range rule.Literals {
		lits = append(lits, &rule.Literals[i])
	}

	for i := range rule.Iterators {
		lits = append(lits, &rule.Iterators[i].Head)
		for j := range rule.Iterators[i].Literals {
			lits = append(lits, &rule.Iterators[i].Literals[j])
		}
	}
	return
}

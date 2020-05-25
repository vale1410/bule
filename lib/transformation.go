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

		// collect free vars in head of generators,
		// they can occur alone but due to unrolling the generator
		// they do occur multiple times. Example. q(X,D):d[D]. % X occurs alone but is rolled out
		inHead := strset.New()
		for _, g := range rule.Generators {
			for i := range g.Head.Terms {
				inHead = strset.Union(inHead, g.Head.Terms[i].FreeVars())
			}
		}

		for v, c := range countVars {
			if !inHead.Has(v) && c < 2 {
				return fmt.Errorf("In the following rule the variables %s is not marked as free and unbound (starting with underscore _).\n %s", v, rule.String())
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
	for row, rule := range p.Rules {
		if check(rule) {
			changed = true
			tmpRules, err := expand(rule)
			if err != nil {
				return false, fmt.Errorf("Rele Expansion: %w\n in Rule %v:  %v ", err, row, rule)
			}
			newRules = append(newRules, tmpRules...)
		} else {
			newRules = append(newRules, rule)
		}
	}
	p.Rules = newRules
	return
}

// Check terms in literals and expands the rules according to the *first* term found according to it's expansion.
func (p *Program) TermExpansionOnlyLiterals(check func(r Term) bool, expand func(Term) []Term) (changed bool, err error) {

	checkRule := func(r Rule) bool {
		for _, l := range r.Literals {
			for _, t := range l.Terms {
				if check(t) {
					return true
				}
			}
		}
		return false
	}

	expandRule := func(r Rule) (newRules []Rule, err error) {
		for il, literal := range r.Literals {
			for it, term := range literal.Terms {
				if check(term) {
					for _, newTerm := range expand(term) {
						newRule := r.Copy()
						newRule.Literals[il].Terms[it] = newTerm
						newRules = append(newRules, newRule)
					}
					return
				}
			}
		}
		return
	}
	return p.RuleExpansion(checkRule, expandRule)
}

func (r *Rule) TermTranslation(transform func(Term) (Term, bool)) (changed bool) {
	var ok bool
	for _, term := range r.AllTerms() {
		*term, ok = transform(*term)
		changed = ok || changed
	}
	return
}

func (r *Rule) AllTerms() (terms []*Term) {
	for i := range r.Head.Terms {
		terms = append(terms, &r.Head.Terms[i])
	}
	for _, l := range r.Literals {
		for i := range l.Terms {
			terms = append(terms, &l.Terms[i])
		}
	}
	for i := range r.Constraints {
		terms = append(terms, &r.Constraints[i].LeftTerm)
		terms = append(terms, &r.Constraints[i].RightTerm)
	}
	for _, g := range r.Generators {
		for i := range g.Head.Terms {
			terms = append(terms, &g.Head.Terms[i])
		}
		for _, l := range g.Literals {
			for i := range l.Terms {
				terms = append(terms, &l.Terms[i])
			}
		}
		for j := range g.Constraints {
			terms = append(terms, &g.Constraints[j].LeftTerm)
			terms = append(terms, &g.Constraints[j].RightTerm)
		}
	}
	return
}

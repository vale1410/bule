package lib


// goes through all rules and expands expands if check is true.
// Note that this does not expand the generated rules. (i.e. run until fixpoint)
func (p *Program) RuleExpansion(check func(r Rule) bool, expand func(Rule) []Rule) (changed bool) {
	var newRules []Rule
	for _, rule := range p.Rules {
		if check(rule) {
			changed = true
			for _, newRule := range expand(rule) {
				newRules = append(newRules, newRule)
			}
		} else {
			newRules = append(newRules, rule)
		}
	}
	p.Rules = newRules
	return
}

func (p *Program) TermExpansionOnlyLiterals(check func(r Term) bool, expand func(Term) []Term) (changed bool) {

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

	expandRule := func(r Rule) (newRules []Rule) {
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
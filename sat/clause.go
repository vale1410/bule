package sat

type Clause struct {
	Desc     string
	Literals []Literal
}

type ClauseSet []Clause

func (cs *ClauseSet) AddClause(literals ...Literal) {
	*cs = append(*cs, Clause{"-", literals})
}

func (cs *ClauseSet) AddTaggedClause(tag string, literals ...Literal) {
	*cs = append(*cs, Clause{tag, literals})
}

func (cs *ClauseSet) AddClauseSet(cl ClauseSet) {
	*cs = append(*cs, cl...)
}


package sat

import (
	"fmt"
	"os"
	"strconv"
)

type Clause struct {
	Desc     string
	Literals []Literal
}

type ClauseSet []Clause

func (cs *ClauseSet) AddClause(desc string, literals ...Literal) {
	*cs = append(*cs, Clause{desc, literals})
}

func (cs *ClauseSet) AddClauseSet(cl ClauseSet) {
	*cs = append(*cs, cl...)


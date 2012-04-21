Bule: A State of the Art CNF Grounder
=====================================

A refreshing new CNF grounder - combining features of Gringo and Sugar. The input will be some kind of logic program as
in Gringo - with support for pure clauses, implications, rules with completion semantics, cardinality and weight
constraints. The generated output will pure plain DIMACS CNF. This is a long term project and I am just at the very
beginning.  The grounder will be entirely programmed in GO!

Milestones
----------

1. Grounding as in Gringo with simple facts, only clauses
2. Rules, implications and equivalences
3. Aggregates as in Gringo
4. Cardinality constraints
5. Weight constraints with a selection of different translations
6. Optimization statements - calling solvers parallel with branching on the optimization function


Links
-----
* http://potassco.sourceforge.net/
* http://bach.istc.kobe-u.ac.jp/sugar/
* http://golang.org/

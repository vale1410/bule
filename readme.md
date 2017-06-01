Bule 1.0: A simple CNF Generator (depricated)
=====================================

Bule is a framework that helps to create good CNF (conjunctive normal form) encodings.

At the current stage it supports the following features

* Special Treatment of atmost1 and ex1, and other minor special cases
* Cardinality constraints through sorting networks
* Hybrid translation of Pseudo Boolean Constraints through sorting networks and MDDs

We plan to extend the framework to other global constraints and incorporate
a incremental grounding technique, and several enumeration techniques for
optimization problems. 

Furthermore, it helps debugging the CNF by a human readable output of the clauses and statistics. 

Features
----------

* Cardinality and Weight Constraints (Pseudo Booleans) through sorters 
* Grounding to CNF and statistics 
* Bule solves PB decision problems 
* MDD based translation of PBs
* Combinators of translation (PB + AMO/EX1)
* Incremental/enumeration of optimization statements 
* Calling SAT solvers from bule

Milestones
----------

* Consolidate statistic of SAT solvers
* Portfolio solving in parallel with multiple SAT solvers
* Counter Based encoding for Cardinality
* DSL Rule language for GOLANG independent use of the framework (similar to gringo maybe)
* Support for other constraints (sequence, regular, alldiff etc.)

Links
-----
* http://minisat.se/MiniSat+.html
* http://golang.org/
* http://potassco.sourceforge.net/
* http://bach.istc.kobe-u.ac.jp/sugar/
* http://www.cs.utexas.edu/users/vl/tag/SAT-grounders


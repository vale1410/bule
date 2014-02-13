Bule: A simple CNF Generator
=====================================

Bule is a framework that helps to generate good CNF (conjunctive normal form) encodings.

At the current stage it supports these basic constraints: 

* Cardinality Constraints through sorting networks
* Pseudo Boolean Constraints through sorting networks by a novel encoding

Furthermore, it helps debugging the CNF by a human readable output of the clauses and statistics. 

We plan to extend the framework to other global constraints and incorporate
a incremental grounding technique, and several enumeration techniques for
optimization problems. 


Milestones
----------

* Cardinality and Weight Constraints (Pseudo Booleans) [done]
* Grounding to CNF and statistics [done]
* Incremental/enumeration of optimization statements 
* Calling SAT solvers through the framework, consolidate their statistics
* DSL Rule language for GOLANG independent use of the framework (similar to gringo maybe)
* Support for other constraints (sequence, regular etc.)

Links
-----
* http://minisat.se/MiniSat+.html
* http://golang.org/
* http://potassco.sourceforge.net/
* http://bach.istc.kobe-u.ac.jp/sugar/
* http://www.cs.utexas.edu/users/vl/tag/SAT-grounders


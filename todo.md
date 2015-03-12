today
=====
* produce sat package with interface to sat solver:
* call encode, solve, etc.  all on fundamental solver
* print variables and auxiliary variables (might help to find the bug!)
* add solver interface for SAT(cryptominisat?, clasp etc. does not matter)

this week
=========
* finish the test suite, a mix of problems
* categorize benchmarks into (type of constraints, possible matchings, rankings)
* prepare the matching problem of real PBs to constraints 
* structure paper, touch and update, print for Monday
* do solving and parsing!
* finish the categorization of the constraint

long term
==========
* treat objective function by iterative calls to SAT solver (different solver!!)
* integrate AMO into BDD/SN, compactify (split this task into smaller ones)
* treat equality for pseudo booleans with BDDs and SN
* translate MDDs instead of BDDs, introduce IntVar

open tasks (unspecified due date)
==========
* Improve SN by smart cardinality netowrks (ignasis idea)
* Improve BDDs by pruning units before translating to SAT

Milestones
==========
* Table for Pseudo Boolean competition

ideas
=====
* Prune units before the SAT translation
* Often PB problems have several PBs with the same “structure”. 
    For the future,  memoize encodings and “structure” of PB for reuse (ie.e. 2x1+2x2+x3+x4+x5<=5) many times. 

done
====
* treat equality for Cardinality Constraints
* produce 3 examples for rewriting
* get bug (number of results on current test set incorrect)
* write three more test for categorize cardinality
* design the integration of matchings

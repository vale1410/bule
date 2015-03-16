today(monday)
=====
* Simple Match PB with AMO with auxiliary variables(counter encoding)  
       and apply it on BDD. 
* Prepare the matching problem of real PBs to constraints. 

this week
=========
* Restructure paper. 
* Apply it at the same time to SN
* Categorize benchmarks into (type of constraints, possible matchings, rankings).

long term
==========
* treat objective function by iterative calls to SAT solver (different solver!!)
* integrate AMO into BDD/SN, compactify (split this task into smaller ones)
* treat equality for pseudo booleans with BDDs and SN
* translate MDDs instead of BDDs, introduce IntVar
* add solver interface for SAT with several solvers and statistics (cryptominisat, lingeling, ), call them in parallel ...

open tasks (unspecified due date)
==========
* Improve SN by smart cardinality netowrks (ignasis idea)
* Improve BDDs by pruning units before translating to SAT

Milestones
==========
* Table for Pseudo Boolean competition
* Table for BP of MIP
* First Draft of Paper

ideas
=====
* Prune units before the SAT translation
* Often PB problems have several PBs with the same “structure”. 
    For the future,  memoize encodings and “structure” of PB for reuse (ie.e. 2x1+2x2+x3+x4+x5<=5) many times. 

done
====
* Finish the test suite, a mix of problems.
* print variables and auxiliary variables
* solver interface with clasp - parse variables, print assignment
* treat equality for Cardinality Constraints
* produce 3 examples for rewriting
* get bug (number of results on current test set incorrect)
* write three more test for categorize cardinality
* design the integration of matchings

Today
=====
* Do Table for categorization, group benchmarks into interesting and not interesting!
* Prepare the matching problem of real PBs to constraints 

This week
=========
* Do BDD with AMO
* Do SN with AMO
* Restructure paper. 
* Integrate convert.go into bule.go; call clasp with the program!
* Simple Match PB with AMO with auxiliary variables(counter encoding)  
       and apply it on BDD. 
* Prepare the matching problem of real PBs to constraints. 

Long term
==========
* Treat objective function by iterative calls to SAT solver (different solver!!)
* Treat equality for pseudo booleans with BDDs and SN
* Introduce IntVar
* Add solver interface for SAT with several solvers and statistics (cryptominisat, lingeling, ), call them in parallel ...

Open tasks (unspecified due date)
==========
* Improve SN by smart cardinality networks (ignasis idea)
* Improve BDDs by pruning units before translating to SAT

Milestones
==========
* Integrate AMO into BDD/SN, compactify (split this task into smaller ones)
* Table for Pseudo Boolean competition
* Table for BP of MIP
* First Draft of Paper sent to Toby
* Treat Optimization Problems naively

Ideas
=====
* Prune units before the SAT translation
* Often PB problems have several PBs with the same “structure”. 
    For the future,  memoize encodings and “structure” of PB for reuse (ie.e. 2x1+2x2+x3+x4+x5<=5) many times. 

Done
====
* Do Rewrite EX1 and PB
* Structure paper, touch and update, print for Monday
* Finish the test suite, a mix of problems (did some stuff)
* Print variables and auxiliary variables
* Solver interface with clasp - parse variables, print assignment
* Treat equality for Cardinality Constraints
* Produce 3 examples for rewriting
* Get bug (number of results on current test set incorrect)
* Write three more test for categorize cardinality
* Design the integration of matchings
* Finish the test suite, a mix of problems.

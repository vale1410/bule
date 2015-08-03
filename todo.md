Today
=====

This week
=========

Open tasks (unspecified due date)
=================================
* Improve SN by smart cardinality networks (ignasis idea)
* Improve BDDs by pruning units before translating to SAT
* Treat equality for pseudo booleans with BDDs and SN
* Introduce IntVar for modelling with counter encoding
* Integrate convert.go into bule.go; call clasp with the program!
* Use incremental interface of SAT solvers (minisat and cmsat?)

Milestones
==========
* Integrate AMO into BDD/SN
* Support Ignasi type Cardinality Networks

Ideas
=====
* Meta translations of a networks and nodes. by super function of types of
  nodes and recursive translation. This could subsume many types of
  translations, for instance, sorting networks, circuits etc. 
* Prune unit clauses before the SAT translation
* Often PB problems have several PBs with the same “structure”. 
    For the future,  memoize encodings and “structure” of PB for reuse (ie.e. 2x1+2x2+x3+x4+x5<=5) many times. 

Done
====

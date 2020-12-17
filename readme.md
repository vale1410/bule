# ![Bule Logo](logo.png =50) The SAT Programming Language **Bule**

Bule is a tool to create beautiful CNF encodings.
Bule is a sophisticated grounder for the modelling language Bule that translates to CNF for SAT Solving. 
Bule provides a front end for various SAT technologies. 

## Features

* Grounding with the declarative modelling language Bule
* satisfiability solving - allowing any number of SAT solvers to be called with the grounded CNF formula
* debugging facilities for CNF formulas, statistics on size and quality 
* QBF solving
* Model Counting and Approximate model counting
* Various encodings for cardinality constraints and Pseudo Boolean constaints. 
* Multiple cardinality encodings
* Full Pseudo Boolean Translations to CNF

## Basic Configuration For SAT and QBF

Now you can now add SAT and QBF solvers to the configuration of bule and solve your formulas with them!
As a start, for QBF add depqbf and caqe and for SAT kissat and cryptominisat to your path. 
The installation instructions can be found here:  
* http://lonsing.github.io/depqbf/
* https://github.com/ltentrup/caqe
* https://github.com/msoos/cryptominisat/
* http://fmv.jku.at/kissat/


Now let's add the solvers to the configuration file (usually at `~/.bule.yaml`): 

```
>>> bule add depqbf @"--no-dynamic-nenofex --qdo" QBF -l default
>>> bule add cryptominisat @ SAT -l default
```

with `bule list` you can list all configurations: 

```
>>> bule list 
```

To test the configuration take a look at the small examples in bule/examples/:
* `qbf_false.bul`
* `sat_false.bul`
* `qbf_true.bul`
* `sat_true.bul`

with `solve` bule should give you the expected answers. 

For example: 

```
>>> bule solve examples/sat_true.bul
C program grounded in 11.870796ms.
This is a SAT problem
Using a SAT solver instance 'cmsat '
Solving. . .

-------- Solver output -----
2020/12/17 02:31:15 solver>>  c Outputting solution to console
2020/12/17 02:31:15 solver>>  c CryptoMiniSat version 5.7.1
...
...
2020/12/17 02:31:15 solver>>  c Mem used                 : 0.00        MB
2020/12/17 02:31:15 solver>>  c Total time (this thread) : 0.00
2020/12/17 02:31:15 solver>>  s SATISFIABLE
-------- Solver output end -----
s SAT
--------
~q.
p.
--------
```

## Bule's syntax and simple programs

### Literals and basic clauses

Let us have a 0-arity literal `q`.\
Also, let's have a 1-clause rule of form:

```prolog
q.
```

We can observe that this rule is easily satisfiable when `q <=> True`.

```prolog
>>> bule solve prog.bul
SAT
```

---

Let us have another 0-arity literal `p`.\
Also, let's have a 2-clause rule of form:

```prolog
q.
p.
```

That effectively translates to `p AND q`\
We can observe that this rule is satisfiable when both literals are True.
```prolog
SAT
```

---

Adding a negation of one of the literals to our rule breaks satisfiability.

```prolog
q.
p.
~q.
```

Because `q AND p AND (NOT q)` <=> `(q AND (NOT q)) AND p` <=> `False AND p` <=> `False`.
```prolog
UNSAT
```

---

### Ranges and generators

Say we want to define a domain `dom` on set `{1,2,3}`.\
We can achieve this with range expression (both brackets are inclusive):

```prolog
dom[1..3].
```

Will translate to:

```prolog
dom[1].
dom[2].
dom[3].
```

Let us have a 1-arity literal `p(X)`
Then, we can generate a set of clauses of form `p(X)` with variable `X` bound to `dom`:

```prolog
dom[X] :: p(X).
```

Which translates to:

```prolog
p(1).
p(2).
p(3).
```

Let us have  another 1-arity literal `q(Y)`
We can then iterate over Y within a single clause to add more literals:

```prolog
dom[X] :: p(X), ~q(Y*Y) : dom[Y] : Y < 3.
```

Gives:

```prolog
p(1), ~q(1), ~q(4).
p(2), ~q(1), ~q(4).
p(3), ~q(1), ~q(4).
```

Note that adding the rule `Y < 3` skips last iteration step `~q(9)` as `3 < 3` <=> `False`.

---

### Modelling Sudoku game in Bule


Let's have a domain for single row / single column indexing

```prolog
dom[1..9].
```

Similarly, let's define a 2D domain for our `X, Y` coordinates:

```prolog
domCoords[1..9,1..9].
```

Also, let's define a 2D domain for inner-box starting coords:

```prolog
boxBegin[1,1].
boxBegin[1,4].
boxBegin[1,7].
boxBegin[4,1].
boxBegin[4,4].
boxBegin[4,7].
boxBegin[7,1].
boxBegin[7,4].
boxBegin[7,7].
```

Lastly, let's define a 2D domain for coordinates offset within a box:

```prolog
boxOffset[0..2,0..2].
```

Now, we can start applying Sudoku rules in Bule!

---

**Rule 1**: in each cell on board at least 1 value from range `1..9`

```prolog
domCoords[X,Y] :: q(X,Y,Z) : dom[Z].
```

Which will generate a grand total of 81 clauses of length 9 grouped by `X, Y`:

```prolog
q(1,1,1), q(1,1,2), q(1,1,3), q(1,1,4), q(1,1,5), q(1,1,6), q(1,1,7), q(1,1,8), q(1,1,9).
q(1,2,1), q(1,2,2), q(1,2,3), q(1,2,4), q(1,2,5), q(1,2,6), q(1,2,7), q(1,2,8), q(1,2,9).
q(1,3,1), q(1,3,2), q(1,3,3), q(1,3,4), q(1,3,5), q(1,3,6), q(1,3,7), q(1,3,8), q(1,3,9).
..
..
q(1,9,1), q(1,9,2), q(1,9,3), q(1,9,4), q(1,9,5), q(1,9,6), q(1,9,7), q(1,9,8), q(1,9,9).
q(2,2,1), q(2,2,2), q(2,2,3), q(2,2,4), q(2,2,5), q(2,2,6), q(2,2,7), q(2,2,8), q(2,2,9).
..
..
q(2,9,1), q(2,9,2), q(2,9,3), q(2,9,4), q(2,9,5), q(2,9,6), q(2,9,7), q(2,9,8), q(2,9,9).
..
..
q(9,8,1), q(9,8,2), q(9,8,3), q(9,8,4), q(9,8,5), q(9,8,6), q(9,8,7), q(9,8,8), q(9,8,9).
q(9,9,1), q(9,9,2), q(9,9,3), q(9,9,4), q(9,9,5), q(9,9,6), q(9,9,7), q(9,9,8), q(9,9,9).
```

--- 

**Rule 2**: each value from range `1..9` in at least 1 cell on board

```prolog
dom[Z] :: q(X,Y,Z) : domCoords[X,Y].
```

Which will generate a grand total of 9 clauses of length 81 grouped by `Z`:

```prolog
q(1,1,1), q(1,2,1), .., q(1,9,1), q(2,1,1), q(2,2,1), .., q(2,9,1), .. .., q(9,9,1).
q(1,1,2), q(1,2,2), .., q(1,9,2), q(2,1,2), q(2,2,2), .., q(2,9,2), .. .., q(9,9,2).
..
..
q(1,1,9), q(1,2,9), .., q(1,9,9), q(2,1,9), q(2,2,9), .., q(2,9,9), .. .., q(9,9,9).
```

---

**Rule 3**: no two same values in a column

```prolog
dom[X1], dom[X2], X1 < X2 :: ~q(X1,Y,Z), ~q(X2,Y,Z).
```

Here, we bind `X1, X2` and generate clauses for all `X1, X2` pairs, where `X1 < X2` holds.\
Restriction `X1 != X2` is also valid, but generates redundant symmetrical literals.

Knowing that `X1 < X`2, hence `X1 != X2`, `Y` is a column index and `Z` is a value,\
`~q(X1,Y,Z), ~q(X2,Y,Z)` evaluates to False if both literals are True. \
We can't ever satisfy this clause with two same values `Z` in different rows in the same column `Y`.

---

**Rule 4**: no two same values in a row

```prolog
dom[Y1], dom[Y2], Y1 < Y2 :: ~q(X,Y1,Z), ~q(X,Y2,Z).
```

Here, we follow the same logic as in **rule 3**, but for rows.

--- 

**Rule 5**: no repeating values in a box

```prolog
boxBegin[ROOTX,ROOTY],
boxOffset[X1,Y1], box[X2,Y2],X1 <= X2, Y1 != Y2
		:: ~q(ROOTX + X1,ROOTY + Y1,Z), ~q(ROOTX + X2,ROOTY + Y2,Z).
```

We bind `ROOTX, ROOTY` to the starting index of our inner box\
We bind `X1, X2, Y` to offset within that box

For any pair `X1, X2` including `X1 == X2`, there can't exist the same value `Z` in different columns `Y1, Y2`.\
This rule is executed for all 9 value-pairs of `ROOTX, ROOTY` (each for 1 box within Sudoku board).\
For `Y1 == Y2`, **rule 3** holds implicitly.

---

We can pre-fill our Sudoku game with literals such as:

```prolog
q(1,1,4).
q(5,3,6).
q(7,9,3).
q(9,9,5).
```

etc. then solve it for that instance


Links
-----
* http://www.satlive.org/ 
* http://www.cs.utexas.edu/users/vl/tag/SAT-grounders
* http://minisat.se/MiniSat+.html
* http://potassco.sourceforge.net/
* http://bach.istc.kobe-u.ac.jp/sugar/
* https://accu.org/journals/overload/27/150/horenovsky_2640/

Paper on Bule
-------------

See folder ./doc




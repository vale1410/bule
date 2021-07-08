module AST = struct module T = struct
type cname = string
type vname = string

type eoperator = Add | Mult | Max | Min
type ord_operator = Lt | Gt | Leq | Geq
type eq_operator = Eq | Neq

type expr = VarE of vname | Int of int | ListE of (eoperator * expr * expr list) | Subtract of (expr * expr list)
type term = Exp of expr | Fun of (cname * term list) | Var of vname
type atom = cname * term list
type ground_literal = In of atom | Notin of atom | Sorted of (expr * ord_operator * expr) | Equal of (term * eq_operator * term)
type tuple = Term of term | Range of (expr * expr)

type glits = ground_literal list

type literals = glits * bool * atom
type ground_decl = glits * cname * tuple list
type search_decl = glits * bool * expr * atom
type clause_decl  = glits * literals list * literals list

type decl = G of ground_decl | S of search_decl | C of clause_decl
type file = decl list
end end

module CIRCUIT = struct module T = struct
type ground_term = Fun of (AST.T.cname * ground_term list)
type search_var = AST.T.cname * ground_term list
type quantifier_block = bool * search_var list
type literal = bool * search_var
type clause = literal list * literal list
type file = quantifier_block list * clause list
end end

module DIMACS = struct module T = struct
type search_var = int
type quantifier_block = bool * search_var list
type literal = bool * search_var
type clause = literal list
type file = int * int * quantifier_block list * clause list
end end

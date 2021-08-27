module AST = struct module T = struct
type cname = string
type vname = string

type eoperator = Add | Div | Log | Max | Min | Mod | Mult | Pow | Sub
type comparison_operator = Lt | Gt | Leq | Geq | Eq | Neq

type expr = VarE of vname | Int of int | BinE of (expr * eoperator * expr)
type term = Exp of expr | Fun of (cname * term list)
type atom = cname * term list
type ground_literal = In of atom | Notin of atom | Comparison of (term * comparison_operator * term) | Set of (vname * term)
type tuple = Term of term | Range of (expr * expr)

type glits = ground_literal list

type literal = bool * atom
type literals = glits * bool * atom
type ground_decl = glits * cname * tuple list
type search_decl = glits * bool * expr * atom list
type clause_decl = glits * literals list * literals list
type hide_decl = glits * literal list

type decl = G of ground_decl | S of search_decl | C of clause_decl | H of hide_decl
type file = decl list
end end

module CIRCUIT = struct module T = struct
type ground_term = Fun of (AST.T.cname * ground_term list)
type search_var = AST.T.cname * ground_term list
type quantifier_block = bool * search_var list
type literal = bool * search_var
type clause = literal list * literal list
type file = quantifier_block list * clause list * literal list
end end

module DIMACS = struct module T = struct
module IMap = Map.Make (Int)
module ISet = Set.Make (Int)
module VMap = Map.Make (struct type t = CIRCUIT.T.search_var let compare = compare end)
type search_var = int
type quantifier_block = bool * search_var list
type literal = bool * search_var
type clause = literal list
type file = int * int * quantifier_block list * clause list
end end


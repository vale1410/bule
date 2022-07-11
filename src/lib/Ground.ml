open Printf

include Types.AST
open T

module P = Print
module Print =
struct
  let id x = x
  let cname = id

  let eoperator = function
    | Add -> "+"
    | Div -> "/"
    | Mult -> "*"
    | Max -> failwith "max operation not supported with gringo"
    | Min -> failwith "min operation not supported with gringo"
    | Log -> failwith "// operation not supported with gringo"
    | Mod -> "\\"
    | Pow -> "**"
    | Sub -> "-"
  let comparison_operator = function
    | Lt -> "<"
    | Gt -> ">"
    | Leq -> "<="
    | Geq -> ">="
    | Eq -> "=="
    | Neq -> "!="

  let rec expr = function
    | BinE (e1, bo, e2) -> sprintf "%s %s %s" (inner_expr e1) (eoperator bo) (inner_expr e2)
    | VarE _ | Int _ as e -> inner_expr e
  and inner_expr = function
    | VarE n -> n
    | Int i -> Print.int i
    | BinE _ as e -> sprintf "(%s)" (expr e)

  let paren_list f = Print.list' "(" ", " ")" f

  let rec term : term -> string = function
    | Exp e -> expr e
    | Fun (c, ts) -> sprintf "%s%s" c (paren_list term ts)

  let atom (name, terms) = sprintf "%s%s" name (paren_list term terms)
  let rec tuple = function
    | ExpTu e -> expr e
    | FunTu (c, ts) -> sprintf "%s%s" c (paren_list tuple ts)
    | Range (e1, e2) -> sprintf "%s..%s" (expr e1) (expr e2)
  let atomd (name, tuples) = sprintf "%s%s" name (paren_list tuple tuples)

  let ground_literal = function
    | In ga -> sprintf "ground(%s)" (atom ga)
    | Notin ga -> sprintf "not ground(%s)" (atom ga)
    | Comparison (t1, c, t2) -> sprintf "%s %s %s" (term t1) (comparison_operator c) (term t2)
    | Set (v, t) -> failwith (sprintf "'%s := %s' not implemented yet for gringo" v (term t))

  let glits gls =
    if gls = [] then "."
    else Print.list' " :- " ", " "." ground_literal gls

  let ground_decl (gls, ats) = sprintf "ground(%s)%s" (atomd ats) (glits gls)

  let ground = Print.unlines ground_decl
end

let run_gringo cmd input =
  let (out, inp, err) = Unix.open_process_full cmd [||] in
  eprintf "gringo:\n%s\n" input;
  fprintf inp "%s" input;
  flush inp;
  close_out inp;
  let out = Misc.read_in_channel out in
  let err = Misc.read_in_channel err in
  (out, err)

let all_ground cmd gs =
  let f (glits, facts) = Misc.map (fun fact -> (glits, fact)) facts in
  let flat = List.concat_map f gs in
  let out, err = run_gringo cmd (Print.ground flat) in
  let facts = Parse.facts out in
  let aux = function
    | ("ground", [g]) -> g
    | _ -> assert false in
  let grounds = List.rev_map aux facts in
  ignore err;
  grounds


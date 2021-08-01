open Printf

include Types.AST
open T

module Print =
struct
  let id x = x
  let cname = id

  let eoperator_aux = function
    | Add -> "+"
    | Div -> "/"
    | Mult -> "*"
    | Max -> "\\max"
    | Min -> "\\min"
    | Log -> "//"
    | Mod -> "#mod"
    | Pow -> "**"
    | Sub -> "-"
  let comparison_operator = function
    | Lt -> "<"
    | Gt -> ">"
    | Leq -> "<="
    | Geq -> ">="
    | Eq -> "=="
    | Neq -> "!="
  let eoperator o = "\\" ^ eoperator_aux o

  let list_tuple f = Print.list' "(" ", " ")" f

  let rec expr = function
    | BinE (e1, bo, e2) -> sprintf "%s %s %s" (inner_expr e1) (eoperator bo) (inner_expr e2)
    | VarE _ | Int _ as e -> inner_expr e
  and inner_expr = function
    | VarE n -> n
    | Int i -> Print.int i
    | BinE _ as e -> sprintf "(%s)" (expr e)
  let rec atom (name, terms) = sprintf "%s[%s]" name (Print.list' "" ", " "" term terms)
  and ground_literal = function
    | In ga -> atom ga
    | Notin ga -> sprintf "~%s" (atom ga)
    | Comparison (t1, c, t2) -> sprintf "%s %s %s" (term t1) (comparison_operator c) (term t2)
    | Set (v, t) -> sprintf "%s := %s" v (term t)
  and term : term -> string = function
    | Exp e -> expr e
    | Fun (c, ts) -> sprintf "%s%s" c (list_tuple term ts)
  let tuple = function
    | Term t -> term t
    | Range (e1, e2) -> sprintf "%s..%s" (expr e1) (expr e2)

  let glits gls = Print.list' "" ", " "" ground_literal gls
  let search_var (name, terms) = sprintf "%s(%s)" name (Print.list' "" ", " "" term terms)
  let literals (gls, pol, var) =
    let gls = match gls with
    | [] -> ""
    | _ :: _ -> glits gls ^ " : " in
    let pol = if pol then "" else "~" in
    let var = search_var var in
    gls ^ pol ^ var

  let ground_decl (gls, name, tuples) = sprintf "%s :: %s[%s]" (glits gls) name (Print.list' "" ", " "" tuple tuples)
  let search_decl (gls, exi, depth, var) =
    let quant = sprintf "#%s[%s]" (if exi then "exists" else "forall") (expr depth) in
    let gls = match gls with [] -> "" | _ :: _ -> ", " ^ glits gls in
    quant ^ gls ^ " :: " ^ search_var var ^ "?"
  let clause_decl (gls, hyps, ccls) = sprintf "%s :: %s -> %s." (glits gls) (Print.list' "" " & " "" literals hyps) (Print.list' "" " | " "" literals ccls)
  let hide_decl (gls, var) =
    let gls = match gls with [] -> "" | _ :: _ -> ", " ^ glits gls in
    sprintf "%%#hide %s :: %s." gls (search_var var)
  let decl = function
    | G gd -> ground_decl gd
    | S sd -> search_decl sd
    | C cd -> clause_decl cd
    | H hd -> hide_decl hd

  let file = Print.unlines decl
end

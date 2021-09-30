%token <string> VNAME
%token <string> CNAME
%token <int> INT
%token UNDERSCORE
%token NOT
%token CLAUSE EXISTS FORALL GROUND HIDE (*QMARK*)
%token CONJ DISJ
%token LPAREN RPAREN LBRACKET RBRACKET
%token DEFINE DCOLON COLON IMPLIED IMPLIES COMMA DOT RANGE
%token DIV PLUS MULT LOG MOD POW MINUS (*eop*)
%token EQ NEQ LEQ GEQ LT GT
(*%token MAX MIN beop*)
%token EOF

%left MOD
%left PLUS MINUS
%left MULT DIV
%right POW LOG

%{
(*module Ast = Ast.T*)
let unroll_comparion_chain t l =
  let aux (accu_l, accu_t) (op, t) =
    (Ast.T.Comparison (accu_t, op, t) :: accu_l, t) in
  fst (List.fold_left aux ([], t) l)

let add_list l = function
  | Ast.T.G (l', ts) -> Ast.T.G (l @ l', ts)
  | Ast.T.S (l', b, e, a) -> Ast.T.S (l @ l', b, e, a)
  | Ast.T.C (l', hyps, ccls) -> Ast.T.C (l @ l', hyps, ccls)
  | Ast.T.H (l', a) -> Ast.T.H (l @ l', a)

let make_var =
  let counter = ref 0 in
  (fun () -> incr counter; Printf.sprintf "__v%d" (!counter))
%}

(*%type <(Ast.keyword, Ast.free) Ast.atomic> keyword_atomic*)
%type <Ast.T.ground_literal list> ground_literal
(*%type <Ast.T.search_decl> search_decl*)
%type <Ast.T.file> file
%start file
%%

%public %inline iboption(X):
| /* nothing */ { false }
| X { true }

separated_many_slist(Sep, Sub):
| a = Sub s = Sep r = separated_nonempty_list(Sep, Sub) { (s, a, r) }

%inline co_list(X):
| l = separated_nonempty_list(COMMA, X) { l }
%inline pr_list(X):
| { [] }
| LPAREN l = separated_list(COMMA, X) RPAREN { l }
%inline br_list(X):
| LBRACKET l = separated_list(COMMA, X) RBRACKET { l }

%inline eoperator: | DIV { Ast.T.Div } | LOG { Ast.T.Log } | MOD { Ast.T.Mod } | MULT { Ast.T.Mult } | POW { Ast.T.Pow } | PLUS { Ast.T.Add } | MINUS { Ast.T.Sub }
%inline loperator: | LT { Ast.T.Lt } | LEQ { Ast.T.Leq }
%inline goperator: | GT { Ast.T.Gt } | GEQ { Ast.T.Geq }
%inline qoperator: | EQ { Ast.T.Eq } | NEQ { Ast.T.Neq }

expr:
| e1 = expr bo = eoperator e2 = expr { Ast.T.BinE (e1, bo, e2) }
| LPAREN e = expr RPAREN { e }
| n = VNAME { Ast.T.VarE n }
| i = INT { Ast.T.Int i }
| MINUS e = expr { Ast.T.BinE (Ast.T.Int 0, Ast.T.Sub, e) }

term:
| name = CNAME ts = pr_list(term) { Ast.T.Fun (name, ts)  }
| UNDERSCORE { Ast.T.Exp (Ast.T.VarE (make_var ())) }
| e = expr { Ast.T.Exp e }

tuple:
| name = CNAME ts = pr_list(tuple) { Ast.T.FunTu (name, ts)  }
| e = expr { Ast.T.ExpTu e }
| e1 = expr RANGE e2 = expr { Ast.T.Range (e1, e2) }

atom:
| n = CNAME ts = pr_list(term) { (n, ts) }
%inline literal:
| pol = iboption(NOT) a = atom { (not pol, a) }
%inline ground_atom:
| n = CNAME ts = br_list(term) { (n, ts) }
%inline ground_atomd:
| n = CNAME ts = br_list(tuple) { (n, ts) }
%inline search_atomd:
| n = CNAME ts = pr_list(tuple) { (n, ts) }
ground_literal:
| ga = ground_atom { [Ast.T.In ga] }
| NOT ga = ground_atom { [Ast.T.Notin ga] }
| ch = chain { ch }
| e1 = term o = qoperator e2 = term { [Ast.T.Comparison (e1, o, e2)] }
| v = VNAME DEFINE t = term { [Ast.T.Set (v, t)] }
chain:
| t = term l = nonempty_list(pair(loperator,term)) { unroll_comparion_chain t l }
| t = term l = nonempty_list(pair(goperator,term)) { unroll_comparion_chain t l }

grounding_prefix:
gls = separated_list(COMMA, ground_literal) { List.flatten gls }
nonempty_grounding_prefix:
gls = separated_nonempty_list(COMMA, ground_literal) { List.flatten gls }

%inline quantifier: | EXISTS { true } | FORALL { false }
quantifier_block:
| b = quantifier LBRACKET e = expr RBRACKET { (b, e) }
literals:
| gp = nonempty_grounding_prefix COLON pa = literal { let (pol, a) = pa in (gp, pol, a) }
| pa = literal { let (pol, a) = pa in ([], pol, a) }
clause_body:
hyps = separated_list(CONJ, literals) { hyps }
clause_head:
ccls = separated_list(DISJ, literals) { ccls }
clause_part:
| hyps = clause_body IMPLIES ccls = clause_head { ([], hyps, ccls) }
| ccls = clause_head IMPLIED hyps = clause_body { ([], hyps, ccls) }
| ccls = clause_head { ([], [], ccls) }
(*pre_decl:
| gd = ground_head { Ast.T.G gd }
| sd = search_decl { Ast.T.S sd }
| cd = clause_part { Ast.T.C cd }
| hd = hide_decl { Ast.T.H hd }

decl:
| gp = grounding_prefix DCOLON d = pre_decl DOT { add_list gp d }*)
pre_decl:
| GROUND gd = co_list(ground_atomd) { Ast.T.G ([], gd) }
| qb = quantifier_block vars = co_list(search_atomd) { Ast.T.S ([], fst qb, snd qb, vars) }
| cd = clause_part { (Ast.T.C cd) }
| HIDE hd = co_list(literal) { Ast.T.H ([], hd) }

decl:
| gp = grounding_prefix DCOLON d = pre_decl DOT { add_list gp d }
| d = pre_decl DOT { d }

file:
| l = list(decl) EOF { l }


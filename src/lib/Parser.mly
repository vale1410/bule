%token <string> VNAME
%token <string> CNAME
%token <int> INT
%token UNDERSCORE
%token NOT
%token EXISTS FORALL GROUND HIDE SHOW
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
let make_var =
  let counter = ref 0 in
  (fun () -> incr counter; Printf.sprintf "__V%d" (!counter))

open Types
%}

(*%type <Types.PARSE.T.search_decl> search_decl*)
%type <Types.PARSE.T.file> file
%type <(Types.AST.T.cname * Types.CIRCUIT.T.search_var list) list> ground_gringo
%start file ground_gringo
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

%inline eoperator: | DIV { AST.T.Div } | LOG { AST.T.Log } | MOD { AST.T.Mod } | MULT { AST.T.Mult } | POW { AST.T.Pow } | PLUS { AST.T.Add } | MINUS { AST.T.Sub }

%inline c_loperator:  | LT  { AST.T.Lt } | LEQ { AST.T.Leq }
%inline c_goperator:  | GT  { AST.T.Gt } | GEQ { AST.T.Geq }
%inline c_eloperator: | EQ  { AST.T.Eq } | LT  { AST.T.Lt } | LEQ { AST.T.Leq }
%inline c_egoperator: | EQ  { AST.T.Eq } | GT  { AST.T.Gt } | GEQ { AST.T.Geq }
%inline c_eoperator:  | EQ  { AST.T.Eq }
%inline c_noperator:  | NEQ { AST.T.Neq }

expr:
| e1 = expr bo = eoperator e2 = expr { AST.T.BinE (e1, bo, e2) }
| LPAREN e = expr RPAREN { e }
| n = VNAME { AST.T.VarE n }
| i = INT   { AST.T.Int i }
| MINUS e = expr { AST.T.BinE (AST.T.Int 0, AST.T.Sub, e) }

term:
| name = CNAME ts = pr_list(term) { AST.T.Fun (name, ts)  }
| UNDERSCORE { AST.T.Exp (AST.T.VarE (make_var ())) }
| e = expr { AST.T.Exp e }

tuple:
| name = CNAME ts = pr_list(tuple) { AST.T.FunTu (name, ts)  }
| e = expr { AST.T.ExpTu e }
| e1 = expr RANGE e2 = expr { AST.T.Range (e1, e2) }

atom:
| n = CNAME ts = pr_list(term) { (n, ts) }
literal:
| pol = iboption(NOT) a = atom { (not pol, a) }
grounding_atom:
| n = CNAME ts = br_list(term) { (n, ts) }
ground_atomd:
| n = CNAME ts = br_list(tuple) { (n, ts) }
search_atomd:
| n = CNAME ts = pr_list(tuple) { (n, ts) }
ground_literal:
| ga = grounding_atom { PARSE.T.In ga }
| NOT ga = grounding_atom { PARSE.T.Notin ga }
| ch = chain { PARSE.T.Chain ch }
| e1 = term o = c_noperator e2 = term { PARSE.T.Chain (e1, [o, e2]) }
| v = VNAME DEFINE t = term { PARSE.T.Set (v, t) }
chain:
(*| t = term l = nonempty_list(pair(loperator,term)) { (t, l) }
| t = term l = nonempty_list(pair(goperator,term)) { (t, l) }*)
| t1 = term l1 = list(pair(c_eoperator, term)) o = c_loperator t2 = term l2 = list(pair(c_eloperator,term)) { (t1, l1 @ (o, t2) :: l2) }
| t1 = term l1 = list(pair(c_eoperator, term)) o = c_goperator t2 = term l2 = list(pair(c_egoperator,term)) { (t1, l1 @ (o, t2) :: l2) }
| t = term l = nonempty_list(pair(c_eoperator,term)) { (t, l) }

grounding_prefix:
gls = separated_list(COMMA, ground_literal) { gls }

literals:
| gp = grounding_prefix COLON pa = literal { let (pol, a) = pa in (gp, pol, a) }
| pa = literal { let (pol, a) = pa in ([], pol, a) }
clause_body:
hyps = separated_list(CONJ, literals) { hyps }
clause_head:
ccls = separated_list(DISJ, literals) { ccls }
clause_part:
| hyps = clause_body IMPLIES ccls = clause_head { (hyps, ccls) }
| ccls = clause_head IMPLIED hyps = clause_body { (hyps, ccls) }
| ccls = clause_head { ([], ccls) }

quantifier_block:
| LBRACKET e = expr RBRACKET { e }
pre_decl:
| GROUND gd = co_list(ground_atomd) { PARSE.T.G gd }
| EXISTS                       vars = co_list(search_atomd) { PARSE.T.S (AST.T.ExistentialInnerMost vars) }
| EXISTS qb = quantifier_block vars = co_list(search_atomd) { PARSE.T.S (AST.T.Level (true,  qb, vars)) }
| FORALL qb = quantifier_block vars = co_list(search_atomd) { PARSE.T.S (AST.T.Level (false, qb, vars)) }
| cd = clause_part { PARSE.T.C cd }
| HIDE hd = co_list(literal) { PARSE.T.H (true, hd) }
| SHOW hd = co_list(literal) { PARSE.T.H (false, hd) }

decl:
| gp = grounding_prefix DCOLON d = pre_decl DOT { (gp, d) }
| d = pre_decl DOT { ([], d) }

file:
| l = list(decl) EOF { l }

ground_term:
| i = INT { CIRCUIT.T.Fun (string_of_int i, []) }
| name = CNAME ts = pr_list(ground_term) { CIRCUIT.T.Fun (name, ts) }
ground_atom:
| name = CNAME ts = pr_list(ground_term) { (name, ts) }
fact:
| ground = CNAME LPAREN a = ground_atom RPAREN DOT { (ground, [a]) }
| quanti = CNAME LPAREN d = INT COMMA a = ground_atom RPAREN DOT { (quanti, [(string_of_int d, []); a]) }
| hidesh = CNAME LPAREN s = CNAME COMMA a = ground_atom RPAREN DOT { (hidesh, [(s, []); a]) }

ground_gringo:
| l = list(fact) EOF { l }

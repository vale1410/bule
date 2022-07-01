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
  (fun () -> incr counter; Printf.sprintf "__v%d" (!counter))
%}

(*%type <Types.PARSE.T.search_decl> search_decl*)
%type <Types.PARSE.T.file> file
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

%inline c_loperator: | LT { Ast.T.Lt } | LEQ { Ast.T.Leq }
%inline c_goperator: | GT { Ast.T.Gt } | GEQ { Ast.T.Geq }
%inline c_eloperator: | EQ { Ast.T.Eq } | LT { Ast.T.Lt } | LEQ { Ast.T.Leq }
%inline c_egoperator: | EQ { Ast.T.Eq } | GT { Ast.T.Gt } | GEQ { Ast.T.Geq }
%inline c_eoperator: | EQ { Ast.T.Eq }
%inline c_noperator: | NEQ { Ast.T.Neq }

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
literal:
| pol = iboption(NOT) a = atom { (not pol, a) }
ground_atom:
| n = CNAME ts = br_list(term) { (n, ts) }
ground_atomd:
| n = CNAME ts = br_list(tuple) { (n, ts) }
search_atomd:
| n = CNAME ts = pr_list(tuple) { (n, ts) }
ground_literal:
| ga = ground_atom { Types.PARSE.T.In ga }
| NOT ga = ground_atom { Types.PARSE.T.Notin ga }
| ch = chain { Types.PARSE.T.Chain ch }
| e1 = term o = c_noperator e2 = term { Types.PARSE.T.Chain (e1, [o, e2]) }
| v = VNAME DEFINE t = term { Types.PARSE.T.Set (v, t) }
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
| GROUND gd = co_list(ground_atomd) { Types.PARSE.T.G gd }
| EXISTS                       vars = co_list(search_atomd) { Types.PARSE.T.S (Ast.T.ExistentialInnerMost vars) }
| EXISTS qb = quantifier_block vars = co_list(search_atomd) { Types.PARSE.T.S (Ast.T.Level (true,  qb, vars)) }
| FORALL qb = quantifier_block vars = co_list(search_atomd) { Types.PARSE.T.S (Ast.T.Level (false, qb, vars)) }
| cd = clause_part { (Types.PARSE.T.C cd) }
| HIDE hd = co_list(literal) { Types.PARSE.T.H (true, hd) }
| SHOW hd = co_list(literal) { Types.PARSE.T.H (false, hd) }

decl:
| gp = grounding_prefix DCOLON d = pre_decl DOT { (gp, d) }
| d = pre_decl DOT { ([], d) }

file:
| l = list(decl) EOF { l }


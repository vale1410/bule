%token <string> VNAME
%token <string> CNAME
%token <int> INT
%token NOT
%token FORALL EXISTS (*QMARK*)
%token CONJ DISJ
%token LPAREN RPAREN LBRACKET RBRACKET
%token DEFINE COLON IMPLIES COMMA DOT RANGE
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
let add_list l = function
  | Ast.T.G (l', c, ts) -> Ast.T.G (l @ l', c, ts)
  | Ast.T.S (l', b, e, a) -> Ast.T.S (l @ l', b, e, a)
  | Ast.T.C (l', hyps, ccls) -> Ast.T.C (l @ l', hyps, ccls)
%}
(*%type <(Ast.keyword, Ast.free) Ast.atomic> keyword_atomic*)
%type <Ast.T.ground_literal> ground_literal
%type <Ast.T.search_decl> search_decl
%type <Ast.T.file> file
%start file
%%

%public %inline iboption(X):
| /* nothing */ { false }
| X { true }

separated_many_slist(Sep, Sub):
| a = Sub s = Sep r = separated_nonempty_list(Sep, Sub) { (s, a, r) }

%inline pr_list(X):
| { [] }
| LPAREN l = separated_list(COMMA, X) RPAREN { l }
%inline br_list(X):
| LBRACKET l = separated_list(COMMA, X) RBRACKET { l }

%inline eoperator: | DIV { Ast.T.Div } | LOG { Ast.T.Log } | MOD { Ast.T.Mod } | MULT { Ast.T.Mult } | POW { Ast.T.Pow } | PLUS { Ast.T.Add } | MINUS { Ast.T.Sub }
%inline coperator: | LT { Ast.T.Lt } | GT { Ast.T.Gt } | LEQ { Ast.T.Leq } | GEQ { Ast.T.Geq } | EQ { Ast.T.Eq } | NEQ { Ast.T.Neq }

expr:
| e1 = expr bo = eoperator e2 = expr { Ast.T.BinE (e1, bo, e2) }
| LPAREN e = expr RPAREN { e }
| n = VNAME { Ast.T.VarE n }
| i = INT { Ast.T.Int i }
| MINUS e = expr { Ast.T.BinE (Ast.T.Int 0, Ast.T.Sub, e) }

term:
| name = CNAME ts = pr_list(term) { Ast.T.Fun (name, ts)  }
| e = expr { Ast.T.Exp e }

%inline tuple:
| t = term { Ast.T.Term t }
| e1 = expr RANGE e2 = expr { Ast.T.Range (e1, e2) }

%inline search_atom:
| pol = iboption(NOT) n = CNAME ts = pr_list(term) { (not pol, (n, ts)) }
%inline ground_atom:
| n = CNAME ts = br_list(term) { (n, ts) }
%inline ground_head:
| n = CNAME ts = br_list(tuple) { ([], n, ts) }
%inline ground_literal:
| ga = ground_atom { Ast.T.In ga }
| NOT ga = ground_atom { Ast.T.Notin ga }
| e1 = term o = coperator e2 = term { Ast.T.Comparison (e1, o, e2) }

%inline quantifier: | EXISTS { true } | FORALL { false }
%inline search_decl:
| b = quantifier LBRACKET e = expr RBRACKET n = CNAME ts = pr_list(term) { ([], b, e, (n, ts)) }
%inline literals:
| COLON gls = separated_nonempty_list(COMMA, ground_literal) COLON pa = search_atom { let (pol, a) = pa in (gls, pol, a) }
| pa = search_atom { let (pol, a) = pa in ([], pol, a) }
%inline clause_part:
| hyps = separated_list(CONJ, literals) IMPLIES ccls = separated_list(DISJ, literals) { ([], hyps, ccls) }
| ccls = separated_list(DISJ, literals) { ([], [], ccls) }
%inline pre_decl:
| gd = ground_head { Ast.T.G gd }
| sd = search_decl { Ast.T.S sd }
| cd = clause_part { Ast.T.C cd }

decl:
| DEFINE gls = separated_list(COMMA, ground_literal) DEFINE d = pre_decl DOT { add_list gls d }
| d = pre_decl DOT { d }

file:
| l = list(decl) EOF { l }


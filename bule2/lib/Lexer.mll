{
  module P = Parser
  exception Error of string

  let incr_linenum lexbuf =
    let pos = lexbuf.Lexing.lex_curr_p in
    lexbuf.Lexing.lex_curr_p <-
      { pos with
        Lexing.pos_lnum = pos.Lexing.pos_lnum + 1;
        Lexing.pos_bol = pos.Lexing.pos_cnum }

  let print_error lexer =
    let position = Lexing.lexeme_start_p lexer in
    let line = position.Lexing.pos_lnum
    and char = position.Lexing.pos_cnum - position.Lexing.pos_bol in
    Printf.sprintf "Lexer : at line %d, column %d: unexpected character.\n%!" line char

  let string_of_char c = Printf.sprintf "%c" c
}

let digit = ['0' - '9']
let lcase = ['a' - 'z']
let ucase = ['A' - 'Z']
(*let letter = lcase | ucase | digit | ['_'] | ['-'] | ['+'] | ['<'] | ['>'] | ['*'] | ['/'] | ['='] | ['@'] | ['#']*)
let letter = lcase | ucase | digit | ['_']
let vname = ucase letter*
let cname = lcase letter*
let integer = digit+
let linefeed = "\r\n" | ['\n' '\r']

rule line_comment = parse
  | ([^'\n']* '\n') { incr_linenum lexbuf; token lexbuf }
  | ([^'\n']* eof) { P.EOF }
  | _   { raise (Error (print_error lexbuf)) }

and token = parse
  | '(' { P.LPAREN   } | ')' { P.RPAREN   } | '[' { P.LBRACKET } | ']' { P.RBRACKET }
  | '<' { P.LT } | '>' { P.GT } | "<=" { P.LEQ } | ">=" { P.GEQ }
  | "==" { P.EQ } | "!=" { P.NEQ }
  | ".." { P.RANGE }
  | '/' { P.DIV } | '+' { P.PLUS } | '-' { P.MINUS } | "*" { P.MULT } | "//" { P.LOG } | "#mod" { P.MOD } | "**" { P.POW }
  | "~" { P.NOT } | "->" { P.IMPLIES }
  | ',' { P.COMMA } | "::" { P.DEFINE } | ':' { P.COLON } | '.' { P.DOT } (*| "?" { P.QMARK }*)
  | [' ' '\t'] { token lexbuf }
  | linefeed   { incr_linenum lexbuf; token lexbuf }
  | vname as n { P.VNAME n }
  | cname as n { P.CNAME n }
  | "#exists" { P.EXISTS } | "#forall" { P.FORALL }
  | "|" { P.DISJ } | "&" { P.CONJ }
  | integer as i { P.INT (int_of_string i) }
  | eof  { P.EOF }
  | "(*" { comment lexbuf }
  | "%"  { line_comment lexbuf }
  | _    { raise (Error (print_error lexbuf)) }
  (*| variable as v { let v' = Scanf.sscanf v "?%s" (fun s -> s) in P.VARIABLE v' }*)

and comment = parse
  | "*)"     { token lexbuf }
  | linefeed { incr_linenum lexbuf; comment lexbuf }
  | _        { comment lexbuf }

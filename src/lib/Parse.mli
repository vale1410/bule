val from_file : unit -> string -> Types.PARSE.T.file
val facts : string -> (Types.AST.T.cname * Types.CIRCUIT.T.search_var list) list
val clingo_facts : string -> (Types.AST.T.cname * Types.CIRCUIT.T.search_var list) list

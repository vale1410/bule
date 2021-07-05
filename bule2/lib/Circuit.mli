include module type of Types.CIRCUIT with module T = Types.CIRCUIT.T

module Print : sig
  val search_var : T.search_var -> string
  val file : T.file -> string
end

val file : Ast.T.file -> T.file

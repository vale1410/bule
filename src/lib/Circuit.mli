include module type of Types.CIRCUIT with module T = Types.CIRCUIT.T

module Print : sig
  val search_var : T.search_var -> string
  val literal : T.literal -> string
  val file : T.file -> string
end

val file : bool -> Ast.T.file -> T.file

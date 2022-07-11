include module type of Types.CIRCUIT with module T = Types.CIRCUIT.T

module Print : sig
  val ground_term : T.ground_term -> string
  val search_var : T.search_var -> string
  val literal : T.literal -> string
  val file : T.file -> string
end

type grounder = Native | CommandLine of string

val file : bool -> grounder -> Ast.T.file -> T.file

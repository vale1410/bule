include module type of Types.DIMACS with module T = Types.DIMACS.T

module Print : sig
  val file : T.file -> string
end

val file : Circuit.T.file -> T.file

include module type of Types.DIMACS with module T = Types.DIMACS.T

module Print : sig
  val literal : T.literal -> string
  val quantifier_block : T.quantifier_block -> string
  val sat_file : T.file -> string
  val qbf_file : T.file -> string
end

val ground : Circuit.T.file -> T.file * int T.VMap.t * Circuit.T.search_var T.IMap.t * T.ISet.t
val file : Circuit.T.file -> T.file

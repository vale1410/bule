include module type of Types.AST with module T = Types.AST.T

module Print :
sig
  val cname : T.cname -> string
  val term : T.term -> string
  val tuple : T.tuple -> string
  val atom : T.atom -> string
  val eoperator : T.eoperator -> string
  val ground_decl : (T.glits * T.ground_decl) -> string
  val file : T.file -> string
end

val file : Types.PARSE.T.file -> T.file

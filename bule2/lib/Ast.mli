include module type of Types.AST with module T = Types.AST.T

module Print :
sig
  val cname : T.cname -> string
  val term : T.term -> string
  val atom : T.atom -> string
  val eoperator : T.eoperator -> string
  val decl : T.decl -> string
  val file : T.file -> string
end

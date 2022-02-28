val bits : int -> string
val bool : bool -> string
val int : int -> string
val float : float -> string
val string : string -> string
val couple' : string -> string -> string -> ('a -> string) -> ('b -> string) -> ('a * 'b) -> string
val couple : ('a -> string) -> ('b -> string) -> ('a * 'b) -> string
val option : ('a -> string) -> 'a option -> string
val list' : string -> string -> string -> ('a -> string) -> 'a list -> string
val list : ('a -> string) -> 'a list -> string
val unlines : ('a -> string) -> 'a list -> string
val unspaces : ('a -> string) -> 'a list -> string
val array : ('a -> string) -> 'a array -> string
val array_of_list : ('a -> string) -> 'a list -> string
val tuple_of_list : ('a -> string) -> 'a list -> string
val matrix : ('a -> string) -> 'a array array -> string
val to_file : string -> string -> unit
val pass : 'a -> string

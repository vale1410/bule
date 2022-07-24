open Printf

let rec pow x n = if n <= 0 then 1 else x * pow x (n - 1)
let bits b =
  let s = ref "" in
  for i = 0 to 8 do
    s := sprintf "%s%d" !s ((b / pow 2 i) mod 2)
  done;
  !s

let bool = sprintf "%b"
let int = sprintf "%d"
let float = sprintf "%f"
let string = sprintf "%s"
let couple' ld id rd pr1 pr2 (x, y) = sprintf "%s%s%s%s%s" ld (pr1 x) id (pr2 y) rd
let couple pr1 pr2 = couple' "(" ", " ")" pr1 pr2
let option pr a = match a with | None -> "None" | Some x -> sprintf "Some %s" (pr x)
let list' ld id rd pr l = sprintf "%s%s%s" ld (String.concat id (List.rev_map pr (List.rev l))) rd
let list pr = list' "[" "; " "]" pr
let unlines pr = list' "" "\n" "" pr
let unspaces pr = list' "" " " "" pr
let array pr l = list' "[|" "; " "|]" pr (Array.to_list l)
let array_of_list pr = list' "[|" "; " "|]" pr
let tuple_of_list pr = list' "(" ", " ")" pr
let matrix pr m = list' "[|" ";\n  " "|]" (array pr) (Array.to_list m)

let to_file filename str =
  let file = open_out filename in
  fprintf file "%s" str;
  close_out file

let pass _ = "."

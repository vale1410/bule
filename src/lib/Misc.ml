let range i j =
  let rec aux accu n =
    if n < i then accu
    else aux (n :: accu) (n - 1) in
  aux [] (j - 1)

let sum l = List.fold_left (+) 0 l
let product l = List.fold_left ( * ) 1 l


let function_inverse f b a = f a b
let cons a r = a :: r
let flatten l = List.rev (List.fold_left (function_inverse List.rev_append) [] l)
let map f l = List.rev (List.rev_map f l)
let cross_products l =
  let aux accu elts = flatten (map (fun elt -> map (cons elt) accu) elts) in
  let results = List.fold_left aux [[]] l in
  map List.rev results

let read_in_channel inc =
  let maybe_read_line () =
    try Some(input_line inc)
    with End_of_file -> close_in inc; None in
  let rec loop acc =
    match maybe_read_line () with
    | Some(line) -> loop (line :: acc)
    | None -> List.rev acc in
  Print.list' "" "\n" "" Print.string (loop [])


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
let filter_map f l = fst (List.partition_map (fun x -> match f x with | Some y -> Either.Left y | None -> Either.Right None) l)
let cross_products l =
  let aux accu elts = flatten (map (fun elt -> map (cons elt) accu) elts) in
  let results = List.fold_left aux [[]] l in
  map List.rev results

let read_in_channel inc =
  let maybe_read_line () =
    try Some (input_line inc)
    with End_of_file -> close_in inc; None in
  let rec loop acc =
    match maybe_read_line () with
    | Some line -> loop (line :: acc)
    | None -> List.rev acc in
  Print.unlines Print.string (loop [])

let is_even n =
  n mod 2 = 0

(* https://stackoverflow.com/a/37184495 *)
let pow base exponent =
  assert (exponent >= 0);
  let rec aux accu base = function
    | 0 -> accu
    | 1 -> base * accu
    | e -> if e mod 2 = 0 then aux accu (base * base) (e / 2) else aux (base * accu) (base * base) ((e - 1) / 2) in
  aux 1 base exponent

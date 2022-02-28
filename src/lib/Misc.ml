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

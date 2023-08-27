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

let remove_first x l =
  let rec aux before = function
    | [] -> None
    | a :: rest when a = x -> Some (List.rev_append before rest)
    | a :: rest -> aux (a :: before) rest in
  aux [] l
let replace x y l = map (fun z -> if z = x then y else x) l

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

let input_lines inp =
  let l = ref [] in
  (try
     while true do l := input_line inp :: !l done
   with End_of_file -> ());
  List.rev !l

let run_process cmd question =
  let stdin_name  = Filename.temp_file "bule." ".in" in
  let stdout_name = Filename.temp_file "bule." ".out" in
  let stderr_name = Filename.temp_file "bule." ".err" in
  let (cmd, args) = match Str.split (Str.regexp "[ ]+") cmd with
    | [] -> assert false
    | hd :: tl -> (hd, tl) in
  let cmd = Filename.quote_command cmd ~stdin:stdin_name ~stdout:stdout_name ~stderr:stderr_name args in
  Print.to_file stdin_name question;
  let status = Unix.system cmd in
  let stdout_f = open_in stdout_name
  and stderr_f = open_in stderr_name in
  let out_lines = input_lines stdout_f
  and err_lines = input_lines stderr_f in
  close_in stdout_f;
  close_in stderr_f;
  List.iter Sys.remove [stdin_name; stdout_name; stderr_name];
  let answer = match status with
    | Unix.WEXITED code -> Either.Left code
    | _ -> Either.Right status in
  (answer, out_lines, err_lines)

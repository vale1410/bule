open Printf

let parse_status line =
  Scanf.sscanf line "s cnf %d %d %d" (fun p _ _ -> p = 1)

let parse_line line =
  let abs x = assert (x <> 0); if x < 0 then (false, -x) else (true, x) in
  Scanf.sscanf line "V %d 0" abs

let parse_output = function
  | [] -> assert false
  | status :: rest ->
     let sat = parse_status status in
     if sat then Some (List.map parse_line rest) else None

let input_lines inp =
  let l = ref [] in
  (try
    while true do l := input_line inp :: !l done
  with End_of_file -> ());
  List.rev !l

let run_process cmd dimacs =
  let inp, out = Unix.open_process cmd in
  fprintf out "%s%!" (Dimacs.Print.file dimacs);
  close_out out;
  let r = input_lines inp in
  close_in inp;
  r

module IntSet = Set.Make (Int)

let run_solver cmd keys dimacs =
  let lines = run_process cmd dimacs in
  match parse_output lines with
  | None -> None
  | Some model ->
     let assigned = List.fold_left (fun accu (_, x) -> IntSet.add x accu) IntSet.empty model in
     let missing = IntSet.diff keys assigned in
     let actual_model = IntSet.fold (fun x l -> (false, x) :: l) missing model in
     Some actual_model

let print_literal imap (pol, var) =
  let tilde = if pol then " " else "~" in
  let sv = match Dimacs.T.IMap.find_opt var imap with None -> assert false | Some sv -> sv in
    sprintf "%s%s" tilde (Circuit.Print.search_var sv)
let compare_literals (px, x) (py, y) = -(compare (x, px) (y, py))
let print_one_model imap model hide =
  let model = List.filter (fun (_, x) -> not (Dimacs.T.ISet.mem x hide)) model in
  Print.unlines (print_literal imap) (List.sort compare_literals model)

let print_all_models imap model hide =
  let model = List.filter (fun (_, x) -> not (Dimacs.T.ISet.mem x hide)) model in
  Print.unspaces (print_literal imap) (List.sort compare_literals model)

let map_keys map =
  let add k _ set = IntSet.add k set in
  Dimacs.T.IMap.fold add map IntSet.empty
let solve_one cmd (dimacs, _, imap, hide) =
  let keys = map_keys imap in
  match run_solver cmd keys dimacs with
  | None -> printf "UNSAT\n"
  | Some model -> printf "SAT\n";
                  eprintf "%s\n" (print_one_model imap model hide)

let next_instance (nbvar, nbcls, blocks, cls) model =
  let flip_literal (pol, var) = (not pol, var) in
  let nmodel = List.map flip_literal model in
  (nbvar, nbcls + 1, blocks, nmodel :: cls)

let solve_all cmd bound (dimacs, _, imap, hide) =
  eprintf "Instance ground. Starts solving\n%!";
  let keys = map_keys imap in
  let rec aux i dm =
    if i >= bound && bound > 0 then ()
    else
      match run_solver cmd keys dm with
      | None ->
         if i = 0 then printf "UNSAT\n%!";
         eprintf "No more models. Total: %d.\n%!" i
      | Some model ->
         if i = 0 then printf "SAT\n%!";
         eprintf "Model %d: %s\n%!" (i+1) (print_all_models imap model hide);
         aux (i+1) (next_instance dm model) in
  aux 0 dimacs

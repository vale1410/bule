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

let run_solver cmd dimacs =
  let lines = run_process cmd dimacs in
  parse_output lines

let print_model imap model =
  let pr_one (pol, var) =
    let tilde = if pol then "" else "~" in
    let sv = match Dimacs.T.IMap.find_opt var imap with None -> assert false | Some sv -> sv in
    sprintf "%s%s" tilde (Circuit.Print.search_var sv) in
  Print.unlines pr_one model

let solve_one cmd (dimacs, _, imap) =
  match run_solver cmd dimacs with
  | None -> printf "UNSAT\n"
  | Some model -> printf "SAT\n";
                  eprintf "%s\n" (print_model imap model)

let next_instance (nbvar, nbcls, blocks, cls) model =
  let flip_literal (pol, var) = (not pol, var) in
  let nmodel = List.map flip_literal model in
  (nbvar, nbcls + 1, blocks, nmodel :: cls)

let solve_all cmd bound (dimacs, _, imap) =
  let rec aux i dm =
    if i >= bound && bound > 0 then ()
    else
      match run_solver cmd dm with
      | None ->
         if i = 0 then printf "UNSAT\n%!";
         eprintf "No more models. Total: %d.\n%!" i
      | Some model ->
         if i = 0 then printf "SAT\n%!";
         eprintf "Model %d:\n%s\n%!" (i+1) (print_model imap model);
         aux (i+1) (next_instance dm model) in
  aux 0 dimacs

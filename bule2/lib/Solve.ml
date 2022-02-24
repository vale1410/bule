open Printf

type solver = CommandLine of string | Minisat | Quantor

module IntSet = Set.Make (Int)
module ModelSet = Set.Make (struct type t = (bool * int) list let compare = compare end)

let compare_literals (px, x) (py, y) = -(compare (x, px) (y, py))

module CL = struct
  let parse_status line =
    Scanf.sscanf line "s cnf %d %d %d" (fun p _ _ -> p = 1)

  let abs x = assert (x <> 0); if x < 0 then (false, -x) else (true, x)
  let parse_line line =
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
  
  let isnt_comment line = 
    String.length line > 0 && line.[0] <> 'c'

  let run_process cmd dimacs =
    let inp, out = Unix.open_process cmd in
    fprintf out "%s%!" (Dimacs.Print.qbf_file dimacs);
    close_out out;
    let lines = input_lines inp in
    close_in inp;
    let lines = List.filter isnt_comment lines in 
    parse_output lines

  let run_solver keys dimacs cmd =
    match run_process cmd dimacs with
    | None -> None
    | Some model ->
       let assigned = List.fold_left (fun accu (_, x) -> IntSet.add x accu) IntSet.empty model in
       let missing = IntSet.diff keys assigned in
       let actual_model = IntSet.fold (fun x l -> (false, x) :: l) missing model in
       let sorted_model = List.sort compare_literals actual_model in
       Some sorted_model
end

module MS = struct
  let literal (sign, var) =
    (if sign then Fun.id else Minisat.Lit.neg) (Minisat.Lit.make var)
  let clause lits = List.map literal lits
  let clauses list =
    let solver = Minisat.create () in
    try List.iter (fun c -> Minisat.add_clause_l solver (clause c)) list;
        Minisat.solve solver;
        Some solver
    with Minisat.Unsat -> None

  let extract_model keys solver =
    let extract_var var =
      let lit = Minisat.Lit.make var in
      let polarity = match Minisat.value solver lit with | Minisat.V_true -> true | Minisat.V_false -> false | Minisat.V_undef -> false in
      (polarity, var) in
    let model = IntSet.fold (fun var l -> extract_var var :: l) keys [] in
    List.sort compare_literals model

  let run_solver keys (_, _, blocks, cls) =
    match blocks with
    | [] -> failwith "no blocks"
    | _ :: _ :: _ | (false, _) :: _ -> failwith (sprintf "Minisat cannot handle the quantifier structures %s." (Print.list Dimacs.Print.quantifier_block blocks))
    | (true, _) :: [] ->
       let solver = clauses cls in
       Option.map (extract_model keys) solver
end

module QT = struct

  let literal (p, i) = if p then Qbf.Lit.make i else Qbf.Lit.neg (Qbf.Lit.make i)
  let clause lits = List.map literal lits
  let qcnf (_, _, prefix, matrix) =
    let matrix = Qbf.QCNF.prop (List.map clause matrix) in
    let aux (q, block) f =
      let block = List.map Qbf.Lit.make block in
      if q then Qbf.QCNF.exists block f else Qbf.QCNF.forall block f in
    List.fold_right aux prefix matrix

  let assignment a var =
    let l = Qbf.Lit.make var in
    match a l with
    | Qbf.True -> (true, var)
    | Qbf.False -> (false, var)
    | Qbf.Undef -> (false, var)

  let extract_model = function
    | Qbf.Unsat -> None
    | Qbf.Timeout  -> failwith "timeout in qbf solver"
    | Qbf.Spaceout -> failwith "spaceout in qbf solver"
    | Qbf.Unknown  -> failwith "unknown error in qbf solver"
    | Qbf.Sat a -> Some a

  let run_solver keys (file : Dimacs.T.file) =
    let f = qcnf file in
    let r = Qbf.solve ~solver:Quantor.solver f in
    let m = extract_model r in
    match m with
    | Some a ->
       let model = IntSet.fold (fun var l -> assignment a var :: l) keys [] in
       Some (List.sort compare_literals model)
    | None -> None

end

let run_solver keys dimacs = function
  | CommandLine cmd -> CL.run_solver keys dimacs cmd
  | Minisat -> MS.run_solver keys dimacs
  | Quantor -> QT.run_solver keys dimacs

let print_literal imap (pol, var) =
  let tilde = if pol then " " else "~" in
  let sv = match Dimacs.T.IMap.find_opt var imap with None -> assert false | Some sv -> sv in
    sprintf "%s%s" tilde (Circuit.Print.search_var sv)

let filtered_model show_default model hide show =
  let printed_lit (px, x) =
    let l = if px then x else -x in
    let hidden = Dimacs.T.ISet.mem l hide
    and force_shown = Dimacs.T.ISet.mem l show in
    (show_default && not hidden) || (not show_default && force_shown) in
  List.filter printed_lit model
let print_one_model imap = Print.unlines (print_literal imap)
let print_all_models imap model = Print.unspaces Fun.id (List.sort compare (List.map (print_literal imap) model))

let map_keys map =
  let add k _ set = IntSet.add k set in
  Dimacs.T.IMap.fold add map IntSet.empty
let solve_one cmd show_default (dimacs, _, imap, hide, show) =
  let keys = map_keys imap in
  match CL.run_solver keys dimacs cmd with
  | None -> printf "UNSAT\n"
  | Some model ->
     printf "SAT\n";
     let fmodel = filtered_model show_default model hide show in
     eprintf "%s\n" (print_one_model imap fmodel)

let next_instance (nbvar, nbcls, blocks, cls) model =
  let flip_literal (pol, var) = (not pol, var) in
  let nmodel = List.map flip_literal model in
  (nbvar, nbcls + 1, blocks, nmodel :: cls)

let solve_all (cmd, show_default, bound) (dimacs, _, imap, hide, show) =
  eprintf "Instance ground. Starts solving\n%!";
  let keys = map_keys imap in
  let rec aux models displayed iteration dm =
    if iteration >= bound && bound > 0 then ()
    else
      match run_solver keys dm cmd with
      | None ->
         if iteration = 0 then printf "UNSAT\n%!";
         eprintf "No more models. Total: %d displayed models out of %d models.\n%!" displayed iteration
      | Some model ->
         if iteration = 0 then printf "SAT\n%!";
         let next = next_instance dm model in
         let fmodel = filtered_model show_default model hide show in
         if ModelSet.mem fmodel models then aux models displayed (iteration+1) next
         else
           (eprintf "Model %d: %s\n%!" (iteration+1) (print_all_models imap fmodel);
            aux (ModelSet.add fmodel models) (displayed+1) (iteration+1) next) in
  aux ModelSet.empty 0 0 dimacs

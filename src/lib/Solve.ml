open Printf

type solver = CommandLine of string | Minisat | Quantor

module IntSet = Set.Make (Int)
module ModelSet = Set.Make (struct type t = (bool * int) list let compare = compare end)

let debug = false
let deprintf fmt = if debug then eprintf fmt else ifprintf stderr fmt

let compare_literals (px, x) (py, y) = -(compare (x, px) (y, py))

module CL = struct
  let abs x = assert (x <> 0); if x < 0 then (false, -x) else (true, x)

  let parse_s_status line =
    Scanf.sscanf line "s %s" (fun p -> p = "SATISFIABLE")

  let parse_s_line line =
    let tokens = Str.split (Str.regexp "[ ]+") line in
    match tokens with
    | [] -> assert false
    | first :: rest ->
       assert (first = "v");
       let ints = List.rev_map int_of_string rest in
       match ints with
       | 0 :: mid -> List.rev_map abs mid
       | _ -> assert false

  let parse_s_output = function
    | [] -> eprintf "Error: no output.\n"; assert false
    | status :: rest ->
       let sat = parse_s_status status in
       if sat then
         match rest with [] | _ :: _ :: _ -> assert false | assign :: [] -> Some (parse_s_line assign)
       else None

  let parse_q_status line =
    Scanf.sscanf line "s cnf %d %d %d" (fun p _ _ -> p = 1)

  let parse_q_line line =
    Scanf.sscanf line "V %d 0" abs

  let parse_q_output = function
    | [] -> eprintf "Error: no output.\n"; assert false
    | status :: rest ->
       let sat = parse_q_status status in
       if sat then Some (List.map parse_q_line rest) else None

  let input_lines inp =
    let l = ref [] in
    (try
       while true do l := input_line inp :: !l done
     with End_of_file -> ());
    List.rev !l

  let isnt_comment line =
    String.length line > 0 && line.[0] <> 'c'

  let run_process (cmd, use_dimacs) dimacs =
    let inp, out = Unix.open_process cmd in
    if use_dimacs then fprintf out "%s%!" (Dimacs.Print.sat_file dimacs)
    else               fprintf out "%s%!" (Dimacs.Print.qbf_file dimacs);
    close_out out;
    let lines = input_lines inp in
    close_in inp;
    let lines = List.filter isnt_comment lines in
    if use_dimacs then parse_s_output lines
    else parse_q_output lines

  let run_solver _keys dimacs cmd =
    match run_process cmd dimacs with
    | None -> None
    | Some model ->
       (*let assigned = List.fold_left (fun accu (_, x) -> IntSet.add x accu) IntSet.empty model in
       let missing = IntSet.diff _keys assigned in
       let actual_model = IntSet.fold (fun x l -> (false, x) :: l) missing model in
       deprintf "model: %s\n" (Print.list (Print.couple Print.bool Print.int) model);
       deprintf "actual model: %s\n" (Print.list (Print.couple Print.bool Print.int) actual_model);
       let sorted_model = List.sort compare_literals actual_model in*)
       let sorted_model = List.sort compare_literals model in
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
    let cnf = List.map clause matrix in
    let matrix = Qbf.QCNF.prop cnf in
    let aux (q, block) f =
      let block = List.map Qbf.Lit.make block in
      if q then Qbf.QCNF.exists block f else Qbf.QCNF.forall block f in
    List.fold_right aux prefix matrix

  let assignment a var =
    let l = Qbf.Lit.make var in
    match a l with
    | Qbf.True  -> deprintf "true %d " var; Some (true,  var)
    | Qbf.False -> deprintf "fals %d " var; Some (false, var)
    | Qbf.Undef -> deprintf "unde %d " var; Some (false, var)

  let extract_model keys = function
    | Qbf.Unsat -> None
    | Qbf.Timeout  -> failwith "timeout in qbf solver"
    | Qbf.Spaceout -> failwith "spaceout in qbf solver"
    | Qbf.Unknown  -> failwith "unknown error in qbf solver"
    | Qbf.Sat a ->
       let add var l = match assignment a var with
         | None -> l
         | Some ass -> ass :: l in
       let model = IntSet.fold add keys [] in
       Some (List.sort compare_literals model)

  let first_block (_,_,blocks,_) = match blocks with
    | [] -> failwith "no blocks"
    | (true, vars) :: _ -> vars
    | (false, _) :: _ -> [] (*failwith "can't start with univ block"*)
  let run_solver keys (file : Dimacs.T.file) =
    let f = qcnf file in
    deprintf "file0: %s\n" (Dimacs.Print.qbf_file file);
    deprintf "file:%!"; if debug then (Qbf.QCNF.print Format.err_formatter f; Format.print_flush ());
    let relevant_vars = first_block file in
    if false then eprintf "\n\n\n\nrelevant: %s\n%!" (Print.unspaces Print.int relevant_vars);
    let r = Qbf.solve ~solver:Quantor.solver f in
    if false then (Qbf.pp_result Format.err_formatter r; Format.print_flush ();
                   Format.pp_print_flush Format.err_formatter ());
    deprintf "res\n%!";
    let keys = IntSet.filter (fun v -> List.mem v relevant_vars) keys in
    let model = extract_model keys r in
    model

end

let run_solver keys dimacs = function
  | (CommandLine cmd, use_dimacs) -> CL.run_solver keys dimacs (cmd, use_dimacs)
  | (Minisat, _) -> MS.run_solver keys dimacs
  | (Quantor, _) -> QT.run_solver keys dimacs

let print_literal imap (pol, var) =
  let tilde = if pol then " " else "~" in
  let sv = match Dimacs.T.IMap.find_opt var imap with None -> assert false | Some sv -> sv in
    sprintf "%s%s" tilde (Circuit.Print.search_var sv)

let filtered_model prefix model show =
  let (q, varsq) = match prefix with | s :: _ -> s | [] -> failwith "Nothing to solve" in
  let surface_vars = if q then varsq else [] in
  let printed_lit (px, x) =
    let l = if px then x else -x in
    let shown = Dimacs.T.ISet.mem l show
    and surface = List.mem x surface_vars in
    surface && shown in
  List.filter printed_lit model
let print_one_model imap = Print.unlines (print_literal imap)
let print_all_models imap model = Print.unspaces Fun.id (List.sort compare (List.map (print_literal imap) model))

let map_keys map =
  let add k _ set = IntSet.add k set in
  Dimacs.T.IMap.fold add map IntSet.empty
let solve_one cmd (dimacs, _, imap, show) =
  let (_, _, prefix, _) = dimacs in
  let keys = map_keys imap in
  match run_solver keys dimacs cmd with
  | None -> printf "UNSAT\n";
  | Some model -> printf "SAT\n";
                  let fmodel = filtered_model prefix model show in
                  eprintf "%s\n" (print_one_model imap fmodel)

let next_instance (nbvar, nbcls, blocks, cls) model =
  let flip_literal (pol, var) = (not pol, var) in
  let nmodel = List.map flip_literal model in
  (nbvar, nbcls + 1, blocks, nmodel :: cls)

let solve_all (cmd, bound) (dimacs, _, imap, show) =
  eprintf "Instance ground. Starts solving\n%!";
  let (_, _, prefix, _) = dimacs in
  let keys = map_keys imap in
  let rec aux models displayed iteration dm =
    if iteration >= bound && bound > 0 then eprintf "Total: %d displayed models out of at least %d models.\n%!" displayed iteration
    else
      match run_solver keys dm cmd with
      | None ->
         if iteration = 0 then printf "UNSAT\n%!";
         eprintf "No more models. Total: %d displayed models out of %d models.\n%!" displayed iteration
      | Some model ->
         if iteration = 0 then printf "SAT\n%!";
         let next = next_instance dm model in
         let fmodel = filtered_model prefix model show in
         if ModelSet.mem fmodel models then aux models displayed (iteration+1) next
         else
           (eprintf "Model %d: %s\n%!" (iteration+1) (print_all_models imap fmodel);
            aux (ModelSet.add fmodel models) (displayed+1) (iteration+1) next) in
  aux ModelSet.empty 0 0 dimacs

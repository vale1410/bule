open Printf

type solver = CommandLine of string | Minisat | Quantor

module IntSet = Set.Make (Int)
module ModelSet = Set.Make (struct type t = (bool * int) list let compare = compare end)
let print_model_set m = Print.list (Print.unspaces Dimacs.Print.literal) (ModelSet.elements m)

let debug = false
let deprintf fmt = if debug then eprintf fmt else ifprintf stderr fmt
let sscanf_opt string arg x =
  try Some (Scanf.sscanf string arg x) with
  | Scanf.Scan_failure _ -> None
let sscanf_int_opt string arg x =
  try Some (Scanf.sscanf string arg x) with
  | Scanf.Scan_failure _ -> None

let compare_literals (px, x) (py, y) = -(compare (x, px) (y, py))

module CL = struct
  let abs x =
    if      x = 0 then None
    else if x < 0 then Some (false, -x)
    else               Some (true, x)

  let parse_s_solution line =
    match (sscanf_opt line "%s") (fun p -> p = "SATISFIABLE") with
    | Some status -> status
    | None -> eprintf "I could not interpret the solver solution line. I read '%s', but I expected 's %%s'" line; assert false

  let parse_q_solution line =
    match sscanf_opt line "cnf %d %d %d" (fun p _ _ -> p = 1) with
    | Some status -> status
    | None -> eprintf "I could not interpret the solver solution line. I read '%s', but I expected 's cnf %%d %%d %%d'\n" line; assert false

  let input_lines inp =
    let l = ref [] in
    (try
       while true do l := input_line inp :: !l done
     with End_of_file -> ());
    List.rev !l

  let loose_parse_line line =
    let tokens = Str.split (Str.regexp "[ ]+") line in
    let lit_of_string s = Option.bind (int_of_string_opt s) abs in
    List.filter_map lit_of_string tokens

  let loose_parse_file lines =
    let one_line line =
      let head, rest = if String.length line <= 1 then ("c ", "") else (String.sub line 0 2, String.sub line 2 (String.length line - 2)) in
      match head with
      | "c " | "C " -> None
      | "s " | "S " -> Some (Either.Left rest)
      | "v " | "V " -> Some (Either.Right rest)
      | _ -> eprintf "Ignoring this line in the solver output '%s'\n" line; None in
    let meaningful_lines = List.filter_map one_line lines in
    let solutions, values = List.partition_map Fun.id meaningful_lines in
    match solutions with
    | [] -> None
    | _ :: _ :: _ -> eprintf "Too many solution lines in the solver output: %s\n" (Print.list Fun.id solutions); assert false
    | solution :: [] ->
      let values = List.concat_map loose_parse_line values in
      let values = if values <> [] then Some values else None in
      Some (solution, values)

  let run_process cmd question =
    let stdin_name  = Filename.temp_file "bule." ".in" in
    let stdout_name = Filename.temp_file "bule." ".out" in
    let stderr_name = Filename.temp_file "bule." ".err" in
    let (cmd, args) = match Str.split (Str.regexp "[ ]+") cmd with
      | [] -> assert false
      | hd :: tl -> (hd, tl) in
    let cmd = Filename.quote_command cmd ~stdin:stdin_name ~stdout:stdout_name ~stderr:stderr_name args in
    Print.to_file stdin_name question;
    (*let stdin_f = open_out stdin_name in
    fprintf stdin_f "%s%!" question;
    close_out stdin_f;*)
    deprintf "cmd: %s\n%!" cmd;
    let status = Unix.system cmd in
    let stdout_f = open_in stdout_name
    and stderr_f = open_in stderr_name in
    let out_lines = input_lines stdout_f
    and err_lines = input_lines stderr_f in
    close_in stdout_f;
    close_in stderr_f;
    List.iter Sys.remove [stdin_name; stdout_name; stderr_name];
    let answer = match status with
      | Unix.WEXITED 10 -> Some true
      | Unix.WEXITED 20 -> Some false
      | Unix.WEXITED code -> eprintf "Unknown solver exit code: %d\n%!" code; None
      | _ -> assert false in
    (*eprintf "Answer:\n%s\n" (Print.list Fun.id lines);*)
    (answer, out_lines, err_lines)

  let run_solver _keys formula (cmd, use_dimacs) =
    let input = if use_dimacs then Dimacs.Print.sat_file formula else Dimacs.Print.qbf_file formula in
    let (answer, out_lines, err_lines) = run_process cmd input in
    deprintf "%s\n%!" (Print.unlines Fun.id err_lines);
    let certificate = loose_parse_file out_lines in
    match answer, certificate with
    | None, None -> None
    | None, Some _ -> eprintf "Solver outputs a certificate but its output code indicates it was unable to solve the instance.\n%!"; assert false
    | Some answer, None -> Some (answer, None)
    | Some sol1, Some (sol2, values) ->
       let sol2 =  if use_dimacs then parse_s_solution sol2 else parse_q_solution sol2 in
       if sol1 <> sol2 then (eprintf "Solver certificate inconsistent with its output code.\n%!"; assert false);
       Some (sol1, values)
end

module MS = struct
  let extract_model keys solver =
    let extract_var var =
      let lit = Minisat.Lit.make var in
      let polarity = match Minisat.value solver lit with | Minisat.V_true -> true | Minisat.V_false -> false | Minisat.V_undef -> false in
      (polarity, var) in
    let model = IntSet.fold (fun var l -> extract_var var :: l) keys [] in
    model

  let literal (sign, var) =
    (if sign then Fun.id else Minisat.Lit.neg) (Minisat.Lit.make var)
  let clause lits = List.map literal lits
  let clauses keys list =
    let solver = Minisat.create () in
    try List.iter (fun c -> Minisat.add_clause_l solver (clause c)) list;
        Minisat.solve solver;
        let solution = extract_model keys solver in
        Some (true, Some solution)
    with Minisat.Unsat -> Some (false, None)

  let run_solver keys (_, _, blocks, cls) =
    match blocks with
    | [] -> failwith "no blocks"
    | _ :: _ :: _ | (false, _) :: _ -> failwith (sprintf "Minisat cannot handle the quantifier structures %s." (Print.list Dimacs.Print.quantifier_block blocks))
    | (true, _) :: [] ->
       clauses keys cls
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
    | Qbf.Unsat -> false, None
    | Qbf.Timeout  -> failwith "timeout in qbf solver"
    | Qbf.Spaceout -> failwith "spaceout in qbf solver"
    | Qbf.Unknown  -> failwith "unknown error in qbf solver"
    | Qbf.Sat a ->
       let add var l = match assignment a var with
         | None -> l
         | Some ass -> ass :: l in
       let model = IntSet.fold add keys [] in
       true, Some model

  let first_block (_,_,blocks,_) = match blocks with
    | [] -> failwith "no blocks"
    | (_, vars) :: _ -> vars
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
    Some model

end

let run_solver keys dimacs solver =
  let solver_output = match solver with
    | (CommandLine cmd, use_dimacs) -> CL.run_solver keys dimacs (cmd, use_dimacs)
    | (Minisat, _) -> MS.run_solver keys dimacs
    | (Quantor, _) -> QT.run_solver keys dimacs in
  match solver_output with
  | None -> None
  | Some (solution, values) ->
     let sorted_values = Option.map (List.sort compare_literals) values in
     Some (solution, sorted_values)

let print_variable imap var =
  let sv = match Dimacs.T.IMap.find_opt var imap with None -> assert false | Some sv -> sv in
  Circuit.Print.search_var sv
let print_literal imap (pol, var) =
  let sv = match Dimacs.T.IMap.find_opt var imap with None -> assert false | Some sv -> sv in
  Circuit.Print.literal (pol, sv)

let warn_bad_values imap surface_vars values =
  let non_surface_assigned = List.filter (fun lit  -> not (List.mem (snd lit) surface_vars)) values in
  if non_surface_assigned <> []
  then eprintf "Bug: the solver has assigned variables that are not in the outer quantifier block: %s\n" (Print.unspaces (print_literal imap) non_surface_assigned)

let filtered_model surface_vars show values =
  let is_external         var  = Dimacs.T.ISet.mem var show || Dimacs.T.ISet.mem (-var) show in
  let is_external_lit (_, var) = Dimacs.T.ISet.mem var show || Dimacs.T.ISet.mem (-var) show in
  let is_flexible var = not (List.mem (true, var) values) && not (List.mem (false, var) values) in
  let is_shown (pol, var) = Dimacs.T.ISet.mem (if pol then var else -var) show in
  let assign_arbitrary var = match is_shown (true, var), is_shown (false, var) with
    | true,  true  -> Either.Left (true, var)
    | true,  false -> Either.Right (false, var)
    | false, true  -> Either.Right (true, var)
    | false, false -> Either.Right (true, var) in
  let flexible_vars = List.filter is_flexible surface_vars in
  let arbitrary_vars, free_vars = List.partition is_external flexible_vars in
  let external_lits, internal_lits = List.partition is_external_lit values in
  let show_forced, omit_forced = List.partition is_shown external_lits in
  let show_arbitrary, omit_arbitrary = List.partition_map assign_arbitrary arbitrary_vars in
  let show_lits = Misc.flatten [show_forced; show_arbitrary]
  and omit_lits = Misc.flatten [omit_forced; omit_arbitrary] in
  (show_lits, omit_lits, internal_lits, free_vars)

let filtered_model2 surface_vars show values =
  let is_shown (pol, var) = Dimacs.T.ISet.mem (if pol then var else -var) show in
  let is_committed var = is_shown (true, var) || is_shown (false, var) in
  let has_fixed_value var =
    let plit, nlit = (true, var), (false, var) in
    if List.mem plit values then Either.Left plit
    else if List.mem nlit values then Either.Left nlit
    else Either.Right var in
  let fixed, flexible = List.partition_map has_fixed_value surface_vars in
  let show_fixed, hide_fixed = List.partition is_shown fixed in
  let committed, free = List.partition is_committed flexible in
  let show_committed, hide_committed = List.partition is_shown (List.map (fun var -> (true, var)) committed) in
  (*eprintf "show %s\n%!" (Print.list Print.int (IntSet.elements show));
  eprintf "show fixed %s, hide fixed %s\n%!" (Print.list (Print.couple Print.bool Print.int) show_fixed) (Print.list (Print.couple Print.bool Print.int) hide_fixed);
  eprintf "committed %s, free %s\n%!" (Print.list Print.int committed) (Print.list Print.int free);
  eprintf "show committed %s, hide committed %s\n%!" (Print.list (Print.couple Print.bool Print.int) show_committed) (Print.list (Print.couple Print.bool Print.int) hide_committed);*)
  let show_all = List.rev_append show_committed show_fixed
  and hide_all = List.rev_append hide_committed hide_fixed in
  (show_all, hide_all, [], free)

let print_one_model imap = Print.unlines (print_literal imap)
let print_enumerate_model imap model = Print.unspaces Fun.id (List.sort compare (List.map (print_literal imap) model))

let map_keys map =
  let add k _ set = IntSet.add k set in
  Dimacs.T.IMap.fold add map IntSet.empty

let next_sat_instance (nbvar, nbcls, blocks, cls) model =
  let flip_literal (pol, var) = (not pol, var) in
  let nmodel = List.map flip_literal model in
  (nbvar, nbcls + 1, blocks, nmodel :: cls)

let naive_next_taut_instance (nbvar, nbcls, blocks, cls) counterexample =
  let weaken_clause clause = Misc.map (fun lit -> lit :: clause) counterexample in
  let wcls = List.concat_map weaken_clause cls in
  (nbvar, nbcls * List.length counterexample, blocks, wcls)

let extract_objective (_, _, prefix, _) = match prefix with
  | [] -> None
  | head :: _ -> Some head

let vocabulary = function
  | true  -> "SAT", "UNSAT", "model"
  | false -> "VALID", "INVALID", "counterexample"

let solve_one cmd (dimacs, _, imap, show) =
  match extract_objective dimacs with
  | None -> eprintf "Trivial formula, nothing to solve.\n%!"
  | Some (sat, surface) ->
  let success, failure, _ = vocabulary sat in
  let keys = map_keys imap in
  match run_solver keys dimacs cmd with
  | None -> eprintf "Couldn't solve the instance.\n%!"
  | Some (true,  None) -> printf "%s\n" success
  | Some (false, None) -> printf "%s\n" failure
  | Some (solution, Some model) ->
     if solution then printf "%s\n" success else printf "%s\n" failure;
     warn_bad_values imap surface model;
     let (shown, _, _, _) = filtered_model surface show model in
     eprintf "%s\n" (print_one_model imap shown)

type stats =
  { displayed : int;
    computed : int;
    inferred : int; }

let update_solutions sat stats shown_solutions formula (show, omit, hide, free) =
  let next_instance = if sat then next_sat_instance else naive_next_taut_instance in
  let relevant = List.sort compare (Misc.flatten [show; omit]) in
(*  let solution = Misc.flatten [show; omit; hide; Misc.map (fun var -> (false, var)) free] in*)
  let solution = Misc.flatten [show; omit; hide] in
  let next_formula = next_instance formula solution in
  let already_displayed = ModelSet.mem relevant shown_solutions in
  deprintf "shown: %s,\nshow: %s, omit: %s, hide: %s, free: %s\n%!" (print_model_set shown_solutions) (Print.unspaces Dimacs.Print.literal show) (Print.unspaces Dimacs.Print.literal omit) (Print.unspaces Dimacs.Print.literal hide) (Print.unspaces Print.int free);
  (*eprintf "Already %B, Next formula:\n%s\n" already_displayed (Dimacs.Print.qbf_file next_formula);*)
  let stats = { displayed = if already_displayed then stats.displayed else stats.displayed + 1;
                computed = stats.computed + 1;
                inferred = stats.inferred + Misc.pow 2 (List.length free) - 1} in
  let shown_solutions = if already_displayed then shown_solutions else ModelSet.add relevant shown_solutions in
  (stats, shown_solutions, next_formula)

let solve_all (cmd, bound) (dimacs, _, imap, show) =
  eprintf "Instance ground. Starts solving\n%!";
  deprintf "show %s\n%!" (Print.list Print.int (IntSet.elements show));
  match extract_objective dimacs with
  | None -> eprintf "Trivial formula, nothing to solve.\n%!"
  | Some (sat, surface) ->
  let success, failure, values_name = vocabulary sat in
  let keys = map_keys imap in
  let init_stats = { displayed = 0; computed = 0; inferred = 0 } in
  let rec aux stats models dm =
    (*eprintf "Current formula:\n%s\n" (Dimacs.Print.qbf_file dm);*)
    let iteration = stats.computed in
    let total = stats.computed + stats.inferred in
    let print_result answer = if stats.computed = 0 then if answer then printf "%s\n%!" success else printf "%s\n%!" failure in
    if iteration >= bound && bound > 0 then eprintf "Total: %d displayed %ss out of at least %d %ss. Computed: %d, inferred: %d.\n%!" stats.displayed values_name total values_name stats.computed stats.inferred
    else
      match run_solver keys dm cmd with
      | None -> ()
      | Some (answer, None) ->
         print_result answer;
         if answer = sat then eprintf "No more %ss identified. %d %ss displayed. Computed: %d, inferred: %d.\n%!" values_name stats.displayed values_name stats.computed stats.inferred
         else eprintf "All %d %ss identified, %d displayed. Computed: %d, inferred: %d.\n%!" total values_name stats.displayed stats.computed stats.inferred
      | Some (answer, Some values) ->
         print_result answer;
         warn_bad_values imap surface values;
         let (shown, _omit, _hidden, _free) as fmodel = filtered_model surface show values in
         (*eprintf "Output: %s, %s, %s\n%!" (Print.list (print_literal imap) shown) (Print.list (print_literal imap) _hidden) (Print.list (print_variable imap) _free);*)
         let next_stats, next_shown_solutions, next_formula = update_solutions sat stats models dm fmodel in
         (*eprintf "Found values %s\nNext formula:\n%s\n" (print_enumerate_model imap values) (Dimacs.Print.qbf_file next_formula);*)
         if next_stats.displayed <> stats.displayed then eprintf "%s %d: %s\n%!" (String.capitalize_ascii values_name) (next_stats.displayed) (print_enumerate_model imap shown);
         aux next_stats next_shown_solutions next_formula in
  aux init_stats ModelSet.empty dimacs

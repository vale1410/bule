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

let compare_literals (px, x) (py, y) = compare (x, px) (y, py)

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

  let loose_parse_line line =
    let tokens = Str.split (Str.regexp "[ ]+") line in
    let lit_of_string s = Option.bind (int_of_string_opt s) abs in
    List.filter_map lit_of_string tokens

  let loose_parse_file lines =
    let one_line line =
      let head, rest = if String.length line <= 1 then ("c ", "") else (String.sub line 0 2, String.sub line 2 (String.length line - 2)) in
      match head with
      | "c " | "C " -> Some (Either.Left rest)
      | "s " | "S " -> Some (Either.Right (Either.Left rest))
      | "v " | "V " -> Some (Either.Right (Either.Right rest))
      | _ -> eprintf "Ignoring this line in the solver output '%s'\n" line; None in
    let meaningful_lines = List.filter_map one_line lines in
    let comments, answers = List.partition_map Fun.id meaningful_lines in
    let solutions, values = List.partition_map Fun.id answers in
    match solutions with
    | [] -> (comments, None)
    | _ :: _ :: _ -> eprintf "Too many solution lines in the solver output: %s\n" (Print.list Fun.id solutions); assert false
    | solution :: [] ->
      let values = List.concat_map loose_parse_line values in
      let values = if values <> [] then Some values else None in
      (comments, Some (solution, values))

  let run_process cmd question =
    deprintf "cmd: %s\n%!" cmd;
    let (status, out_lines, err_lines) = Misc.run_process cmd question in
    let answer = match status with
      | Either.Left 10 -> Some true
      | Either.Left 20 -> Some false
      | Either.Left code -> eprintf "Unknown solver exit code: %d\n%!" code; None
      | Either.Right _ -> assert false in
    (answer, out_lines, err_lines)

  let run_solver formula (cmd, (display_comments, use_dimacs)) =
    let input = if use_dimacs then Dimacs.Print.sat_file formula else Dimacs.Print.qbf_file formula in
    let (answer, out_lines, err_lines) = run_process cmd input in
    deprintf "%s\n%!" (Print.unlines Fun.id err_lines);
    let comments, certificate = loose_parse_file out_lines in
    if display_comments then List.iter (eprintf "c %s\n") comments;
    match answer, certificate with
    | None, None -> None
    | None, Some _ -> eprintf "Solver outputs a certificate but its output code indicates it was unable to solve the instance.\n%!"; assert false
    | Some answer, None -> Some (answer, None)
    | Some sol1, Some (sol2, values) ->
       let sol2 = if use_dimacs then parse_s_solution sol2 else parse_q_solution sol2 in
       if sol1 <> sol2 then (eprintf "Solver certificate inconsistent with its output code.\n%!"; assert false);
       Some (sol1, values)
end

module MS = struct
  let extract_model keys solver =
    let extract_var var =
      let lit = Minisat.Lit.make var in
      match Minisat.value solver lit with | Minisat.V_true -> Some (true, var) | Minisat.V_false -> Some (false, var) | Minisat.V_undef -> None in
    let model = List.filter_map extract_var keys in
    model

  let literal (sign, var) =
    (if sign then Fun.id else Minisat.Lit.neg) (Minisat.Lit.make var)
  let clause lits = List.map literal lits
  let clauses vars list =
    let solver = Minisat.create () in
    try List.iter (fun c -> Minisat.add_clause_l solver (clause c)) list;
        Minisat.solve solver;
        let solution = extract_model vars solver in
        Some (true, Some solution)
    with Minisat.Unsat -> Some (false, None)

  let run_solver (sat, vars) (_, _, blocks, cls) =
    if not sat || List.length blocks > 1 then failwith (sprintf "Minisat cannot handle the quantifier structures %s." (Print.list Dimacs.Print.quantifier_block blocks));
    clauses vars cls
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
    | Qbf.Undef -> deprintf "unde %d " var; None

  let extract_model keys = function
    | Qbf.Unsat -> None
    | Qbf.Timeout  -> failwith "timeout in qbf solver"
    | Qbf.Spaceout -> failwith "spaceout in qbf solver"
    | Qbf.Unknown  -> failwith "unknown error in qbf solver"
    | Qbf.Sat a ->
       let model = List.filter_map (assignment a) keys in
       deprintf "\nkeys: %s\n%!" (Print.unspaces Print.int keys);
       Some model

  let run_solver (_, surface_vars) (file : Dimacs.T.file) =
    let f = qcnf file in
    eprintf "file0: %s\n" (Dimacs.Print.qbf_file file);
    deprintf "file:%!"; if debug then (Qbf.QCNF.print Format.err_formatter f; Format.print_flush ());
    if false then eprintf "\n\n\n\nrelevant: %s\n%!" (Print.unspaces Print.int surface_vars);
    let r = Qbf.solve ~solver:Quantor.solver f in
    if false then (Qbf.pp_result Format.err_formatter r; Format.print_flush ();
                   Format.pp_print_flush Format.err_formatter ());
    deprintf "res\n%!";
    match extract_model surface_vars r with
    | None -> Some (false, None)
    | Some values ->
    Some (true, Some values)
(*    match Preprocessing.simplify values file with
    | None -> assert false
    | Some (model, _) ->
     eprintf "QT: return_model %s, UP %s\n%!" (Print.unspaces Dimacs.Print.literal values) (Print.unspaces Dimacs.Print.literal model);
     Some (true, Some model)
*)

end

let run_solver objective dimacs solver =
  let solver_output = match solver with
    | (CommandLine cmd, use_dimacs) -> CL.run_solver dimacs (cmd, use_dimacs)
    | (Minisat, _) -> MS.run_solver objective dimacs
    | (Quantor, _) -> QT.run_solver objective dimacs in
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
    | true,  true  -> Either.Left  (true,  var)
    | true,  false -> Either.Right (false, var)
    | false, true  -> Either.Right (true,  var)
    | false, false -> Either.Right (true,  var) in
  let flexible_vars = List.filter is_flexible surface_vars in
  let arbitrary_vars, free_vars = List.partition is_external flexible_vars in
  let external_lits, internal_lits = List.partition is_external_lit values in
  let show_forced, omit_forced = List.partition is_shown external_lits in
  let show_arbitrary, omit_arbitrary = List.partition_map assign_arbitrary arbitrary_vars in
  let show_lits = Misc.flatten [show_forced; show_arbitrary]
  and omit_lits = Misc.flatten [omit_forced; omit_arbitrary] in
  (show_lits, omit_lits, internal_lits, free_vars)

let print_one_model imap values = Print.unlines (print_literal imap) (List.sort compare_literals values)
let print_enumerate_model imap model = Print.unspaces (print_literal imap) (List.sort compare_literals model)

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
  | true  -> "SAT", "UNSAT"
  | false -> "VALID", "INVALID"
let solution_name = function
  | true  -> "model"
  | false -> "counterexample"

type stats =
  { already_displayed : bool;
    displayed : int;
    computed : int;
    inferred : int; }
let init_stats = { already_displayed = false; displayed = 0; computed = 0; inferred = 0 }

let display_result sat answer stats =
  let success, failure = vocabulary sat in
  if stats.computed = 0 then if answer then printf "%s\n%!" success else printf "%s\n%!" failure
let display_all_solutions_found sat stats =
  let values_name = solution_name sat in
  eprintf "All %ss identified, %d displayed. Computed: %d, inferred: %d.\n%!" values_name stats.displayed stats.computed stats.inferred
let display_no_more_solutions_found sat stats =
  let values_name = solution_name sat in
  eprintf "No more %ss identified, %d displayed. Computed: %d, inferred: %d.\n%!" values_name stats.displayed stats.computed stats.inferred
let display_budget_over sat stats =
  let values_name = solution_name sat in
  let total = stats.computed + stats.inferred in
  eprintf "Total: %d displayed %ss out of at least %d %ss. Computed: %d, inferred: %d.\n%!" stats.displayed values_name total values_name stats.computed stats.inferred
let display_solution sat stats imap (shown, _, _, _) =
  let values_name = solution_name sat in
  let solution = print_enumerate_model imap shown in
  if not stats.already_displayed then eprintf "%s %d: %s\n%!" (String.capitalize_ascii values_name) (stats.displayed) solution

let solve_one cmd (dimacs, _, imap, show) =
  match extract_objective dimacs with
  | None -> eprintf "Trivial formula, nothing to solve.\n%!"
  | Some (sat, surface) ->
  match run_solver (sat, surface) dimacs cmd with
  | None -> eprintf "Couldn't solve the instance.\n%!"
  | Some (solution, certificate) ->
     display_result sat solution init_stats;
     match certificate with
  | None -> ()
  | Some values ->
     warn_bad_values imap surface values;
     let (shown, _, _, _) = filtered_model surface show values in
     eprintf "%s\n" (print_one_model imap shown)

let update_solutions sat stats shown_solutions formula (show, omit, hide, free) =
  let next_instance = if sat then next_sat_instance else naive_next_taut_instance in
  let relevant = List.sort compare (Misc.flatten [show; omit]) in
(*  let solution = Misc.flatten [show; omit; hide; Misc.map (fun var -> (false, var)) free] in*)
  let solution = Misc.flatten [show; omit; hide] in
  let formula = next_instance formula solution in
  let already_displayed = ModelSet.mem relevant shown_solutions in
  deprintf "shown: %s,\nshow: %s, omit: %s, hide: %s, free: %s\n%!" (print_model_set shown_solutions) (Print.unspaces Dimacs.Print.literal show) (Print.unspaces Dimacs.Print.literal omit) (Print.unspaces Dimacs.Print.literal hide) (Print.unspaces Print.int free);
  deprintf "sat %B; Already %B, Next formula:\n%s\n" sat already_displayed (Dimacs.Print.qbf_file formula);
  let stats = { already_displayed;
                displayed = if already_displayed then stats.displayed else stats.displayed + 1;
                computed = stats.computed + 1;
                inferred = stats.inferred + Misc.pow 2 (List.length free) - 1} in
  let shown_solutions = if already_displayed then shown_solutions else ModelSet.add relevant shown_solutions in
  (stats, shown_solutions, formula)

let solve_all (cmd, bound) (dimacs, _, imap, show) =
  eprintf "Instance ground. Starts solving\n%!";
  deprintf "show %s\n%!" (Print.list Print.int (IntSet.elements show));
  match extract_objective dimacs with
  | None -> eprintf "Trivial formula, nothing to solve.\n%!"
  | Some (sat, surface) ->
  let rec aux stats models dm =
    (*eprintf "Current formula:\n%s\n" (Dimacs.Print.qbf_file dm);*)
    match run_solver (sat, surface) dm cmd with
    | None -> ()
    | Some (answer, certificate) ->
       display_result sat answer stats;
       match certificate with
       | None ->
          if answer = sat then display_no_more_solutions_found sat stats
          else display_all_solutions_found sat stats
       | Some values ->
          warn_bad_values imap surface values;
          let fmodel = filtered_model surface show values in
          (*eprintf "Output: %s, %s, %s\n%!" (Print.list (print_literal imap) shown) (Print.list (print_literal imap) _hidden) (Print.list (print_variable imap) _free);*)
          let next_stats, next_shown_solutions, next_formula = update_solutions sat stats models dm fmodel in
          (*eprintf "Found values %s\nNext formula:\n%s\n" (print_enumerate_model imap values) (Dimacs.Print.qbf_file next_formula);*)
          display_solution sat next_stats imap fmodel;
          let budget_over = next_stats.computed >= bound && bound > 0 in
          if budget_over then display_budget_over sat next_stats
          else aux next_stats next_shown_solutions next_formula in
  aux init_stats ModelSet.empty dimacs

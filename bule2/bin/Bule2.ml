open Printf
open Bule2

let update reference value = reference := value
let get () =
(*  let debug = debug_false () in
  let verbose = ref 0 in
  let asp = ref false in
  let output_name = ref "" in
  let speclist =
    [("--verbose",        Arg.Set_int verbose, "Enable verbose mode.", sprintf "%d" !verbose);
     ("-o",               Arg.Set_string output_name, "Set the output path.", !output_name);
     ("--direction",      Arg.Symbol (directions, set_direction), "Set the proof direction.", sprintf "%s" (Opt.show_proof_direction !direction));
     ("--asp",            Arg.Bool (update asp), "Output ASP instead of FOL.", sprintf "%B" !asp);
    ] in*)
  let input_names = ref [] in
  let dimacs = ref false in
  let models = ref 1 in
  let solver = ref "none" in
  let speclist = [("--dimacs", Arg.Bool (update dimacs), "Output DIMACS format rather than BULE. Default false.");
                  ("-",        Arg.Unit (fun () -> update input_names ("-" :: !input_names)), "Read the BULE code from the standard input.");
                  ("--models", Arg.Set_int models, "Number of models to generate. Default: 1.");
                  ("--solver", Arg.Set_string solver, "Set the solver to be used. If \"none\" then no solving takes place. Example \"depqbf --no-dynamic-nenofex --qdo\". Default: \"none\"")
] in
(*  let speclist = [] in*)
  let usage_msg = "BULE Grounder. Options available:" in
  let add_name s = input_names := s :: !input_names in
  Arg.parse speclist add_name usage_msg;
  let files = match List.rev !input_names with
  | [] -> failwith "Wrong number of arguments. Usage: bule2 file"
  | _ :: _ as names -> names in
  let command = if !solver = "none" then None else Some !solver in
  (command, !dimacs, !models, files)


(*let convert g =
  let formula = Formula.file g in
(*  let qbf = QBF.form formula in
  let qcir = QBF.to_qcir qbf in*)
  let qcir = QBF.model_checking_empty formula in
  let qcir = QCIR.sanitize_names qcir in
  (*eprintf "%s\n" (Ast.Print.file p);*)
  (*eprintf "%s\n" (Ground.Print.file g);*)
  (*printf "%s\n" (Formula.Print.formula formula);*)
  (*printf "%s\n" (QBF.Print.formula qbf);*)
  printf "%s\n" (QCIR.Print.file qcir);
  ()*)
(*
let ground g =
  printf "%s\n\n" (Circuit.Print.file g);
  let f = Formula.file g in
  printf "%s\n" (Formula.Print.formula f)

let ground_d g =
  let f = Formula.file g in
  let _d = Desugar.formula f in
  (*printf "%s\n" (Desugar.Print.formula _d);*)
  ()
*)

(*let solve dim = Solve.solve "depqbf --no-dynamic-nenofex --qdo" dim*)
let solve models dim command = Solve.solve_all command models dim

let start () =
  let command, dimacs, models, fs = get () in
  let ps = List.map (Parse.from_file ()) fs in
  let p = List.flatten ps in
  (*printf "%s\n\n" (Ast.Print.file p);*)
  let g = Circuit.file p in
  let (d, vm, im, hs) = Dimacs.ground g in
  match command with
  | Some comm -> solve models (d, vm, im, hs) comm
  | None ->
     let output = if dimacs then Dimacs.Print.file d else Circuit.Print.file g in
     printf "%s\n" output


  (*if gr then ground g else solve g*)
  (*if gr then convert g else solve g*)

let _ = start ()


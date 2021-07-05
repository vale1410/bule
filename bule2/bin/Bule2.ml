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
  let ground = ref false in
  let speclist = [("--ground", Arg.Bool (update ground), "Ground rather than Solve.");
                  ("-",        Arg.Unit (fun () -> update input_names ("-" :: !input_names)), "Read the BULE code from the standard input.");
] in
(*  let speclist = [] in*)
  let usage_msg = "BULE Grounder. Options available:" in
  let add_name s = input_names := s :: !input_names in
  Arg.parse speclist add_name usage_msg;
  let files = List.rev !input_names in
  (*eprintf "files=%s\n%!" (Bule2.Print.list Print.string files);*)
  match files with
  | [] -> failwith "Wrong number of arguments. Usage: dlpag file"
  | _ :: _ -> (!ground, files)


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

let start () =
  let _gr, fs = get () in
  let ps = List.map (Parse.from_file ()) fs in
  let p = List.flatten ps in
  (*printf "%s\n\n" (Ast.Print.file p);*)
  let g = Circuit.file p in
  printf "%s\n\n" (Circuit.Print.file g)
  (*if gr then ground g else solve g*)
  (*if gr then convert g else solve g*)

let _ = start ()


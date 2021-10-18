open Printf
open Bule2

type output_format = Bule | Dimacs | Qdimacs

let read_output_format = function
  | "bule" -> Bule
  | "dimacs" -> Dimacs
  | "qdimacs" -> Qdimacs
  | _ -> assert false

let update reference value = reference := value
let get () =
(*  let debug = debug_false () in
  let output_name = ref "" in
  let speclist =
    [("--verbose",        Arg.Set_int verbose, "Enable verbose mode.", sprintf "%d" !verbose);
     ("-o",               Arg.Set_string output_name, "Set the output path.", !output_name);
     ("--direction",      Arg.Symbol (directions, set_direction), "Set the proof direction.", sprintf "%s" (Opt.show_proof_direction !direction));
     ("--asp",            Arg.Bool (update asp), "Output ASP instead of FOL.", sprintf "%B" !asp);
    ] in*)
  let input_names = ref [] in
  let default_show = ref true in
  let models = ref 1 in
  let solver = ref "none" in
  let solve = ref false in
  let facts = ref false in
  let print_mode = ref Bule in
  let output_symbols = ["qdimacs"; "dimacs"; "bule"]
  and output_treat s = print_mode := read_output_format s in
  let speclist = [("-",        Arg.Unit (fun () -> update input_names ("-" :: !input_names)), "Read the BULE code from the standard input.");
                  ("--solve",  Arg.Set solve, "Enable solving. Default: \"false\"");
                  ("--models", Arg.Set_int models, "Number of models to generate. The option has no effect if \"solve\" is set to \"false\". Default: 1.");
                  ("--solver", Arg.Set_string solver, "Set the solver to be used. If \"none\" then Minisat 1.14 is used. Example \"depqbf --no-dynamic-nenofex --qdo\". The option has no effect if \"solve\" is set to \"false\". Default: \"none\"");
                  ("--output", Arg.Symbol (output_symbols, output_treat), "Output format (QDIMACS, DIMACS, or BULE. The option has no effect if \"solve\" is set to \"true\". Default \"bule\".");
                  ("--facts",  Arg.Set facts, "Enable printing of grounding facts. The option has no effect if \"solve\" is set to \"true\". Default: \"false\".");
                  ("--default_show", Arg.Bool (update default_show), "Default showing behaviour for literals. The option has no effect if \"solve\" is set to \"false\". Default \"true\".");
                 ] in
  let usage_msg = "BULE Grounder. Options available:" in
  let add_name s = input_names := s :: !input_names in
  Arg.parse speclist add_name usage_msg;
  let files = match List.rev !input_names with
  | [] -> failwith "Wrong number of arguments. Usage: bule2 file"
  | _ :: _ as names -> names in
  let solver = if !solver = "none" then None else Some !solver in
  let mode = if !solve then Either.Right (solver, !default_show, !models) else Either.Left (!facts, !print_mode) in
  (mode, files)

let solve_mode file solve_options =
  let circuit = Circuit.file false file in
  let file = Dimacs.ground circuit in
  Solve.solve_all solve_options file
let ground_mode file (facts, format) =
  let circuit = Circuit.file facts file in
  let d = Dimacs.file circuit in
  let output = match format with
  | Bule -> Circuit.Print.file circuit
  | Dimacs -> Dimacs.Print.sat_file d
  | Qdimacs ->  Dimacs.Print.qbf_file d in
  printf "%s\n" output

let start () =
  let mode, fs = get () in
  let file = List.concat_map (Parse.from_file ()) fs in
  let file = Ast.file file in
  (*printf "%s\n\n" (Ast.Print.file p);*)
  match mode with
  | Either.Right comm -> solve_mode file comm(*solve models (d, vm, im, hs) comm*)
  | Either.Left comm -> ground_mode file comm

let _ = start ()


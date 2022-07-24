open Printf
open Bule

type output_format = Bule | Dimacs | Qdimacs

let version_string = "<development>"
(*let version_string = "4.0.0"*)

let read_output_format = function
  | "bule" -> Bule
  | "dimacs" -> Dimacs
  | "qdimacs" -> Qdimacs
  | _ -> assert false
let read_show_format = function
  | "all" -> Circuit.ShowAll
  | "positive" -> Positive
  | "none" -> Circuit.ShowNone
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
  let models = ref 1 in
  let solver = ref "quantor" in
  let grounder = ref "native" in
  let solve = ref false in
  let facts = ref false in
  let print_version () = eprintf "BULE Grounder, version %s.\n" version_string; exit 0 in
  let print_mode = ref Bule in
  let output_symbols = ["qdimacs"; "dimacs"; "bule"]
  and output_treat s = print_mode := read_output_format s in
  let default_show = ref Circuit.Positive in
  let show_symbols = ["all"; "positive"; "none"]
  and show_treat s = default_show := read_show_format s in
  let speclist = [("-",        Arg.Unit (fun () -> update input_names ("-" :: !input_names)), "Read the BULE code from the standard input.");
                  ("--solve",  Arg.Set solve, "Enable solving. Default: \"false\"");
                  ("--models", Arg.Set_int models, "Number of models to generate. The option has no effect if \"solve\" is set to \"false\". Default: 1.");
                  ("--solver", Arg.Set_string solver, "Set the solver to be used. If \"quantor\" then Quantor 3.2 is used, if \"minisat\" then Minisat 1.14 is used, otherwise the argument is assumed to be a command-line tool. Example \"depqbf --no-dynamic-nenofex --qdo\". The option has no effect if \"solve\" is set to \"false\". Default: \"quantor\"");
                  ("--grounder", Arg.Set_string grounder, "Set the grounder to be used. If \"native\" then the default embedded grounder is used, if \"gringo\" then the Potassco grounder gringo is used with suitable options. Otherwise the argument is assumed to be a command-line tool. Default: \"native\"");
                  ("--output", Arg.Symbol (output_symbols, output_treat), " Output format (QDIMACS, DIMACS, or BULE). The option has no effect if \"solve\" is set to \"true\". Default \"bule\".");
                  ("--facts",  Arg.Set facts, "Enable printing of grounding facts. The option has no effect if \"solve\" is set to \"true\". Default: \"false\".");
                  ("--default_show", Arg.Symbol (show_symbols, show_treat), " Default showing behaviour for literals. The option has no effect if \"solve\" is set to \"false\". Default \"positive\".");
                  ("--version", Arg.Unit print_version, "Display the version number.")
                 ] in
  let usage_msg = sprintf "BULE Grounder %s. Options available:" version_string in
  let add_name s = input_names := s :: !input_names in
  Arg.parse speclist add_name usage_msg;
  let files = match List.rev !input_names with
  | [] -> failwith "Wrong number of arguments. Usage: bule file"
  | _ :: _ as names -> names in
  let solver = match !solver with
    | "quantor" -> Solve.Quantor
    | "minisat" -> Solve.Minisat
    | _ -> Solve.CommandLine !solver in
  let grounder = match !grounder with
    | "native" -> Circuit.Native
    | "gringo" -> Circuit.CommandLine "gringo --text"
    | _ -> Circuit.CommandLine !grounder in
  let goptions = { Circuit.facts = !facts; tool = grounder; show = !default_show } in
  let mode = if !solve then Either.Right (solver, !models) else Either.Left !print_mode in
  (mode, goptions, files)

let solve_mode goptions file solve_options =
  let circuit = Circuit.file goptions file in
  let file = Dimacs.ground circuit in
  Solve.solve_all solve_options file
let ground_mode goptions file format =
  let circuit = Circuit.file goptions file in
  let d = Dimacs.file circuit in
  let output = match format with
  | Bule -> Circuit.Print.file circuit
  | Dimacs -> Dimacs.Print.sat_file d
  | Qdimacs ->  Dimacs.Print.qbf_file d in
  if output <> "" then printf "%s\n" output else printf "%s" output

let start () =
  let mode, goptions, fs = get () in
  let file = List.concat_map (Parse.from_file ()) fs in
  let file = Ast.file file in
  (*printf "%s\n\n" (Ast.Print.file p);*)
  match mode with
  | Either.Right comm -> solve_mode goptions file comm(*solve models (d, vm, im, hs) comm*)
  | Either.Left comm -> ground_mode goptions file comm

let _ = start ()


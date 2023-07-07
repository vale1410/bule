open Printf
module B = Bule

let version_string = "<development>"
(*let version_string = "4.0.0"*)

type output_format = Bule | Dimacs | Qdimacs
type mode = Ground | Solve | Enumerate

let read_output_format = function
  | "bule" -> Bule
  | "dimacs" -> Dimacs
  | "qdimacs" -> Qdimacs
  | _ -> assert false
let read_mode = function
  | "enumerate" -> Enumerate
  | "ground" -> Ground
  | "solve" -> Solve
  | _ -> assert false
let read_show_format = function
  | "all" -> B.Circuit.ShowAll
  | "positive" -> B.Circuit.Positive
  | "none" -> B.Circuit.ShowNone
  | _ -> assert false

type options = { mode : mode;
                 output_format : output_format;
                 models : int;
                 solver : B.Solve.solver;
                 ground_options : B.Circuit.option;
                 files : string list }

let update reference value = reference := value
let get () =
  let input_names = ref [] in
  let models = ref 1 in
  let solver = ref "quantor" in
  let grounder = ref "native" in
  let facts = ref false in
  let print_version () = eprintf "BULE Grounder, version %s.\n" version_string; exit 0 in
  let mode_option = ref Ground in
  let mode_symbols = ["enumerate"; "ground"; "solve"]
  and mode_treat s = mode_option := read_mode s in
  let output_option = ref Bule in
  let output_symbols = ["qdimacs"; "dimacs"; "bule"]
  and output_treat s = output_option := read_output_format s in
  let show_option = ref B.Circuit.Positive in
  let show_symbols = ["all"; "positive"; "none"]
  and show_treat s = show_option := read_show_format s in
  let speclist =
    [("-",        Arg.Unit (fun () -> update input_names ("-" :: !input_names)), "Read the BULE code from the standard input.");
     ("--solve",  Arg.Unit (fun () -> update mode_option Enumerate), "Set mode to enumerate.");
     ("--mode", Arg.Symbol (mode_symbols, mode_treat), " Running mode. Default \"ground\".");
     ("--models", Arg.Set_int models, "Number of models to generate. The option has no effect if \"mode\" is not set to \"enumerate\". \"0\" generates all models. Default: 1.");
     ("--solver", Arg.Set_string solver, "Set the solver to be used. If \"quantor\" then Quantor 3.2 is used, if \"minisat\" then Minisat 1.14 is used, otherwise the argument is assumed to be a command-line tool. Example \"depqbf --no-dynamic-nenofex --qdo\". The option has no effect if \"mode\" is set to \"ground\". Default: \"quantor\"");
     ("--grounder", Arg.Set_string grounder, "Set the grounder to be used. If \"native\" then the default embedded grounder is used, if \"gringo\" then the Potassco grounder gringo is used with suitable options. Otherwise the argument is assumed to be a command-line tool. Default: \"native\"");
     ("--output", Arg.Symbol (output_symbols, output_treat), " Output format (QDIMACS, DIMACS, or BULE). Default \"bule\".");
     ("--facts",  Arg.Set facts, "Enable printing of grounding facts. The option has no effect if \"solve\" is set to \"true\". Default: \"false\".");
     ("--default_show", Arg.Symbol (show_symbols, show_treat), " Default showing behaviour for literals. The option has no effect if \"mode\" is set to \"ground\". Default \"positive\".");
     ("--version", Arg.Unit print_version, "Display the version number.")
    ] in
  let usage_msg = sprintf "Bule Grounder %s. Options available:" version_string in
  let add_name s = input_names := s :: !input_names in
  Arg.parse speclist add_name usage_msg;
  let files = match List.rev !input_names with
  | [] -> failwith "Wrong number of arguments. Usage: bule file"
  | _ :: _ as names -> names in
  let solver = match !solver with
    | "quantor" -> B.Solve.Quantor
    | "minisat" -> B.Solve.Minisat
    | _ -> B.Solve.CommandLine !solver in
  let grounder = match !grounder with
    | "native" -> B.Circuit.Native
    | "gringo" -> B.Circuit.CommandLine "gringo --text"
    | _ -> B.Circuit.CommandLine !grounder in
  let ground_options = { B.Circuit.facts = !facts; tool = grounder; show = !show_option } in
  { mode = !mode_option; output_format = !output_option; models = !models; solver; ground_options; files }

let enumerate_mode options file =
  let circuit = B.Circuit.file options.ground_options file in
  let file = B.Dimacs.ground circuit in
  B.Solve.solve_all ((options.solver, options.output_format = Dimacs), options.models) file
let solve_mode options file =
  let circuit = B.Circuit.file options.ground_options file in
  let file = B.Dimacs.ground circuit in
  B.Solve.solve_one (options.solver, options.output_format = Dimacs) file
let ground_mode options file =
  let circuit = B.Circuit.file options.ground_options file in
  let d = B.Dimacs.file circuit in
  let output = match options.output_format with
  | Bule -> B.Circuit.Print.file circuit
  | Dimacs -> B.Dimacs.Print.sat_file d
  | Qdimacs ->  B.Dimacs.Print.qbf_file d in
(*  if output <> "" then printf "%s\n" output else printf "%s" output*)
  printf "%s" output

let start () =
  (*let mode, goptions, fs = get () in*)
  let options = get () in
  let file = List.concat_map (B.Parse.from_file ()) options.files in
  let file = B.Ast.file file in
  (*printf "%s\n\n" (Ast.Print.file p);*)
  match options.mode with
  | Enumerate -> enumerate_mode options file
  | Ground -> ground_mode options file
  | Solve -> solve_mode options file
  (*| Either.Right comm -> solve_mode goptions file comm(*solve models (d, vm, im, hs) comm*)
  | Either.Left comm -> ground_mode goptions file comm*)

let _ = start ()


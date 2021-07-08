open Printf

include Types.DIMACS

module P = Print
module Print =
struct
  let search_var v = sprintf "%d" v
  let literal (pol, v) = if pol then sprintf "%d" v else  sprintf "-%d" v
  let clause lits = sprintf "%s 0" (Print.unspaces literal lits)
  let quantifier_block (exist, vars) = sprintf "%s %s 0" (if exist then "e" else "a") (Print.unspaces search_var vars)
  let file (vmax, cmax, blocks, clauses) =
    sprintf "p cnf %d %d\n%s\n%s\n" vmax cmax (Print.unlines quantifier_block blocks) (Print.unlines clause clauses)
end


module SVMap = Map.Make (struct type t = Circuit.T.search_var let compare = compare end)

let search_var (map, g) v = match SVMap.find_opt v map with
  | Some i -> ((map, g), i)
  | None -> let i = g+1 in ((SVMap.add v i map, i), i)

let quantifier_block accu (exis, vars) =
  let (accu, vars) = List.fold_left_map search_var accu vars in
  (accu, (exis, vars))
let literal accu (pol, v) = let (naccu, i) = search_var accu v in (naccu, (pol, i))
let flip_polarity (p, v) = (not p, v)
let clause (accu, nbcls) (hyps, ccls) =
  let lits = List.rev_append (List.rev_map flip_polarity hyps) ccls in
  let (naccu, cls) = List.fold_left_map literal accu lits in
  ((naccu, nbcls+1), cls)

let compute_new_vars (_, nb) (cl_map, _) =
  let vars = SVMap.bindings cl_map in
  let filter (f, i) = if i > nb then Some f else None in
  List.filter_map filter vars
let file (qbs, cls) =
  let accu = (SVMap.empty, 0) in
  let (naccu, qbs) = List.fold_left_map quantifier_block accu qbs in
  let ((nnaccu, nbcls), cls) = List.fold_left_map clause (naccu, 0) cls in
  let nvars = compute_new_vars naccu nnaccu in
  if nvars <> [] then eprintf "Warning: undeclared vars:\n%s\n%!" (P.unspaces Circuit.Print.search_var nvars);
  (snd nnaccu, nbcls, qbs, cls)

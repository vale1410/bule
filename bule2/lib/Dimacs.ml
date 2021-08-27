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


let search_var (map1, map2, g) v = match T.VMap.find_opt v map1 with
  | Some i -> ((map1, map2, g), i)
  | None -> let i = g+1 in ((T.VMap.add v i map1, T.IMap.add i v map2, i), i)

let quantifier_block accu (exis, vars) =
  let (accu, vars) = List.fold_left_map search_var accu vars in
  (accu, (exis, vars))
let literal accu (pol, v) = let (naccu, i) = search_var accu v in (naccu, (pol, i))
let flip_polarity (p, v) = (not p, v)
let clause (accu, nbcls) (hyps, ccls) =
  let lits = List.rev_append (List.rev_map flip_polarity hyps) ccls in
  let (naccu, cls) = List.fold_left_map literal accu lits in
  ((naccu, nbcls+1), cls)

let hide_vars vmap hide =
  let hide_one (pol, sv) = match T.VMap.find_opt sv vmap with
    | None -> Either.Left (pol, sv)
    | Some i -> Either.Right (if pol then i else -i) in
  let (undeclared, hide) = List.partition_map hide_one hide in
  if undeclared <> [] then eprintf "Warning. Hiding undeclared variables: %s\n" (P.unspaces Circuit.Print.literal undeclared);
  T.ISet.of_list hide

let compute_new_vars (_, _, nb) cl_map =
  let vars = T.VMap.bindings cl_map in
  let filter (f, i) = if i > nb then Some f else None in
  List.filter_map filter vars
let ground (qbs, cls, hide) : T.file * int T.VMap.t * Circuit.T.search_var T.IMap.t * T.ISet.t =
  let accu = (T.VMap.empty, T.IMap.empty, 0) in
  let (naccu, qbs) = List.fold_left_map quantifier_block accu qbs in
  let (((vmap, imap, nvar), nbcls), cls) = List.fold_left_map clause (naccu, 0) cls in
  let nvars = compute_new_vars naccu vmap in
  if nvars <> [] then eprintf "Warning. Undeclared variables: %s\n%!" (P.unspaces Circuit.Print.search_var nvars);
  let hide = hide_vars vmap hide in
  ((nvar, nbcls, qbs, cls), vmap, imap, hide)
let file (args : Circuit.T.file) : T.file =
  let (dimacs, _, _, _) = ground args in
  dimacs

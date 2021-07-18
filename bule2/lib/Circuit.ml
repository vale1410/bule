open Printf

include Types.CIRCUIT
open T

(*type pre_ground_term = Ground of ground_term | Int of int*)

module P = Print
module Print =
struct
  let rec ground_term = function
    | Fun (c, []) -> sprintf "%s" c
    | Fun (c, (_ :: _ as ts)) -> sprintf "%s(%s)" c (Print.list' "" "," "" ground_term ts)
  let search_var (n, cs) = match cs with
    | [] -> Ast.Print.cname n
    | _ :: _ -> sprintf "%s(%s)" (Ast.Print.cname n) (Print.list' "" "," "" ground_term cs)

  let literal (b, var) = (if b then "" else "~") ^ search_var var
  let clause (hyps, ccls) =
    let hs = Print.list' "" " & " "" literal hyps in
    let cs = Print.list' "" " | " "." literal ccls in
    match hyps with
    | [] -> cs
    | _ :: _ -> hs ^ " -> " ^ cs
  let quantifier b = if b then "exists" else "forall"
  let blocks l =
    let f (i, s) (b, vars) =
      let pr_one var = sprintf "#%s[%d] :: %s?" (quantifier b) i (search_var var) in
      let s' = Print.unlines pr_one vars in
      (i+1, s ^ "\n" ^ s') in
    snd (List.fold_left f (0, "") l)
  let file (bl, cl) = blocks bl ^ "\n" ^ Print.unlines clause cl
end

module SMap = Map.Make (String)
module IMap = Map.Make (Int)
module GTermSet = Set.Make (struct type t = ground_term list let compare = compare end)
let find_default key map d = match SMap.find_opt key map with
  | None -> d
  | Some v -> v

let print_smap pr map =
  let binds = SMap.bindings map in
  let pr (key, elt) = sprintf "%s -> %s" key (pr elt) in
  P.unlines pr binds
let print_gtset s =
  let pr = P.list' "(" "," ")" Print.ground_term in
  P.list' "{" ", " "}" pr (GTermSet.elements s)

let unions ls = List.fold_left GTermSet.union GTermSet.empty ls

let subtract_strings i1 is =
  let i2 = Misc.sum is in
  (i1 - i2)

let perform_eop eop is = match eop with
  | Ast.T.Add -> Misc.sum is
  | Ast.T.Mult -> Misc.product is
  | Ast.T.Max -> List.fold_left max min_int is
  | Ast.T.Min -> List.fold_left min max_int is

let perform_ord_op rop v1 v2 = match rop with
  | Ast.T.Lt -> compare v1 v2 < 0
  | Ast.T.Gt -> compare v1 v2 > 0
  | Ast.T.Leq -> compare v1 v2 <= 0
  | Ast.T.Geq -> compare v1 v2 >= 0
let perform_eq_op rop v1 v2 = match rop with
  | Ast.T.Eq -> compare v1 v2 = 0
  | Ast.T.Neq -> compare v1 v2 <> 0

exception UnboundVar of Ast.T.vname

let int_of_string_opt s = try Some (int_of_string s) with Failure "int_of_string" -> None

let rec expr vmap : Ast.T.expr -> int = function
  | Ast.T.VarE n -> (match SMap.find_opt n vmap with
                    | None -> raise (UnboundVar n)
                    | Some (Fun (si, args) as g) -> match args, int_of_string_opt si with
                                               | [], Some i -> i
                                               | _ :: _ , _
                                               | _, None -> failwith (sprintf "Variable %s ground to a non-int %s" n (Print.ground_term g)))
  | Int i -> i
  | ListE (eop, e, es) -> perform_eop eop (List.map (expr vmap) (e :: es))
  | Subtract (e, es) -> subtract_strings (expr vmap e) (List.map (expr vmap) es)

let match_term pterm gterm vmap =
  let map = ref vmap in
  let rec aux t gt =
    let (c', gts) = match gt with Fun g -> g in
    match t with
    | Ast.T.Fun (c, ts) -> if c <> c' || List.length ts <> List.length gts then raise Exit else List.iter2 aux ts gts
    | Ast.T.Exp e -> let c = string_of_int (expr !map e) in if c <> c' || gts <> [] then raise Exit
    | Ast.T.Var v -> match SMap.find_opt v !map with | Some gt -> if gt <> gterm then raise Exit
                                                     | None -> map := SMap.add v gt !map in
  try aux pterm gterm; Some !map with
  | Exit -> None

let term vmap t =
  let rec aux = function
  | Ast.T.Fun (c, ts) -> Fun (c, List.map aux ts)
  | Ast.T.Exp e -> let c = string_of_int (expr vmap e) in Fun (c, [])
  | Ast.T.Var v -> match SMap.find_opt v vmap with | Some gt -> gt
                                                   | None -> raise (UnboundVar v) in
  try aux t with
  | UnboundVar v -> failwith (sprintf "Unbound variable %s in term %s" v (Ast.Print.term t))
(*let atom gmap vmap (cname, terms) =
  let instances = match SMap.find_opt cname gmap with
    | None -> GTermSet.empty
    | Some set -> set in
  let f m t gt = Option.bind m (match_term t gt) in
  let aux ts gts =
    if List.length ts <> List.length gts then failwith (sprintf "Error: term list %s incompatible with %s" (P.list Ast.Print.term ts) (P.list Print.ground_term gts));
    List.fold_left2 f (Some vmap) ts gts in
  try List.filter_map (aux terms) (GTermSet.elements instances) with
  | UnboundVar n -> failwith (sprintf "Error: variable %s is unbound in an arithmetical expression when grounding %s." n (Ast.Print.atom (cname, terms)))*)

let atom gmap vmap (cname, terms) =
  let instances = find_default cname gmap GTermSet.empty in
  try let ts = List.map (term vmap) terms in
      if GTermSet.mem ts instances then [vmap] else []
  with _ ->
        let f m t gt = Option.bind m (match_term t gt) in
        let aux ts gts =
          if List.length ts <> List.length gts then failwith (sprintf "Error: term list %s incompatible with %s" (P.list Ast.Print.term ts) (P.list Print.ground_term gts));
          List.fold_left2 f (Some vmap) ts gts in
        try List.filter_map (aux terms) (GTermSet.elements instances) with
        | UnboundVar n -> failwith (sprintf "Error: variable %s is unbound in an arithmetical expression when grounding %s." n (Ast.Print.atom (cname, terms)))

let ground_literal gmap vmap = function
  | Ast.T.In ga -> atom gmap vmap ga
  | Ast.T.Notin ga ->
     let maps = atom gmap vmap ga in
     if maps = [] then [vmap] else []
  | Ast.T.Sorted (t1, r, t2) ->
     let t1 = expr vmap t1
     and t2 = expr vmap t2 in
     if perform_ord_op r t1 t2 then [vmap] else []
  | Ast.T.Equal (t1, r, t2) ->
     let t1 = term vmap t1
     and t2 = term vmap t2 in
     if perform_eq_op r t1 t2 then [vmap] else []

let glits gmap vmap l =
  let aux vmaps lit = List.concat_map (fun m -> ground_literal gmap m lit) vmaps in
  List.fold_left aux [vmap] l

let tuple vmap : Ast.T.tuple -> ground_term list  = function
  | Ast.T.Term t -> [term vmap t]
  | Ast.T.Range (e1, e2) ->
     let i1, i2 = expr vmap e1, expr vmap e2 in
     let make_int i = Fun (string_of_int i, []) in
     List.map make_int (Misc.range i1 (i2+1))

let tuple_list vmap l = Misc.cross_products (List.map (tuple vmap) l)

let ground_decl gmap (gls, n, ts) =
  let maps = glits gmap SMap.empty gls in
  let g = find_default n gmap GTermSet.empty in
  let aux gm m =
    let l = tuple_list m ts in
    GTermSet.union gm (GTermSet.of_list l) in
  let set = List.fold_left aux g maps in
  (*eprintf "gmap=%s\nname=%s\nset=%s\n%!" (print_smap print_gtset gmap) n (print_gtset set);*)
  SMap.add n set gmap

let rec ground_decl_component gmap comp =
  let map = List.fold_left ground_decl gmap comp in
  (*eprintf "%s.%!" (print_smap print_gtset gmap);*)
  if map <> gmap then ground_decl_component map comp else map

let find_deps_glit = function
  | Ast.T.In (n, _) -> Some (Either.Right n)
  | Ast.T.Notin (n, _) -> Some (Either.Left n)
  | Ast.T.Sorted _ | Ast.T.Equal _ -> None
let with_neg_cycle negdeps sccs =
  let test_component comp =
    let test_element e =
      assert (SMap.mem e negdeps);
      let deps = SMap.find e negdeps in
      List.exists (fun x -> List.mem x comp) deps in
    if List.exists test_element comp then Some comp else None in
  List.filter_map test_component sccs

let compute_recursive_components decls =
  let add_dep map (gls, n, _) =
    let ds = List.filter_map find_deps_glit gls in
    let negs, poss = List.partition_map Fun.id ds in
    let nl, l = find_default n map ([], []) in
    SMap.add n (negs @ nl, poss @ negs @ l) map in
  let dep_map = List.fold_left add_dep SMap.empty decls in
  let negdeps = SMap.map fst dep_map in
  let alldeps = SMap.map snd dep_map in
  let deps = SMap.bindings alldeps in
  let self_deps = List.filter_map (fun (key, ds) -> if List.mem key ds then Some key else None) deps in
  let sccs = Tsort.sort_strongly_connected_components deps in
  let neg_cycles = with_neg_cycle negdeps sccs in
  if neg_cycles <> [] then failwith (sprintf "Recursion cycle through negation: %s" (P.list (P.unspaces Ast.Print.cname) neg_cycles));
  let is_rec = function
    | [] -> assert false
    | a :: [] -> if List.mem a self_deps then Either.Left [a] else Either.Right a
    | _ :: _ :: _ as comp -> Either.Left comp in
  List.map is_rec sccs

let all_ground decls =
  let sccs = compute_recursive_components decls in
  let left comp = List.filter (fun (_, n, _) -> List.mem n comp) decls
  and right name = List.filter (fun (_, n, _) -> n = name) decls in
  let grouped_decls = List.map (Either.map ~left ~right) sccs in
  let aux gmap = function
    | Either.Left recurs -> ground_decl_component gmap recurs
    | Either.Right simple -> List.fold_left ground_decl gmap simple in
  List.fold_left aux SMap.empty grouped_decls

let search_var vmap ((cname, terms) : Ast.T.atom) = (cname, List.map (term vmap) terms)

let search_decl gmap qmap ((gls, b, e, a) : Ast.T.search_decl) =
  let maps = glits gmap SMap.empty gls in
  let parity = if b then 1 else 0 in
  let update qm i var =
    let f = function | None -> Some [var] | Some l -> Some (var :: l) in
    IMap.update (2 * i + parity) f qm in
  let treat_one qm vmap =
    let i = expr vmap e in
    let var = search_var vmap a in
    update qm i var in
  List.fold_left treat_one qmap maps

let all_search gmap (decls : Ast.T.search_decl list) =
  let qmap = List.fold_left (search_decl gmap) IMap.empty decls in
  if IMap.is_empty qmap then []
  else
    let blocks = IMap.bindings qmap in
    let f (i, l) = (i mod 2 = 1, l) in
    List.map f blocks


let literals gmap vmap (gls, pol, ga) =
  let maps = glits gmap vmap gls in
  List.map (fun m -> (pol, search_var m ga)) maps

let clause_decl gmap (gls, hyps, ccls) =
  let maps = glits gmap SMap.empty gls in
  let make_clause vmap =
    let hyps = List.concat_map (literals gmap vmap) hyps in
    let ccls = List.concat_map (literals gmap vmap) ccls in
    (hyps, ccls) in
  Misc.map make_clause maps

let all_clause gmap decls = List.concat_map (clause_decl gmap) decls

(*let clause_decl gmap accu (gls, hyps, ccls) =
  let maps = glits gmap SMap.empty gls in
  let make_clause acc vmap =
    let hyps = List.concat_map (literals gmap vmap) hyps in
    let ccls = List.concat_map (literals gmap vmap) ccls in
    (hyps, ccls) :: acc in
  List.fold_left make_clause accu maps

let all_clause gmap decls = List.fold_left (clause_decl gmap) [] decls*)

let file decls =
  let aux (gs, ss, cs) = function
    | Ast.T.G gd -> (gd :: gs, ss, cs)
    | S sd -> (gs, sd :: ss, cs)
    | C cd -> (gs, ss, cd :: cs) in
  let gs, ss, cs = List.fold_left aux ([], [], []) decls in
  let gs, ss, cs = List.rev gs, List.rev ss, List.rev cs in
  let gmap = all_ground gs in
  (*eprintf "%s\n%!" (print_smap print_gtset gmap);*)
  let bloc = all_search gmap ss in
  let clau = all_clause gmap cs in
  (bloc, clau)

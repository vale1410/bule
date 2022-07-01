open Printf

include Types.CIRCUIT
open T

(*type pre_ground_term = Ground of ground_term | Int of int*)

module P = Print
module Print =
struct
  let comma_list p l = Print.list' "" "," "" p l
  let comma_s_list p l = Print.list' "" ", " "" p l
  let rec ground_term = function
    | Fun (c, []) -> sprintf "%s" c
    | Fun (c, (_ :: _ as ts)) -> sprintf "%s(%s)" c (comma_list ground_term ts)
  let ground_fact (n, cs) = match cs with
    | [] -> Ast.Print.cname n
    | _ :: _ -> sprintf "%s[%s]" (Ast.Print.cname n) (comma_list ground_term cs)
  let search_var (n, cs) = match cs with
    | [] -> Ast.Print.cname n
    | _ :: _ -> sprintf "%s(%s)" (Ast.Print.cname n) (comma_list ground_term cs)

  let literal (b, var) = (if b then " " else "~") ^ search_var var
  let clause (hyps, ccls) =
    let hs = Print.list' "" " & " "" literal hyps in
    let cs = Print.list' "" " | " "." literal ccls in
    match hyps with
    | [] -> cs
    | _ :: _ -> hs ^ " -> " ^ cs
  let quantifier b = if b then "exists" else "forall"
  let blocks l =
    let f (i, s) (b, vars) =
      let s' = sprintf "#%s[%d] %s." (quantifier b) i (comma_s_list search_var vars) in
      (i+1, s ^ "\n" ^ s') in
    snd (List.fold_left f (0, "") l)
  let file { prefix; matrix; hide; show } =
    let h = if hide <> [] then sprintf "\n#hide %s." (comma_s_list literal hide) else ""
    and s = if show <> [] then sprintf "\n#show %s." (comma_s_list literal show) else "" in
    sprintf "%s\n%s%s%s" (blocks prefix) (Print.unlines clause matrix) h s
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

let rec pow a = function
  | 0 -> 1
  | 1 -> a
  | n ->
    let b = pow a (n / 2) in
    b * b * (if n mod 2 = 0 then 1 else a)

let rec log a = function
  | 0 -> assert false
  | n -> if n < a then 0 else 1 + log a (n / a)

let perform_eop v1 v2 = function
  | Ast.T.Add -> v1 + v2
  | Ast.T.Div -> v1 / v2
  | Ast.T.Mult -> v1 * v2
  | Ast.T.Max -> max v1 v2
  | Ast.T.Min -> min v1 v2
  | Ast.T.Log -> log v1 v2
  | Ast.T.Mod -> v1 mod v2
  | Ast.T.Pow -> pow v1 v2
  | Ast.T.Sub -> v1 - v2

let int_of_gt = function
  | T.Fun (name, []) as gt -> (match int_of_string_opt name with Some i -> Either.Right i | None -> Either.Left gt)
  | T.Fun (_, _ :: _) as gt -> Either.Left gt

let perform_cop_aux v1 v2 = function
  | Ast.T.Lt -> compare v1 v2 < 0
  | Ast.T.Gt -> compare v1 v2 > 0
  | Ast.T.Leq -> compare v1 v2 <= 0
  | Ast.T.Geq -> compare v1 v2 >= 0
  | Ast.T.Eq -> compare v1 v2 = 0
  | Ast.T.Neq -> compare v1 v2 <> 0

let perform_cop v1 v2 cop = perform_cop_aux (int_of_gt v1) (int_of_gt v2) cop

exception UnboundVar of Ast.T.vname
exception NonInt of (Ast.T.vname * ground_term)

(*let int_of_string_opt s = try Some (int_of_string s) with Failure "int_of_string" -> None*)

let find_gt v map = match SMap.find_opt v map with
  | None -> raise (UnboundVar v)
  | Some g -> g

let rec expr vmap : Ast.T.expr -> int = function
  | Ast.T.VarE n -> (match find_gt n vmap with
                    | Fun (si, args) as g -> match args, int_of_string_opt si with
                                             | [], Some i -> i
                                             | _ :: _ , _
                                             | _, None -> raise (NonInt (n, g)))
  | Int i -> i
  | BinE (e1, bo, e2) -> perform_eop (expr vmap e1) (expr vmap e2) bo

let match_term pterm gterm vmap =
  let map = ref vmap in
  let rec aux t gt =
    let (c', gts) = match gt with Fun g -> g in
    match t with
    | Ast.T.Fun (c, ts) -> if c <> c' || List.length ts <> List.length gts then raise Exit else List.iter2 aux ts gts
    | Ast.T.Exp (Ast.T.VarE v) -> (match SMap.find_opt v !map with | Some gt -> if gt <> gterm then raise Exit
                                                      | None -> map := SMap.add v gt !map)
    | Ast.T.Exp e -> let c = string_of_int (expr !map e) in if c <> c' || gts <> [] then raise Exit in
  try aux pterm gterm; Some !map with
  | Exit -> None

let term vmap t =
  let rec aux = function
  | Ast.T.Fun (c, ts) -> Fun (c, List.map aux ts)
  | Ast.T.Exp (Ast.T.VarE v) -> find_gt v vmap
  | Ast.T.Exp e -> let c = string_of_int (expr vmap e) in Fun (c, []) in
(*  | Ast.T.Var v -> find_gt v vmap in*)
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
  | Ast.T.Comparison (t1, c, t2) ->
     let t1 = term vmap t1
     and t2 = term vmap t2 in
     if perform_cop t1 t2 c then [vmap] else []
  | Ast.T.Set (v, t) ->
     if SMap.mem v vmap then failwith (sprintf "Error: variable %s is already bound and cannot be assign %s." v (Ast.Print.term t));
     let t = term vmap t in
     [SMap.add v t vmap]

let glits gmap vmap l =
  let aux vmaps lit = List.concat_map (fun m -> ground_literal gmap m lit) vmaps in
  List.fold_left aux [vmap] l

(*let tuple vmap : Ast.T.tuple -> ground_term list  = function
  | Ast.T.Term t -> [term vmap t]
  | Ast.T.Range (e1, e2) ->
     let i1, i2 = expr vmap e1, expr vmap e2 in
     let make_int i = Fun (string_of_int i, []) in
     List.map make_int (Misc.range i1 (i2+1))*)

let rec tuple_aux vmap = function
  | Ast.T.FunTu (c, ts) ->
     let ls = tuple' vmap ts in
     List.map (fun l -> Fun (c, l)) ls
  | Ast.T.Range (e1, e2) ->
     let i1, i2 = expr vmap e1, expr vmap e2 in
     let make_int i = Fun (string_of_int i, []) in
     List.map make_int (Misc.range i1 (i2+1))
  | Ast.T.ExpTu (Ast.T.VarE v) -> [find_gt v vmap]
  | Ast.T.ExpTu e -> let c = string_of_int (expr vmap e) in [Fun (c, [])]
and tuple' vmap ts =
  let tss = List.map (tuple_aux vmap) ts in
  Misc.cross_products tss
(*let tuple vmap t =
  try tuple_aux vmap t with
  | UnboundVar v -> failwith (sprintf "Unbound variable %s in tuple %s" v (Ast.Print.tuple t))
let atomd vmap (n, l) = tuple vmap (Ast.T.FunTu (n, l))*)

let atomd_aux vmap l = tuple' vmap l

let ground_decl_aux gmap (gls, n, l) =
  let maps = glits gmap SMap.empty gls in
  let set = find_default n gmap GTermSet.empty in
  let aux set m = List.fold_right GTermSet.add (atomd_aux m l) set in
  let set = List.fold_left aux set maps in
  SMap.add n set gmap

 (*    let atoms m = List.concat_map (atomd m) ts in
  * (n, l)
  *  let g n = find_default n gmap GTermSet.empty in
  *    let l = tuple_list m ts in
  *    GTermSet.union gm (GTermSet.of_list l) in
  *  let set = List.fold_left aux g maps in
  *  (\*eprintf "gmap=%s\nname=%s\nset=%s\n%!" (print_smap print_gtset gmap) n (print_gtset set);*\)
  *  SMap.add n set gmap *)

let ground_decl gmap decl =
  try ground_decl_aux gmap decl with
  | NonInt (n, g) ->
     let glits, name, ts = decl in
     let d = (glits, [(name, ts)]) in
     failwith (sprintf "Variable %s ground to a non-int %s in declaration \"%s\"" n (Print.ground_term g) (Ast.Print.ground_decl d))

let rec ground_decl_component gmap comp =
  let map = List.fold_left ground_decl gmap comp in
  (*eprintf "%s.%!" (print_smap print_gtset gmap);*)
  if SMap.equal (GTermSet.equal) map gmap then map else ground_decl_component map comp

let find_deps_glit = function
  | Ast.T.In (n, _) -> Some (Either.Right n)
  | Ast.T.Notin (n, _) -> Some (Either.Left n)
  | Ast.T.Comparison _ | Ast.T.Set _ -> None

(** Filters strongly connected components that contains a negative dependency *)
let with_neg_cycle negdeps sccs =
  let test_component comp =
    let test_element e =
      match SMap.find_opt e negdeps with
      | None -> eprintf "Warning. Grounding predicate %s used in a rule without being defined anywhere.\n%!" e; false
      | Some deps ->
         List.exists (fun x -> List.mem x comp) deps in
    if List.exists test_element comp then Some comp else None in
  List.filter_map test_component sccs

let compute_recursive_components decls =
  let add_dep (nmap, all) (gls, n, _) =
    let ds = List.filter_map find_deps_glit gls in
    let negs, poss = List.partition_map Fun.id ds in
    let nl = find_default n nmap [] in
    (SMap.add n (negs @ nl) nmap, (n, poss) :: (n, negs) :: all) in
  let negdeps, deps = List.fold_left add_dep (SMap.empty, []) decls in
  let self_deps = List.filter_map (fun (key, ds) -> if List.mem key ds then Some key else None) deps in
  let sccs = Tsort.sort_strongly_connected_components deps in
  let neg_cycles = with_neg_cycle negdeps sccs in
  if neg_cycles <> [] then failwith (sprintf "Recursion cycle through negation: %s" (P.list (P.unspaces Ast.Print.cname) neg_cycles));
  let is_rec = function
    | [] -> assert false
    | a :: [] -> if List.mem a self_deps then Either.Left [a] else Either.Right a
    | _ :: _ :: _ as comp -> Either.Left comp in
  List.map is_rec sccs

let all_ground decls : GTermSet.t SMap.t =
  let decls = List.concat_map (fun (glits, atomds) -> List.map (fun (n, atomd) -> (glits, n, atomd)) atomds) decls in
  let sccs = compute_recursive_components decls in
  let left comp = List.filter (fun (_, n, _) -> List.mem n comp) decls
  and right name = List.filter (fun (_, n, _) -> n = name) decls in
  let grouped_decls = List.map (Either.map ~left ~right) sccs in
  let aux gmap = function
    | Either.Left recurs -> ground_decl_component gmap recurs
    | Either.Right simple -> List.fold_left ground_decl gmap simple in
  List.fold_left aux SMap.empty grouped_decls

let search_var ((cname, terms) : Ast.T.atom) vmap = (cname, List.map (term vmap) terms)
let search_lit vmap ((pol, var) : Ast.T.literal) = (pol, search_var var vmap)
let search_lits lits vmap = List.map (search_lit vmap) lits
let search_level vmap (b, e, vars) =
  let parity = if b then 1 else 0 in
  let i = expr vmap e in
  let level = 2 * i + parity in
  let treat_one (n, args) : (int * search_var) list =
    List.rev_map (fun args -> (level, (n, args))) (tuple' vmap args) in
  List.concat_map treat_one vars
let search_exists_inner vmap vars =
  let treat_one (n, args) : search_var list =
    List.rev_map (fun args -> (n, args)) (tuple' vmap args) in
  List.concat_map treat_one vars

let search_decl gmap (qmap, inner) (gls, (decl : Ast.T.search_decl)) =
  let maps = glits gmap SMap.empty gls in
  match decl with
  | Ast.T.Level level ->
     let f var = function | None -> Some [var] | Some l -> Some (var :: l) in
     let update qm (i, var) = IMap.update i (f var) qm in
     let treat_vmap qm vmap = List.fold_left update qm (search_level vmap level) in
     let qm = List.fold_left treat_vmap qmap maps in
     (qm, inner)
  | Ast.T.ExistentialInnerMost vars ->
     let treat_vmap inn vmap =
       let new_inner = search_exists_inner vmap vars in
       List.rev_append new_inner inn in
     (qmap, List.fold_left treat_vmap inner maps)
(*  let parity = if b then 1 else 0 in
  let update i qm var =
    let f = function | None -> Some [var] | Some l -> Some (var :: l) in
    IMap.update (2 * i + parity) f qm in
  let treat_one_var vmap qm (n, args) =
    let i = expr vmap e in
    let vars : search_var list = List.map (fun args -> (n, args)) (tuple' vmap args) in
    List.fold_left (update i) qm vars in
  let treat_one qm vmap =
    List.fold_left (treat_one_var vmap) qm vars in
  List.fold_left treat_one qmap maps*)

let all_search gmap decls =
  let (qmap, inner) = List.fold_left (search_decl gmap) (IMap.empty, []) decls in
  let inner = if inner <> [] then [(true, inner)] else [] in
  if IMap.is_empty qmap then inner
  else
    let blocks = IMap.bindings qmap in
    let f (i, l) = (i mod 2 = 1, l) in
    let prefix = List.rev_map f blocks in
    List.rev_append prefix inner

let all_hide gmap decls : (bool * T.literal) list =
  let hide_decl (gls, ((hide, lits) : Ast.T.hide_decl)) =
    let maps = glits gmap SMap.empty gls in
    let all_lits = List.concat_map (search_lits lits) maps in
    List.map (fun lit -> hide, lit) all_lits in
  List.concat_map hide_decl decls

let literals gmap vmap (gls, pol, ga) =
  let maps = glits gmap vmap gls in
  List.map (fun m -> (pol, search_var ga m)) maps

let clause_decl gmap (gls, (hyps, ccls)) =
  let maps = glits gmap SMap.empty gls in
  let make_clause vmap =
    let hyps = List.concat_map (literals gmap vmap) hyps in
    let ccls = List.concat_map (literals gmap vmap) ccls in
    (hyps, ccls) in
  Misc.map make_clause maps

let all_clause gmap decls = List.concat_map (clause_decl gmap) decls

let print_facts facts gmap =
  let pr_one (key, set) =
    let pr_tuple tuple = Print.ground_fact (key, tuple) ^ "." in
    let elements = GTermSet.elements set in
    if elements = [] then sprintf "%s no fact of the form `%s[...]` was generated." "%%%" key
    else P.unspaces pr_tuple elements in
  if facts then eprintf "%s\n%!" (P.unlines pr_one (SMap.bindings gmap))

let file facts (decls : Ast.T.file) : T.file =
  let { Ast.T.ground; prefix; matrix; hide } = decls in
  let gmap = all_ground ground in
  print_facts facts gmap;
  let prefix = all_search gmap prefix in
  let matrix = all_clause gmap matrix in
  let hide_st = all_hide gmap hide in
  let show, hide = List.partition_map (fun (h, lit) -> if h then Either.Right lit else Either.Left lit) hide_st in
  { prefix;
    matrix;
    hide;
    show }

let _ = ignore (print_smap, print_gtset, unions)

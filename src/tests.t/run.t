  $ ../bin/Bule.exe --facts iss26_badMatch.bul 2> /dev/null
   colored(va) ->  colored1(color(va,red)).
   colored(vb) ->  colored1(color(vb,red)).
  $ ../bin/Bule.exe --facts iss30_badMatch.bul 2> /dev/null
   m(a).
   m(b).
  $ ../bin/Bule.exe --facts iss45_groundingPerf.bul 2> /dev/null
  $ ../bin/Bule.exe --solve iss49_duplicateVars.bul 2> /dev/null
  UNSAT
  $ ../bin/Bule.exe --mode ground --output bule iss52_range_in_defs.bul | diff iss52_range_in_defs.target.bul -
  $ ../bin/Bule.exe iss54_negativeBlock.bul 2> /dev/null
  #exists[0] a.
  $ ../bin/Bule.exe --mode solve --solver "depqbf --no-dynamic-nenofex --qdo" iss59_display_counterexample.bul | diff iss59_display_counterexample.target.out -
  $ ../bin/Bule.exe --mode ground --output bule iss61_shared_grounding_prefix.bul | diff iss61_shared_grounding_prefix.target.bul -

  $ ../bin/Bule.exe --facts iss26_badMatch.bul 2> /dev/null
   colored(va) ->  colored1(color(va,red)).
   colored(vb) ->  colored1(color(vb,red)).
  $ ../bin/Bule.exe --facts iss30_badMatch.bul 2> /dev/null
   m(a).
   m(b).
  $ ../bin/Bule.exe --facts iss45_groundingPerf.bul 2> /dev/null
  $ ../bin/Bule.exe --solve iss49_duplicateVars.bul 2> /dev/null
  UNSAT
  $ ../bin/Bule.exe iss54_negativeBlock.bul 2> /dev/null
  #exists[0] a.

#!/bin/zsh

timelimit=1000

cd minisat
for x in *.pb
do 
    echo minisat $x
    log=$(basename $x .lp).log
    timeout $timelimit minisat+ $x > $log
    echo 
done 
cd ..

#cd gurobi
#for x in *.lp
#do 
#    echo gurobi $x
#    log=$(basename $x .lp).log
#    sol=$(basename $x .lp).sol
#    gurobi_cl ResultFile=$sol TimeLimit=$timelimit FeasibilityTol=1e-09 IntFeasTol=1e-09 MIPGap=0 MIPGapAbs=0 $x > $log
#    echo 
#done 
#
#cd ..
#cd bule_sat
#for x in *.cnf
#do 
#    echo bule $x
#    log=$(basename $x .cnf).log
#    clasp $x -t 4 --time-limit=$timelimit --stat -q --configuration=chatty > $log
#done 
#
#cd ..
#cd sugar_sat
#for x in *.cnf
#do 
#    echo sugar $x
#    log=$(basename $x .cnf).log
#    clasp $x -t 4 --time-limit=$timelimit --stat -q --configuration=chatty > $log
#done 
#
#cd ..
#cd wbo
#for x in *.pb
#do 
#    echo wbo $x
#    log=$(basename $x .pb).log
#    wbo -file-format=opb $x > $log
#done 
#
#
#cd ..
#cd clasp
#for x in *.pb
#do 
#    echo clasp $x
#    log=$(basename $x .pb).log
#    clasp $x -t 4 --time-limit=$timelimit --stat -q --configuration=chatty > $log
#done 
#
#




#for x in data/hard\ knapsack/*.cnf
#do 
#    echo $x
#    #clasp $x -t 4 --time-limit=3600 --stat -q --configuration=chatty
#    timeout 600 plingeling $x
#    echo 
#    echo 
#done 
#
#for x in data/benchs1/*.pb
#do 
#    echo $x
#    #clasp $x -t 4 --time-limit=3600 --stat -q --configuration=chatty
#    #timeout 7200 plingeling $x
#    timeout 900 pbsugar $x -solver 'clasp --stat --configuration=chatty -t 4' -v -v
#    echo 
#    echo 
#done 

#for x in data/benchs*/*.pb
#do 
#    echo
#    echo $(basename $x .pb).cnf
#    ./bule -gurobi -f $x > gurobi/$(basename $x .pb).lp
#    #timeout 600 pbsugar $x -n -vv -sat sugar/$(basename $x .pb).cnf
#    #timeout 600 go run bule.go -f $x -o bule_sat/$(basename $x .pb).cnf
#done 

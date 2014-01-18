#!/bin/zsh

#for x in data/hard\ knapsack/*.cnf
#do 
#    echo $x
#    #clasp $x -t 4 --time-limit=3600 --stat -q --configuration=chatty
#    timeout 600 plingeling $x
#    echo 
#    echo 
#done 
#
for x in data/benchs2/*.cnf
do 
    echo $x
    #clasp $x -t 4 --time-limit=3600 --stat -q --configuration=chatty
    timeout 7200 plingeling $x
    echo 
    echo 
done 

#for x in data/benchs2/*.pb
#do 
#    echo $x
#    go run bule.go -f $x > /dev/null
#done 

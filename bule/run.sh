#!/bin/zsh

for x in data/hard\ knapsack/*.cnf
do 
    echo $x
    #clasp $x -t 4 --time-limit=3600 --stat -q --configuration=chatty
    timeout 600 plingeling $x
    echo 
    echo 
done 

for x in data/japan/*.cnf
do 
    echo $x
    #clasp $x -t 4 --time-limit=3600 --stat -q --configuration=chatty
    timeout 600 plingeling $x
    echo 
    echo 
done 

#for x in data/benchs1/*.pb
#do 
#    echo $x
#    go run bule.go -f $x > /dev/null
#done 

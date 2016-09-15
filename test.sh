#!/bin/zsh

go build bule.go

echo GRINGO
for x in instances/*/*; do echo $x $(./bule -f $x -gringo -solve=false | gringo3 | clasp | grep 'SATIS\|Optimization \|OPTIMUM FOUND'); done
echo 
echo BULE with cat 1 and cmsat
for x in instances/*/*; do ./bule -f $x  -cat 2 -solver cmsat | grep instance; done | column -s ';' -t 
echo 
echo BULE with cat 2 and cmsat
for x in instances/*/*; do ./bule -f $x  -cat 1 -solver cmsat | grep instance; done | column -s ';' -t 

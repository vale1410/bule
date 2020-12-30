#!/bin/zsh

make

echo GRINGO
for x in instances/*/*; do echo $x $(./bule pb -f $x --gringo --solve=false | gringo | clasp | grep 'SATIS\|Optimization \|OPTIMUM FOUND'); done
echo 
echo BULE with cat 1 and cmsat
for x in instances/*/*; 
do 
    ./bule pb -f $x --cat 1 --solver clasp | grep instance; done | column -s ';' -t 
    ./bule pb -f $x --cat 1 --solver clasp | grep instance; done | column -s ';' -t 
echo 
echo BULE with cat 2 and cmsat
for x in instances/*/*; do ./bule pb -f $x --cat 2 --solver clasp| grep instance; done | column -s ';' -t 

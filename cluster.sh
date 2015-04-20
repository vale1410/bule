#!/bin/zsh


#go build; 
results=results
rm -fr $results
mkdir -p $results

for timeout in 300 900 3600
do 
    for seed in  42 132 6534 7654 3456 
    do 

        for amo in 0 1 
        do
            for x in $1/*.pb; 
            do 
            echo ./bule -d -amo-chain=$amo -timeout $timeout -f $x -seed $seed " > "$results/$(basename $x .pb)-amo-$amo-$seed-$timeout.log
            done 
        done
        
        for x in $1/*.pb
        do 
            echo ./bule -f $x -gringo '| gringo | clasp --time-limit' $timeout --seed $seed' > '$results/$(basename $x .pb)-clasp-$seed-$timeout.log
        done 
    done 
done 
    
    

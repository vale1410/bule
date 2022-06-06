#!/bin/zsh


for x in $(seq 1 16) 
do 
    for y in $(seq 1 16) 
    do 
        cat gen_reachability_test.bul | sed 's/XXX/'$x'/g' | sed 's/YYY/'$y'/g' > tmp.bul
        echo $x $y $(bule index.bul implicit_reachability.bul tmp.bul --solve --solver "depqbf --no-dynamic-nenofex --qdo" 2>/dev/null | grep SAT )
    done
done


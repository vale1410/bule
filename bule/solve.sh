#/bin/zsh


timeout=20000

for x in opt/*
do  
    echo $x
    ./bule -solve -f $x -cat 1 -timeout $timeout -solver=minisat -opt-bound 46
    for opt in 0 1; 
    do 
        for amo in 0 1; 
        do 
            ./bule -solve -f $x -cat 2 -opt-rewrite=$opt -timeout $timeout -solver=minisat -opt-bound 46 -amo-reuse=$amo
        done
    done
done


#echo hybrid
#for x in $1/*.*pb; 
#do 
#    time -f';%e' timeout 1000 ./bule -f $x -complex=hybrid -solve
#done

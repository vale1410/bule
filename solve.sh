#/bin/zsh


timeout=1800

go build; 

echo cat 1
for x in $1/*.pb; 
do 
    ./bule  -cat 1 -solve -timeout $timeout -solver=minisat -f $x | grep xxx; 
done 

for amo in 0 1
do 
    for opt in 0 1 
    do echo opt cat 2 opt $opt amo $amo
        for x in $1/*.pb; 
        do ./bule  -cat 2 -solve -amo-reuse=$amo -opt-rewrite=$opt -timeout $timeout -solver=minisat -f $x | grep xxx; 
        done 
    done 
done


#for x in opt/*
#do  
#    echo $x
#    ./bule -solve -f $x -cat 1 -timeout $timeout -solver=minisat -opt-bound 46
#    for opt in 0 1; 
#    do 
#        for amo in 0 1; 
#        do 
#            ./bule -solve -f $x -cat 2 -opt-rewrite=$opt -timeout $timeout -solver=minisat -opt-bound 46 -amo-reuse=$amo
#        done
#    done
#done


#echo hybrid
#for x in $1/*.*pb; 
#do 
#    time -f';%e' timeout 1000 ./bule -f $x -complex=hybrid -solve
#done

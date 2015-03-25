#/bin/zsh

go build bule.go

echo cat 1 ";true ;"
for x in $1/*.*pb; 
do 
    time -f';%e' timeout 1000 ./bule -f $x -cat 1 -solve -timeout 20 -mdd_redundant=true
done

echo cat 2";true ;" 
for x in $1/*.*pb; 
do 
    time -f';%e' timeout 1000 ./bule -f $x -cat 2 -solve -timeout 20 -mdd_redundant=true 
done

echo cat 1 ";false ;"
for x in $1/*.*pb; 
do 
    time -f';%e' timeout 1000 ./bule -f $x -cat 1 -solve -timeout 20 -mdd_redundant=false
done

echo cat 2";false ;" 
for x in $1/*.*pb; 
do 
    time -f';%e' timeout 1000 ./bule -f $x -cat 2 -solve -timeout 20 -mdd_redundant=false
done

#echo hybrid
#for x in $1/*.*pb; 
#do 
#    time -f';%e' timeout 1000 ./bule -f $x -complex=hybrid -solve
#done

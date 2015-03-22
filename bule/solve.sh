#/bin/zsh

go build bule.go

echo mdd
for x in $1/*.*pb; 
do 
    time -f';%e' timeout 1000 ./bule -f $x -complex=mdd -mdd_redundant=false -solve
done

#echo mdd + redundant
#for x in $1/*.*pb; 
#do 
#    time -f';%e' timeout 1000 ./bule -f $x -complex=mdd -mdd_redundant=true
#done

echo sn
for x in $1/*.*pb; 
do 
    time -f';%e' timeout 1000 ./bule -f $x -complex=sn -solve
done

echo hybrid
for x in $1/*.*pb; 
do 
    time -f';%e' timeout 1000 ./bule -f $x -complex=hybrid -solve
done

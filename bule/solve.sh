#/bin/zsh

go build bule.go

echo mdd
for x in $1/*.*pb; 
do 
    time -f';%e' timeout 1000 ./bule -f $x -complex=mdd
done

echo sn
for x in $1/*.*pb; 
do 
    time -f';%e' timeout 1000 ./bule -f $x -complex=sn
done

echo hybrid
for x in $1/*.*pb; 
do 
    time -f';%e' timeout 1000 ./bule -f $x -complex=hybrid
done

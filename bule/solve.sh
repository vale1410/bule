#/bin/zsh

go build bule.go

for x in $1/*.*pb; 
do 
    time -f';%e' timeout 600 ./bule -f $x; 
    echo ;  
done

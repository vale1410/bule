#!/bin/zsh

go build main.go
if [ $? -ne 0 ] ; then 
    echo Could not build Bule
    return 
fi 

mkdir -p test-output

for x in test-input/*.bul
do 
    name=$(basename $x)
    ./main ground -f $x | sort  > test-output/$name
    diff -B -b test-expected/$name test-output/$name > /dev/null
    if [ $? -ne 0 ] ; then 
        echo test failed: $name
        diff -B -b test-expected/$name test-output/$name 
    else 
        echo  ☑ $name 
    fi 

done 


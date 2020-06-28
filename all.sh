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
    ./main ground  -t=1 -u=0 -i=1 -q=0 --facts $x | sed 's/%.*//g' | sed '/^\s*$/d' | sort > test-output/$name
    diff -B -b test-expected/$name test-output/$name > /dev/null
    if [ $? -ne 0 ] ; then 
        echo test failed: $name
        echo input: 
        cat test-input/$name
        echo output: 
        cat test-output/$name
        echo expected: 
        cat test-expected/$name
    else 
        echo  ☑ $name 
    fi 
done 

echo TODO: 
for x in test-todo/*.bul
do 
    echo $(basename $x)
done 

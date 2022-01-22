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
    ./main ground  -t=1 -q=0 --facts $x 2>/dev/null | sed 's/%.*//g' | sed '/^\s*$/d' | sort > test-output/$name 
    diff -B -b test-expected/$name test-output/$name > /dev/null
    if [ $? -ne 0 ] ; then 
        echo ---------
        echo test failed: $name
        echo input: 
        echo ---------
        cat test-input/$name
        echo ---------
        echo output: 
        echo ---------
        cat test-output/$name
        echo ---------
        echo expected: 
        echo ---------
        cat test-expected/$name
        echo ---------
    else 
        echo  ☑ $name 
    fi 
done 

echo TODO: 
for x in test-todo/*.bul
do 
    echo $(basename $x)
done 

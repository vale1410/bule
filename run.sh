#!/bin/zsh

mkdir -p results
s=$1
folder=${s:6}
rm -fr results/$folder
mkdir  results/$folder

mkdir -p pics
rm -fr pics/$folder
mkdir  pics/$folder

for x in $1/*.dot
do 
    name=$(basename $x .dot)
#    echo - $name 
    ./workflow -f $x > results/$folder/$name.dot
    diff tests-expected/$folder/$name.dot results/$folder/$name.dot > /dev/null
    if [ $? -ne 0 ] ; then
        echo test failed: $name.dot
        dot -Tpng tests-expected/$folder/$name.dot -o pics/$folder/$name-exp.png
    fi 
        dot -Tpng $x -o pics/$folder/$name-in.png
        dot -Tpng results/$folder/$name.dot -o pics/$folder/$name-res.png
done 


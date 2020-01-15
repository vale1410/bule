#!/bin/zsh

go build workflow.go

for x in tests/*
do 
    echo running tests $(basename $x)
    ./run.sh $x
done 

echo "the compiled images <name>-in.png and <name>-res.png can be found in folder pics"

rm workflow

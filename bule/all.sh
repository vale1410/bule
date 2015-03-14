#/bin/zsh

go build bule.go

#echo "name;Facts;Clause;AMO;Ex1;Card;BDD;SN"

for x in **/*.*pb; do echo $x; done
for x in **/**/*.*pb; do echo $x; done
for x in **/**/**/**/*.*pb; do echo $x; done
#for x in **/*.opb; do echo $x; done
#timeout 30s ./bule -f $x -solve=false $x


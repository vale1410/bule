#/bin/zsh

go build bule.go

echo "Inst;vars;cons;facts;clause;am1;ex1;card;BDD;SN"

for x in **/**/**/*.*pb; do timeout 30s ./bule -f $x -solve=false $x; done
#for x in **/**/*.*pb; do echo $x; done
#for x in **/**/**/**/*.*pb; do echo $x; done
#for x in **/*.opb; do echo $x; done
#timeout 30s ./bule -f $x -solve=false $x

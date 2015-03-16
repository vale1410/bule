#/bin/zsh

go build bule.go

for x in **/**/*.*pb; do echo $x; timeout 600 ./bule -f $x; echo ;  done

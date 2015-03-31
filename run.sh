#/bin/zsh

go run bule.go -f $1
clasp --stat out.cnf $2

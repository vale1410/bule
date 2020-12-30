#!/bin/zsh

RE=release-$(git describe --tags)
mkdir $RE

env GOOS=linux GOARCH=amd64 go build main.go
mv main $RE/bule_x64
env GOOS=darwin GOARCH=amd64 go build main.go
mv main $RE/bule_mac64
env GOOS=windows GOARCH=amd64 go build main.go
mv main.exe $RE/bule_win64.exe

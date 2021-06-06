#!/bin/zsh

TAG=$(git describe --tags)

echo preparing a release with tag $TAG

RE=release-$TAG
mkdir $RE

### insert git tag in cmd/root.go ; it is set back afterwards
mv cmd/root.go $RE/root.go
DATES=$(date "+%Y-%m-%d") 
cat $RE/root.go | sed s/{{{VERSION}}}/$TAG/g | sed s/{{{DATE}}}/$DATES/g > cmd/root.go

env GOOS=linux GOARCH=amd64 go build main.go
mv main $RE/bule_x64
env GOOS=darwin GOARCH=amd64 go build main.go
mv main $RE/bule_mac64
env GOOS=windows GOARCH=amd64 go build main.go
mv main.exe $RE/bule_win64.exe

### Put version back
mv $RE/root.go cmd/root.go 

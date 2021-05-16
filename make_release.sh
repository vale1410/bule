#!/bin/zsh

TAG=$(git describe --tags)

echo preparing a release with tag $TAG

RE=release-$TAG
mkdir $RE

### insert git tag in cmd/version.go ; it is set back afterwards
mv cmd/version.go $RE/version.go
DATES=$(date "+%Y-%m-%d") 
cat $RE/version.go | sed s/{{{VERSION}}}/$TAG/g | sed s/{{{DATE}}}/$DATES/g > cmd/version.go

env GOOS=linux GOARCH=amd64 go build main.go
mv main $RE/bule_x64
env GOOS=darwin GOARCH=amd64 go build main.go
mv main $RE/bule_mac64
env GOOS=windows GOARCH=amd64 go build main.go
mv main.exe $RE/bule_win64.exe

### Put version back
mv $RE/version.go cmd/version.go 

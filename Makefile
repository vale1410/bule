

build:
	go build main.go
	mv main bule 

test: build
	./all.sh	

release: build
	rm -fr release
	mkdir release
	env GOOS=linux GOARCH=amd64 go build main.go
	mv main release/bule_x64
	env GOOS=darwin GOARCH=amd64 go build main.go
	mv main release/bule_mac64
	env GOOS=windows GOARCH=amd64 go build main.go
	mv main.exe release/bule_win64.exe

clean: 
	rm -fr test-output
	rm -fr main
	rm -fr release

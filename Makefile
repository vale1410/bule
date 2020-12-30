

build:
	go build main.go
	mv main bule 

test: build
	./all.sh	

release: build
	./make_release.sh

clean: 
	rm -fr test-output
	rm -fr main
	rm -fr release-*

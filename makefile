build:
	go build main.go
	mv main bule 

test: build
	./run.sh	

clean: 
	rm -fr test-output
	rm -fr main


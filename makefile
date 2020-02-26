build:
	go build main.go

test: build
	./run.sh	

clean: 
	rm -fr test-output
	rm -fr main


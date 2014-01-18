NAME=test

all: compile

compile:
	mkdir -p tmp
	go test
	pdflatex $(NAME).tex
	evince $(NAME).pdf 2>/dev/null &

clean: 
	rm -fr tmp

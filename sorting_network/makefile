NAME=test

all: compile

compile:
	mkdir -p tmp
	go test
	pdflatex $(NAME).tex
	evince $(NAME).pdf 2>/dev/null &

clean: 
	rm -fr tmp
	rm -fr *.log
	rm -fr *.aux
	rm -fr *.pdf
	rm -fr *.tex

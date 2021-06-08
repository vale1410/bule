
NAME=bule
OUTPUT=$(NAME).pdf

all: compile

compile:
	latexmk -pdf $(NAME).tex 
#	bibtex $(NAME)
#	pdflatex $(NAME).tex 
#	pdflatex $(NAME).tex 

show: compile
	evince $(OUTPUT) 2>/dev/null &

clean: 
	rm -fr $(NAME).fdb_latexmk
	rm -fr $(NAME).fls
	rm -fr $(NAME).toc
	rm -fr $(NAME).pdf
	rm -fr $(NAME).blg
	rm -fr $(NAME).bbl
	rm -fr $(NAME).out
	rm -fr $(NAME).aux
	rm -fr $(NAME).log
	rm -fr *.aux

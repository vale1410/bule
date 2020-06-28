
NAME=bule
OUTPUT=$(NAME).pdf

all: compile

compile:
	pdflatex $(NAME).tex 
	bibtex $(NAME)
#	pdflatex $(NAME).tex 
#	pdflatex $(NAME).tex 

show: compile
	evince $(OUTPUT) 2>/dev/null &

clean: 
	rm -fr $(NAME).pdf
	rm -fr $(NAME).blg
	rm -fr $(NAME).bbl
	rm -fr $(NAME).out
	rm -fr $(NAME).aux
	rm -fr $(NAME).log
	rm -fr *.aux

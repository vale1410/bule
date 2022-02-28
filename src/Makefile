all: bule

.PHONY: clean dlpag

clean:
	rm -f *~ -r __pycache__
	dune clean

bule:
	dune build bin/Bule.exe

install:
	dune build @install
	dune install

test:
	dune runtest

uninstall:
	dune uninstall


errorfile:
	menhir --list-errors lib/Parser.mly > lib/Parser.messages

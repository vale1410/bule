all: dlpag

.PHONY: clean dlpag

clean:
	rm -f *~ -r __pycache__
	dune clean

dlpag:
	dune build bin/Bule2.exe

install:
	dune build @install
	dune install

test:
	dune runtest

uninstall:
	dune uninstall


errorfile:
	menhir --list-errors lib/Parser.mly > lib/Parser.messages

* To compile run
* `opam install dune menhir minisat qbf tsort`
* `make install`


If there are problems of the type: 
```
<><> Error report <><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><>
┌─ The following actions failed
│ λ build qbf 0.3
└─ 
╶─ No changes have been performed
```

C++ compiler should be set to gcc with (does not work with CLION) : 
```
export CC=gcc
```

If there are problems of the type: 
`
^^^^^^^^^^^^^^^
Error: Unbound value List.concat_map
`

Then check the version of opam: 
```
opam --version 
```

if too old, then update to (for example 4.13.1)

```
opam switch create 4.13.1
```

update terminal: 

```
eval $(opam env)
opam update
opam upgrade
```

and check again the version. 



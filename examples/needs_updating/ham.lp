

node(X) :- edge(X,_).  
node(X) :- edge(_,X).  

{ connect(X,Y) } :- edge(X,Y). 

reached(Y) :- connect(X,Y), reached(X). 

reached(1). 
:- node(X), not reached(X). 

:- connect(X,Y1), connect(X,Y2), Y1 < Y2. 
:- connect(Y1,X), connect(Y2,X), Y1 < Y2. 

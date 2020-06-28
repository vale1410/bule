#show p/1. 

p(1..4).
{q(X,Y)}  :- p(X), p(Y). 
:- not q(X,X), p(X).

%:- q(X), q(X+1). 

%:- not q(X), not q(Y), X != Y. 
%:- not q(X), not q(Y), X != Y, p(X), p(Y). 

%:- not q(X), not q(X+1). % doesnt work, X is not bound !

%:- not q(X), not q(X+3), p(X;X+3). % works but needs explicit bound. 

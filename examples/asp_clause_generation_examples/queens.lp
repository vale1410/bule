% THE n-QUEENS PROBLEM
%
% Intuitive readings for predicates:
%
% coord(X)        = X is a valid coordinate value
% dir(R,C)        = Define coordinate differences for adjacent cells
% target(X,Y,R,C) = (X,Y) and (X+R,Y+C) are adjacent
% queen(X,Y)      = There is a queen at position (X,Y)
% attack(X,Y,R,C) = Position (X,Y) is attacked from direction (-R,-C)

#const n = 4.

%%% DOMAIN RULES %%%

coord(1..n).

dir(0,-1). dir(-1,-1). dir(-1,0). dir(-1,1).

target(X,Y,R,C) :- coord(X;Y;X+R;Y+C), dir(R,C).

%%% CLAUSES DEFINING SOLUTIONS %%%

attack(X+R,Y+C,R,C) | -queen(X,Y)      :- target(X,Y,R,C).
attack(X+R,Y+C,R,C) | -attack(X,Y,R,C) :- target(X,Y,R,C), target(X-R,Y-C,R,C).

-attack(X+R,Y+C,R,C) | queen(X,Y) |
 attack(X,Y,R,C) : target(X-R,Y-C,R,C) :- target(X,Y,R,C).

-queen(X+R,Y+C) | -attack(X+R,Y+C,R,C) :- target(X,Y,R,C).

queen(X,1) | attack(X,1,0,-1) : target(X,2,0,-1) :- coord(X).
queen(1,Y) | attack(1,Y,-1,0) : target(2,Y,-1,0) :- coord(Y).

%%% PROJECT SOLUTIONS %%%

#hide.
#show queen/2.
#show -queen/2.
#show attack/4.
#show -attack/4.

% GRAPH COLORING
%
% Intuitive readings for predicates:
%
% edge(X,Y)  = There is a directed edge from X to Y
% node(X)    = X is a node
% color(X,C) = Node X is colored with color C

#const n = 3.

%%% DOMAIN RULES %%%

node(X) :- edge(X,Y).
node(Y) :- edge(X,Y).

%%% CLAUSES DEFINING SOLUTIONS %%%

color(X,C) : C = 1..n :- node(X).

-color(X,C) | -color(Y,C) :- edge(X,Y), C = 1..n.

%%% PROJECT SOLUTIONS %%%

#hide.
#show color/2.
#show -color/2.

%a move(X,Y):X:Y.
%a board(X,Y,T):X:Y.
%% columns c
%% row r
%% connect q
%% 
c=3.
r=3.
q=4.
%K=(c+1)/2.
T={1..c*r}.  
X={1..c}.
Y={1..r}.
D={1..q-1}.

%% WW = {{(X,Y+D):D},{(X+D,Y):D},{(X+D,Y+D):D},{(X-D,Y+D):D}}.
e count(X,Y,T):X:Y.
a abc(1,2).
a move(X,T):X.
a abc(3,4).

%% 
%% exist(T), T #mod 2 = 1, all else
%% exist(T), move(c,T). %% existential player chooses last column. 

%%  % time managment
%~step(c,T), step(c,T-1).
%%  ~step(X-1,T), step(X,T).
%%  ~step(c,T), ~move(X,T), step(X,T).
%%  move(X,T), ~step(X,T), step(X-1,T).
%%  ~step(0,T).
%%  first(X,T), step(X-1,T), ~step(X,T).
%%  ~first(X,T), ~step(X-1,T).
%%  ~first(X,T), step(X,T).
%%  ~count(X,Y,T-1), count(X,Y,T).
%%  count(X,Y-1,T-1), ~count(X,Y,T).
%%  ~first(X,T), ~count(X,Y-1,T-1), count(X,Y,T).
%%  first(X,T), count(X,Y,T-1), ~count(X,Y,T).
%%  count(X,0,0).
%%  ~count(X,1,0).
%%  board(X,Y,T#mod2), count(X,Y,T-1), ~count(X,Y,T).
%%  ~board(X,Y,T#mod2), ~count(X,Y,T-1).
%%  ~board(X,Y,T#mod2), count(X,Y,T).
%%  ~board(X,Y,0), ~board(X,Y,1).

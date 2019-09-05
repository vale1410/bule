%% columns c
%% row r
%% connect q
%% 
c=3.
r=3.
q=4.
K=(c+1)/2.
T={1..c*r}.  
X={1..c}.
Y={1..r}.
D={0,q-1}.
P={0,1}.

%% WW = {{(X,Y+D):D},{(X+D,Y):D},{(X+D,Y+D):D},{(X-D,Y+D):D}}.

a move(X,T):X.
e move(X,T):X.
e step(X,T):X. 
e first(X,T):X.
e count(X,Y,T):X:Y.
e board(X,Y,P):X:Y:P.

% time managment
~step(c,T), step(c,T-1).
% step(C,T) is true if at least one move has been made. 
% this rules enforces that if a move has been made, then also in previous times moves 
% have been made. And once no moves are mode, none moves will be made in the future. 

% determine the essential move by counting. If multiple, then choose the first one in order of columns.
~step(X-1,T), step(X,T).
~step(c,T), ~move(X,T), step(X,T).
move(X,T), ~step(X,T), step(X-1,T).
~step(0,T).

%%symmetrie breaking, choose first move in the right side (including center).
%step(K,1).

%first(X,T) <=> ~step(X-1,T) && step(X,T).
first(X,T), step(X-1,T), ~step(X,T).
~first(X,T), ~step(X-1,T).
~first(X,T), step(X,T).

% counter encoding of the board positions at certain time. count(X,Y,T) is true if at time T cell (X,Y) is occupied. 
~count(X,Y,T-1), count(X,Y,T).
count(X,Y-1,T-1), ~count(X,Y,T).
~first(X,T), ~count(X,Y-1,T-1), count(X,Y,T).
first(X,T), count(X,Y,T-1), ~count(X,Y,T).

count(X,0,0).
~count(X,1,0).

%board(X,Y,T #mod 2) <=> ~count(X,Y,T-1) && count(X,Y,T).
board(X,Y,T#mod2), count(X,Y,T-1), ~count(X,Y,T).
~board(X,Y,T#mod2), ~count(X,Y,T-1).
~board(X,Y,T#mod2), count(X,Y,T).

~board(X,Y,0), ~board(X,Y,1).

%%% Generate by hand 
%%%  %win(P,W) <=> board(X,Y,P) : (X,Y) in W.
%% ~board(I,J,1):(I,J) in W. 
%% win(W) : W. 
%% ~win(W), board(I,J,0):(I,J) in W.  

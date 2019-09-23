%% columns c
%% row r
%% connect q

c=3.
r=3.
q=4.

T={1..c*r-((c*r)#mod2)}.
X={1..c}.
Y={1..r}.
D={0..q-1}.

#exist(T), move(X,T):X, T#mod2==1.
#forall(T), move(X,T):X, T#mod2==0.
#exist(T), step(X,T):X, first(X,T):X, count(X,Y,T):X:Y.

% time management
~step(c,T), step(c,T-1).

% step(c,T) is true if at least one move has been made.
% this rules enforces that if a move has been made, then also in previous times moves 
% have been made. And once no moves are mode, none moves will be made in the future. 

% determine the essential move by counting. If multiple, then choose the first one in order of columns.
~step(X-1,T), step(X,T).
~step(c,T), ~move(X,T), step(X,T).
move(X,T), ~step(X,T), step(X-1,T).
~step(0,T).

%% symmetry breaking, choose first move in the right side (including center).
%step(q/2,1).

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

%%%  %win(P,W) <=> board(X,Y,P) : (X,Y) in W.
P={0..1}.
Z={0..1}.

~board(X+Z*D,Y+(1-Z)*D,1):D, X+Z*(q-1)<=c, Y+(1-Z)*(q-1)<=r.
 board(X+Z*D,Y+(1-Z)*D,1),   X+Z*(q-1)<=c, Y+(1-Z)*(q-1)<=r, ~winstraight(X,Y,Z).
~board(X+Z*D,Y+Z*(q-1)+((-1)**Z)*D,1):D, X+q-1<=c, Y+q-1<=r.
 board(X+Z*D,Y+Z*(q-1)+((-1)**Z)*D,1),   X+q-1<=c, Y+q-1<=r, ~windiagonal(X,Y,Z).

winstraight(X,Y,Z):X+Z*(q-1)<=c:Y+(1-Z)*(q-1)<=r:X:Y:Z.

%, win(X,Y,1,Z):X+q-1<=c:Y+q-1<=r:X:Y:Z.


%% columns c
%% row r
%% connect q

c=3.
r=3.
q=3.

%% board positions (X,Y)
X={1..c}.
Y={1..r}.

%% Number of moves, alternating between player 0 (even T) and player 1 (odd T).
%% Last move should be by player 0, if he can't by then, game is draw or player 1 wins.
T={0..c*r-((c*r)#mod2)-1}.

#exist[T], move[X,T]:X, T#mod2==0.
#forall[T], move[X,T]:X, T#mod2==1.
#exist[T], step[X,T]:X, first[X,T]:X, count[X,Y,T]:X:Y.

%% All remaining variables are innermost layer (default)

% time management
% step[c,T] is true if at least one move has been made.
% This rules enforces that if a move has been made, then also in previous times moves
% have been made. And once no moves are mode, no moves will be made in the future.

~step[c,T], step[c,T-1], T>1.

% determine the essential move by counting.
% If multiple, then choose the first one in order of columns.

~step[X-1,T], step[X,T].
~step[c,T], ~move[X,T], step[X,T].
 move[X,T], ~step[X,T], step[X-1,T].
~step[0,T].

%% symmetry breaking, choose first move in the right side (including center).
%% q/2 rounds down, i.e. it works for odd and even q
% ~step[q/2,0].

%first[X,T] <=> ~step[X-1,T] && step[X,T].

first[X,T], step[X-1,T], ~step[X,T].
~first[X,T], ~step[X-1,T].
~first[X,T], step[X,T].

%% counter encoding of the board positions at certain time.
%% count[X,Y,T] is true if at time T cell [X,Y] is occupied.
~count[X,Y,T-1], count[X,Y,T].
count[X,Y-1,T-1], ~count[X,Y,T].
~first[X,T], ~count[X,Y-1,T-1], count[X,Y,T].
first[X,T], count[X,Y,T-1], ~count[X,Y,T].

 count[X,0,-1].
 count[X,Y,-1].
~count[X,1,-1].

%board[X,Y,T#mod2] <=> ~count[X,Y,T-1], count[X,Y,T].
 board[X,Y,T#mod2], count[X,Y,T-1], ~count[X,Y,T].
~board[X,Y,T#mod2],~count[X,Y,T-1].
~board[X,Y,T#mod2], count[X,Y,T].


~board(X,Y,0), ~board(X,Y,1).

%%% Winning positions
D={0..q-1}.
Z={0..1}.

% horizontal and vertical winning positions (winstr)
~board(X+Z*D,Y+(1-Z)*D,0):D, X+Z*(q-1)<=c, Y+(1-Z)*(q-1)<=r.
 board(X+Z*D,Y+(1-Z)*D,1),   X+Z*(q-1)<=c, Y+(1-Z)*(q-1)<=r, ~winstr(X,Y,Z).

% diagonal winning positions (windia)
~board(X+D,Y+Z*(q-1)+((-1)**Z)*D,0):D, X+q<c, Y+q<r.
 board(X+D,Y+Z*(q-1)+((-1)**Z)*D,1),   X+q<c, Y+q<r, ~windia(X,Y,Z).

% at least one winning positions for player 1
winstr(X,Y,Z):X+Z*(q-1)<=c:Y+(1-Z)*(q-1)<=r:X:Y:Z, windia(X,Y,Z):X+q<c:Y+q<r:X:Y:Z.


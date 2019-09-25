
c=2.
r=2.
q=2.

%% board positions (X,Y)
X={1..c}.
Y={1..r}.
T={0..c*r-((c*r)#mod2)-1}.

~time[T], time[T-1].

move[1,T] <=> 1 <= c, time[T],  chose[0,T], ~chose[1,T], ~chose[2,T].
move[2,T] <=> 2 <= c, time[T], ~chose[0,T],  chose[1,T], ~chose[2,T].
move[3,T] <=> 3 <= c, time[T],  chose[0,T],  chose[1,T], ~chose[2,T].
move[4,T] <=> 4 <= c, time[T], ~chose[0,T], ~chose[1,T],  chose[2,T].
move[5,T] <=> 5 <= c, time[T],  chose[0,T], ~chose[1,T],  chose[2,T].
move[6,T] <=> 6 <= c, time[T], ~chose[0,T],  chose[1,T],  chose[2,T].
move[7,T] <=> 7 <= c, time[T],  chose[0,T],  chose[1,T],  chose[2,T].

~count[X,Y,T-1], count[X,Y,T].
count[X,Y-1,T-1], ~count[X,Y,T].
~move[X,T], ~count[X,Y-1,T-1], count[X,Y,T].
move[X,T], count[X,Y,T-1], ~count[X,Y,T].

count[X,0,-1].
count[X,Y,-1].
~count[X,1,-1].

board[X,Y,T#mod2] <=> ~count[X,Y,T-1], count[X,Y,T].

%% winning positions:
P={0..1}.
Z={0..1}.
V={0..1}.
win[X,Y,V,Z,P] <=> board[X+Z*D,V*(Y+(1-Z)*D)+(1-V)*(Y+Z*(q-1)+((-1)**Z)*D),P]:D.

%%  player 0 needs at least one winning positions
win[X,Y,V,Z,0]:X:Y:V:Z.

%% player 1 should not win
~win[X,Y,V,Z,1].

%% Nobody claims positions outside the board
~board[r+1..r+q,Y+D,P].
~board[X+D,c+1..c+q,P].

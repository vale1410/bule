

c=3.
r=3.
q=3.

T={1..c*r}.
X={1..c}.
Y={1..r}.
D={0..q-1}.


P={0..1}.
Z={0..1}.
V={0..1}.

% definition of winning positions
win[X,Y,V,Z,P] <=> board[X+Z*D,V*(Y+(1-Z)*D)+(1-V)*(Y+Z*(q-1)+((-1)**Z)*D),P]:D.

% player 1 should not win
~win[X,Y,V,Z,1].

% player 0 needs at least one winning positions
win[X,Y,V,Z,0]:X:Y:V:Z.

%%% Nobody claims positions outside the board
%%~board[r+1..r+q,Y+D,P].
%%~board[X+D,c+1..c+q,P].



%%  #exist(T), move(X,T):X, T#mod2==1.
%%  #forall(T), T#mod2==0, move(X,T):X.
%%
%%  hello(T):T#mod2==0:T>3:T.
%%
%%  hello2(T), T#mod2==0, T>3.
%%
%%  %zzz(X,Y):X:Y.

%%  xxxx(X,M):X==M#mod2:X:M.

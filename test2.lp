

c=3.
r=3.
q=3.

K=(c+1)/2.
T={1..c*r}.
M={1..(c*r)/2}
X={1..c}.
Y={1..r}.
D={0..q-1}.
P={0..1}.



%%  M={1..3}.
#exist(T), move(X,T):X, T#mod2==1.
#forall(T), T#mod2==0, move(X,T):X.

hello(T):T#mod2==0:T>3:T.

hello2(T), T#mod2==0, T>3.

%zzz(X,Y):X:Y.

%xxxx(X,M):X==M#mod2:X:M.

Z={0..1}.
V={0..1}.
P={0..1}.
%%, X+Z*(q-1)<=c, V*(Y+(1-Z)*(q-1))+(1-V)*(Y+Z*(q-1)+((-1)**Z)*(q-1))<=r.
win(X,Y,V,Z,P)<=>board(X+Z*D,V*(Y+(1-Z)*D)+(1-V)*(Y+Z*(q-1)+((-1)**Z)*D)+0,P):D.
~win(X,Y,V,Z,0).
win(X,Y,V,Z,1):X:Y:V:Z.

head(X,Y)<=>white(X), blak(Y).


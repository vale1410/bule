


X={1..4}.
Y={1..4}.
T={1..7}.

D={0..2}.
P={0..1}.
Z={0..1}.
V={0..1}.


%%  M={1..3}.
#exist(T), move(X,T):X, T#mod2==1.
#forall(T), T#mod2==0, move(X,T):X.

hello(T):T#mod2==0:T>3:T.

%zzz(X,Y):X:Y.

%xxxx(X,M):X==M#mod2:X:M.

% board(X+Z*D,Y+(1-Z)*D*V+((-1)**Z)*(1-V)*D,1):D:Z:V.
% ~win(X,Y,V,Z), board(X+Z*D,Y+(1-Z)*D*V+((-1)**Z)*(1-V)*D,0).
% ~win(0),win(X,Y,V,Z):X:Y:V:Z.





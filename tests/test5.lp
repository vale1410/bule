
%% %% IDEA: The enumeration of the logarithmic options could be done like this: 
B={0..#log(c)}. % log rounds up, i.e. log(3) = 2.
X<=c, time[T], chose[B,T]:(X/(2**B))#mod2==0:B, ~chose[B,T]:(X/(2**B))#mod2==1:B <=> move[X,T]

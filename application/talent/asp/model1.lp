

actor(1..N) :- numActors(N)
scene(1..N) :- numScenes(N)
pos(1..N) :- numScenes(N)

first(0).
last(N) := numScenes(N).

1 { schedule(P,S) : pos(P) } 1 :- scene(S). 

:- pos(P), not 1 { schedule(P,S) } 1.

posActor(P,A) :- schedule(P,S), actorInScene(A,S).


{ firstSlot(P,A) }. 

first(P,A) :- posActor(P,A), first(P). 
notFirst(P,A) :- posActor(P,A), first(P). 




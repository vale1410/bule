% valentins model based on hakans
include "globals.mzn"; 

% input
int: numActors;
int: numScenes;
array[Actors] of int: actorPay;
array[Scenes] of int: sceneDuration;
set of int: Actors = 1..numActors;
set of int: Scenes = 1..numScenes;
array[Actors,Scenes] of int: actorInScene;

% choices variables
% array[Scenes] of var Scenes: scene;
array[Scenes] of var Scenes: slot;

% First and last slots where each actor plays
array[Actors] of var int: firstSlot;
array[Actors] of var int: lastSlot;

% Expression for the waiting time for each actor
array[Actors] of var int: actorWait;

% Expression representing the global cost
var int: idleCost = sum(a in Actors) (actorPay[a] * actorWait[a]);

solve :: int_search(slot, first_fail, indomain_min, complete) minimize idleCost;
% solve :: int_search(scene, "first_fail", "indomain_min", "complete") satisfy;
% solve :: labelling_ff minimize idleCost;

int: max_cost = sum(a in Actors, s in Scenes)  (actorPay[a] * sceneDuration[s]);
int: min_cost = sum(a in Actors, s in Scenes where actorInScene[a,s] = 0)  (actorPay[a] * sceneDuration[s]);

constraint
 forall(a in Actors) ( 
    lastSlot[a] >= 0
    /\
    firstSlot[a] >= 0
    /\
    let {
        set of int: aIns = { s | s in Scenes where actorInScene[a, s] = 1}
    } in
        if card(aIns) > 0 then
            minimum(firstSlot[a], [slot[s] | s in aIns])
            /\
            maximum(lastSlot[a], [slot[s] | s in aIns])
        else
            true
        endif
 );

constraint
 forall(a in Actors) ( 
    actorWait[a] >= 0
    /\
    %actorWait[a] = sum(s in Scenes where actorInScene[a, s] = 0)  (
    %     (sceneDuration[s] * 
    %     bool2int(firstSlot[a] <= slot[s] /\ slot[s] <= lastSlot[a]))
    %)
    actorWait[a] = sum(s in Scenes where actorInScene[a, s] = 0)  (
         (sceneDuration[s] * 
         bool2int(firstSlot[a] <= slot[s] /\ slot[s] <= lastSlot[a]))
    )
    %
 );

%constraint
%    idleCost <= 148;

constraint all_different(slot) ;

output [
  "idleCost: ", show(idleCost), "\n",
  "slot: ", show(slot), "\n",
  "max_cost: ", show(max_cost), "\n",
  "min_cost: ", show(min_cost), "\n",
];


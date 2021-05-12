#!/bin/zsh

## q=2, c=1. r=2. d=(c*r+1)/2. %%  > black lose!
## q=2, c=1. r=4. d=(c*r+1)/2. %%  > black lose!
##
## q=2, c=2. r=2. d=(c*r+1)/2. %%  > black win!
## q=2, c=3. r=2. d=(c*r+1)/2. %%  > black win!
## q=2, c=3. r=3. d=(c*r+1)/2. %%  > black win!
##
## q=3, c=3. r=3. d=(c*r+1)/2. %%  > black loose!
## q=3, c=4. r=3. d=(c*r+1)/2. %%  > black win!
## q=3, c=4. r=4. d=(c*r+1)/2. %%  > black win!
## q=3, c=5. r=5. d=(c*r+1)/2. %%  > black win!
##
## q=4, c=4. r=4. d=(c*r+1)/2. %%  > draw !
## q=4, c=6. r=4. d=(c*r+1)/2. %%  > white win!
## q=4, c=6. r=7. d=(c*r+1)/2. %%  > black win!
## q=4, c=7. r=6. d=(c*r+1)/2. %%  > original! black win!

typeset -A input
typeset -A output

input[1]='q=2,c=1,r=2,d=1'
output[1]=UNSAT

input[2]='q=2,c=2,r=2,d=1'
output[2]=UNSAT

input[3]='q=2,c=2,r=2,d=3'
output[3]=SAT

input[4]='q=3,c=3,r=3,d=5'
output[4]=UNSAT

input[5]='q=3,c=3,r=3,d=7'
output[5]=UNSAT

input[6]='q=3,c=3,r=3,d=9'
output[6]=UNSAT


for i in $(seq 1 6)
do 
    r=$(bule ground connect.bul --const=$input[$i] | depqbf)

    if [[ $r == $output[$i] ]]
    then 
        echo $input[$i] ☑
    else 
        echo $input[$i] ☐  output $r expected $output[$i]
    fi
done

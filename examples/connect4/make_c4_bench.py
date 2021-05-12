#!/usr/bin/env python3

import sys

def process_line(moves, target):
    assert target != 0
    firstPlayerWin = target > 0
    if firstPlayerWin:
        neededForAWin = 43-len(moves)-2*target
    else:
        neededForAWin = 43-len(moves)+2*target
    heights = [1] * 7
    startingP = True
    inits = []
    for move in moves:
        inits.append((startingP,move,heights[move-1]))
        heights[move-1] += 1
        startingP = not startingP
    output = "#const q=4.\n#const c=7.\n#const r=6.\n#const d={}.\n".format(neededForAWin)
    startingPWins = (len(moves) % 2 == 0) == firstPlayerWin
    for (b, c, r) in inits:
        p = "winner" if b == startingPWins else "loser"
        output += "init[{},{},{}].\n".format(p, c, r)
    print(output)

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("usage: script.py position score")
    else:
        moves = [int(c) for c in sys.argv[1]]
        target = int(sys.argv[2])
        process_line(moves, target)

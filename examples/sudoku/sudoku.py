#!/usr/bin/env python
import sys
import itertools as it
from typing import Optional


# example usage: cat sudoku_grounded.cnf | sudoku.py
# this script expects a grounded .cnf file passed through STDIN
# also, save a cnf solver solution in a text file named "solution" (same directory as script)


class CNFParser:

    def __init__(self):
        self.all_literals = {}

    @staticmethod
    def is_cnf_comment(line):
        return line.startswith('c')

    def read_cnf(self):
        lines = it.dropwhile(lambda ln: not self.is_cnf_comment(ln), sys.stdin)
        comments = it.takewhile(lambda ln: self.is_cnf_comment(ln), lines)
        self.all_literals.update(enumerate(map(lambda s: s.split()[-1], comments), 1))
        # flush stdin
        for ln in sys.stdin:
            pass


    @property
    def parser(self):
        with open("./solution") as solution:
            for line in solution.readlines():
                yield from self.parse_line(line)

    @staticmethod
    def parse_line(line: str):
        return filter(lambda x: not x.startswith('-'), line.removeprefix('SAT').split())


    def __getitem__(self, item):
        lit = self.all_literals[item]
        return tuple(map(int, lit.removeprefix('q(').removesuffix(')').split(',')))


class SudokuBoard9x9:

    def __init__(self):
        self.board: list[list[Optional[int]]] = [[None]*9 for _ in range(9)]

    def fill(self, *, reader=None):
        if not reader:
            raise TypeError(f"can't parse solution, input is {reader}")
        for literal_id in reader:
            x, y, z = parser[int(literal_id)]
            self.board[x - 1][y - 1] = z

    def check(self):
        ok, msg = self._check()
        if not ok:
            print(f"Invalid: {msg}")
        else:
            print("Valid")


    def _check(self) -> tuple[bool, str]:
        def validate_uniqeness(items) -> bool:
            xs = set(filter(bool, items))
            return len(xs) == 9 and sum(xs) == 45
        for i, row in enumerate(self.board, 1):
            if not validate_uniqeness(row):
                return False, f"row{i} = {row}"
        cols = ([self.board[row][col] for row in range(9)] for col in range(9))
        for i, col in enumerate(cols, 1):
            if not validate_uniqeness(col):
                return False, f"col{i} = {col}"
        for box_start in ((0,0), (0,3), (0,6),
                          (3,0), (3,3), (3,6),
                          (6,0), (6,3), (6,6)):
            box_x, box_y = box_start
            box = [self.board[box_x + x][box_y + y] for x in range(3) for y in range(3)]
            if not validate_uniqeness(box):
                return False, f"box({box_x},{box_y}) = {box}"
        return True, ""



    def __repr__(self):
        s = ''
        for x, row in enumerate(self.board):
            for y, col in enumerate(row):
                s += f' {self.board[x][y]} '
                if not (y + 1) % 3:
                    s += '|'
                if not (y + 1) % 9:
                    s += '\n'
            if not (x + 1) % 3:
                s += (f'{"---------+"*3}\n')
        return s




if __name__ == '__main__':
    parser = CNFParser()
    parser.read_cnf()
    board = SudokuBoard9x9()
    board.fill(reader=parser.parser)
    print(f'\n{board}')
    board.check()
#!/usr/bin/env python3

import subprocess
import time


def process_test(test_case, target):
    main_file = "connect.bul"
    args = ["bule", "solve", test_case, main_file]
    start_time = time.time()
    result = subprocess.run(args,text=True,capture_output=True)
    duration = time.time() - start_time
    outcome = 0
    if "s SAT" in result.stdout:
        outcome += 1
    if "s UNSAT" in result.stdout:
        outcome += 2
    ok = (outcome == 1 and target == "SAT") or (outcome == 2 and target == "UNSAT")
    if outcome not in [1, 2]:
        print("strange output {}".format(result.stdout))
    elif ok:
        print("{} ok (target {}) in {:.2f}s".format(test_case, target, duration))
    else:
        print("!!!{} notok (target {})!!!".format(test_case, target))

if __name__ == "__main__":
    filepath = 'tests/all.txt'
    with open(filepath) as fp:
        for line in fp:
            ls = line.split()
            if ls and ls[0][0] == '%':
                continue
            if len(ls) == 2 and ls[1] in ["SAT", "UNSAT"]:
                test_case = "tests/" + ls[0] + ".bul"
                target = ls[1]
                process_test(test_case, target)
            else:
                print("strange line: {}".format(ls))

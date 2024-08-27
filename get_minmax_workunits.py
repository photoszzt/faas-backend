#!/usr/bin/env python3
import json
import os, sys
import re

if len(sys.argv) != 2:
    print("need file path as input")
    sys.exit(1)

ansi_escape = re.compile(r'\x1B(?:[@-Z\\-_]|\[[0-?]*[ -/]*[@-~])')
wus = []
with open(sys.argv[1]) as f:
    for line in f:
        no_colors = ansi_escape.sub('', line)
        #print(no_colors)
        m = re.search("work unit=(\d*)", no_colors)
        if m is not None:
            print("found ", no_colors)
            wus.append(int(m.group(1)))


print(f"work units min, max:   {min(wus)}  ,  {max(wus)} ")
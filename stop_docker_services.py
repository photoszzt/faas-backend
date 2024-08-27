#!/usr/bin/env python3
import re
from subprocess import check_output, call

o = check_output("docker service ls", shell=True).decode()
#print(o)
for line in o.split():
    m = re.search("(func_\S*)", line)
    if m:
        #print(m.group(1))
        call("docker service rm {}".format(m.group(1)), shell=True)

## Diurnal Controller
This controller manipulates the number of replicas maintained by a replication controller throughout the day based on a provided list of times of day and replica counts. It should be run under a replication controller that is in the same namespace as the replication controller that it is manipulating.

Instead of providing replica counts and times of day directly, you may use a script like the one below to generate them using mathematical functions.

```python
from math import *

import os
import sys

def _day_to_2pi(t):
    return float(t) * 2 * pi / (24*3600)

def main(args):
    if len(args) < 3:
        print "Usage: %s sample_interval func" % (args[0],)
        print "func should be a function of the variable t, where t will range from 0"
        print "to 2pi over the course of the day"
        sys.exit(1)
    sampling_interval = int(args[1])
    exec "def f(t): return " + args[2]
    i = 0
    times = []
    counts = []
    while i < 24*60*60:
        hours = i / 3600
        left = i - hours*3600
        min = left / 60
        sec = left - min*60
        times.append("%dh%dm%ds" % (hours, min, sec))
        count = int(round(f(_day_to_2pi(i))))
        counts.append(str(count))
        i += sampling_interval
    print "-times %s -counts %s" % (",".join(times), ",".join(counts))

if __name__ == "__main__":
    main(sys.argv)
```

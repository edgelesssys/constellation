"""Parse the fio logs.

Extracts the bandwidth for I/O,
from various fio benchmarks.

Example log file (extracting read and write bandwidth):
...
Run status group 0 (all jobs):
   READ: bw=5311KiB/s (5438kB/s), 5311KiB/s-5311KiB/s (5438kB/s-5438kB/s), io=311MiB (327MB), run=60058-60058msec
  WRITE: bw=2289KiB/s (2343kB/s), 2289KiB/s-2289KiB/s (2343kB/s-2343kB/s), io=134MiB (141MB), run=60058-60058msec
...
"""


import os
import re
from pathlib import Path
from typing import Dict

# get different mixes of read/write IO as subtests
subtests = {
    'fio_root_async_R70W30':    'fio_async_randR70W30.out',
    'fio_root_async_R100W0':    'fio_async_randR100W0.out',
    'fio_root_async_R0W100':    'fio_async_randR0W100.out',
}


def eval(tests: Dict[str, str]) -> Dict[str, Dict[str, float]]:
    """Read the results of the fio tests.
    Return a result dictionary.
    """
    result = {}
    for t in tests:
        base_path = os.path.join(tests[t], 'dp_fio')
        row = {}
        for subtest in subtests:
            try:
                log_path = next(Path(base_path).rglob(subtests[subtest]))
            except StopIteration:
                raise Exception(
                    f"Error: No iperfclient.out found for network test {subtest} in {base_path}"
                )

            with open(log_path) as f:
                fio = f.readlines()
            if not fio:
                raise Exception(f"Empty fio log {subtest}?")

            for line in fio:
                if "READ" in line:
                    speed = get_io_bw_from_line(line)
                    row[subtest + '_R'] = speed
                elif "WRITE" in line:
                    speed = get_io_bw_from_line(line)
                    row[subtest + '_W'] = speed
        result[t] = row
    return result


# Dictionary to convert units
units = {
    'KiB': 1/1024,
    'MiB': 1,
    'GiB': 1024,
}


def get_io_bw_from_line(line) -> float:
    """Get the IO bandwidth from line and convert to MiB/s.

    Return the IO bandwidth in MiB/s
    """
    #    READ: bw=32.5MiB/s (34.1MB/s), 32.5MiB/s-32.5MiB/s (34.1MB/s-34.1MB/s), io=1954MiB (2048MB), run=60022-60022msec
    match = re.search(r'bw=(\d+\.?\d+)(MiB|KiB|GiB)', line)
    if not match:
        raise Exception("Could not extract bw from fio line.")
    num = float(match.group(1))
    num = num * units[match.group(2)]
    # return in MiB/s
    return num

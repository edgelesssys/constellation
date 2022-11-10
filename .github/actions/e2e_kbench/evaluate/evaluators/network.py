"""Parse the iperf logs.

Extracts the bandwidth for sending and receiving,
from intranode and internode network benchmarks.

Example log file (extract the bitrate for sending and receiving):
...
s1:  - - - - - - - - - - - - - - - - - - - - - - - - -
s1:  [ ID] Interval           Transfer     Bitrate         Retr
s1:  [  5]   0.00-90.00  sec  11.0 GBytes  1.05 Gbits/sec  509             sender
s1:  [  5]   0.00-90.05  sec  11.1 GBytes  1.05 Gbits/sec                  receiver
s1:
s1:  iperf Done.
"""
import os
import re
from pathlib import Path
from typing import Dict

subtests = {
    'net_internode':    'dp_network_internode',
    'net_intranode':    'dp_network_intranode',
}


def eval(tests: Dict[str, str]) -> Dict[str, Dict[str, float]]:
    """Read the results of the network tests.
    Return a result dictionary.
    """
    result = {}
    for t in tests:
        row = {}
        for subtest in subtests:
            base_path = os.path.join(tests[t], subtests[subtest])
            try:
                log_path = next(Path(base_path).rglob('iperfclient.out'))
            except StopIteration:
                raise Exception(
                    f"Error: No iperfclient.out found for network test {subtest} in {base_path}"
                )

            with open(log_path) as f:
                iperfclient = f.readlines()

            if not iperfclient:
                raise Exception("Empty iperfclient?")

            for line in iperfclient:
                if "sender" in line:
                    speed = get_speed_from_line(line)
                    row[subtest + '_snd'] = speed
                    break
                elif "receiver" in line:
                    speed = get_speed_from_line(line)
                    row[subtest + '_rcv'] = speed
                    break
        result[t] = row
    return result


# Dictionary for conversion to Mbit
units = {
    'bits': 1e-6,
    'Mbits': 1,
    'Gbits': 1000,
}


def get_speed_from_line(line) -> float:
    """Extract the network throughput from the line.


    Returns the throughput as Mbit/s.
    """
    match = re.search(
        r'(\d+\.?\d+)\s(bits|Mbits|Gbits)\/sec[\s\d]+(sender|receiver)$', line)
    if not match:
        raise Exception("Could not extract speed from iperf line.")
    num = float(match.group(1))

    # return in Mbit/s with 2 decimal digits
    num = num * units[match.group(2)]
    num = round(num, 2)
    return float(num)

"""Parse the knb logs.

Extracts the bandwidth for sending and receiving,
from k8s-bench-suite network benchmarks.
"""
import json
from typing import Dict


def evaluate(log_path) -> Dict[str, Dict[str, float]]:
    with open(log_path) as f:
        knb = json.load(f)
    if not knb:
        raise Exception(
            f"Error: Empty knb log {log_path}?"
        )

    data = knb['data']
    result = {'pod2pod': {}, 'pod2svc': {}}
    result['pod2pod']['tcp_bw_mbit'] = float(data['pod2pod']['tcp']['bandwidth'])
    result['pod2pod']['udp_bw_mbit'] = float(data['pod2pod']['udp']['bandwidth'])
    result['pod2svc']['tcp_bw_mbit'] = float(data['pod2svc']['tcp']['bandwidth'])
    result['pod2svc']['udp_bw_mbit'] = float(data['pod2svc']['udp']['bandwidth'])
    result['mtu'] = int(data['mtu'])

    return result

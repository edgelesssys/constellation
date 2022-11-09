"""Evaluator for the K-Bench default test."""
import os
import re
from collections import defaultdict
from typing import Dict

pod_latencies = {
    'pod_create':   'create pod latency:',
    'pod_list':     'list pod latency:',
    'pod_get':      'get pod latency:',
    'pod_update':   'update pod latency:',
    'pod_delete':   'delete pod latency:',
}

deployment_latencies = {
    'depl_create':  'create deployment latency:',
    'depl_list':    'list deployment latency:',
    'depl_update':  'update deployment latency:',
    'depl_scale':   'scale deployment latency:',
    'depl_delete':  'delete deployment latency:',
}

service_latencies = {
    'svc_create':   'create service latency:',
    'svc_list':     'list service latency:',
    'svc_get':      'get service latency:',
    'svc_update':   'update service latency:',
    'svc_delete':   'delete service latency:',
}


def eval(tests: Dict[str, str]) -> Dict[str, Dict[str, float]]:
    """Read the results of the default tests.

    Return a result dictionary.
    """
    result = {}
    for t in tests:
        row = defaultdict(float)
        # read the default result file
        kbench = []
        with open(os.path.join(tests[t], 'default', 'kbench.log'), 'r') as f:
            kbench = f.readlines()

        if not kbench:
            raise Exception("Empty kbench.log")

        subtests = [pod_latencies, service_latencies, deployment_latencies]
        for latency_dict in subtests:
            # Get the API Call Latencies (median)
            for key in latency_dict:
                line = get_line_containing_needle(
                    lines=kbench, needle=latency_dict[key])
                median = get_median_from_line(line=line)
                # round API latency to full ms granularity
                row[key] = round(float(median))

        result[t] = row
    return result


def get_median_from_line(line):
    """Extract the value (median) from the line."""
    return re.search(r'\s(\d+\.\d+)(.+)', line).group(1)


def get_line_containing_needle(lines, needle):
    """Find matching line from list of lines."""
    matches = list(filter(lambda l: needle in l, lines))
    if len(matches) > 1:
        raise Exception(f"'{needle}' matched multiple times..")
    return matches[0]

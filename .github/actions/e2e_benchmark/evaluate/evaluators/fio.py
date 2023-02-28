"""Parse the fio logs.

Extracts the bandwidth for I/O,
from a fio benchmark json output.
"""
from typing import Dict
import json



def evaluate(log_path) -> Dict[str, Dict[str, float]]:
    with open(log_path) as f:
        fio = json.load(f)
    if not fio:
        raise Exception(
            f"Error: Empty fio log {log_path}?")
    if len(fio) != 1:
        raise Exception(
            "Error: Unexpected fio log format"
        )
    tests = fio[0]['Raw']['result']['jobs']
    result = {}
    for test in tests:
        if test['jobname'] == 'read_iops':
            result[test['jobname']] = {'iops': float(test['read']['iops'])}
        elif test['jobname'] == 'write_iops':
            result[test['jobname']] = {'iops': float(test['write']['iops'])}
        elif test['jobname'] == 'read_bw':
            result[test['jobname']] = {'bw_kbytes': float(test['read']['bw'])}
        elif test['jobname'] == 'write_bw':
            result[test['jobname']] = {'bw_kbytes': float(test['write']['bw'])}
        else:
            raise Exception(
                f"Error: Unexpected fio test: {test['jobname']}"
            )
    return result

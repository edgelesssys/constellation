"""Compare the current benchmark data against the previous."""
import os
import json
from typing import Tuple

# Progress indicator icons
PROGRESS = ['⬇️', '⬆️']

# List of benchmarks for which higher numbers are better
BIGGER_BETTER = [
    'iops',
    'bw_kbytes',
    'tcp_bw_mbit',
    'upd_bw_mbit',
]

# Lookup for test suite -> unit
UNIT_STR = {
    'iops': 'IOPS',
    'bw_kbytes': 'KiB/s',
    'tcp_bw_mbit': 'Mbit/s',
    'upd_bw_mbit': 'Mbit/s',
}
# API units are ms, so this is shorter than cluttering the dictionary:
API_UNIT_STR = "ms"

# List of allowed deviation
ALLOWED_RATIO_DELTA = {
    'iops': 0.7,
    'bw_kbytes': 0.7,
    'tcp_bw_mbit': 0.7,
    'upd_bw_mbit': 0.7,
}

def is_bigger_better(bench_suite: str) -> bool:
    return bench_suite in BIGGER_BETTER


def get_paths() -> Tuple[str, str]:
    """Read the benchmark data paths.

    Expects ENV vars (required):
    - PREV_BENCH=/path/to/previous.json
    - CURR_BENCH=/path/to/current.json

    Raises TypeError if at least one of them is missing.

    Returns: a tuple of (prev_bench_path, curr_bench_path).
    """
    path_prev = os.environ.get('PREV_BENCH', None)
    path_curr = os.environ.get('CURR_BENCH', None)
    if not path_prev or not path_curr:
        raise TypeError(
            'Both ENV variables PREV_BENCH and CURR_BENCH are required.')
    return path_prev, path_curr


def main() -> None:
    """Compare the current benchmark data against the previous.

    Create a markdown table showing the benchmark progressions.

    Print the result to stdout.
    """
    path_prev, path_curr = get_paths()
    try:
        with open(path_prev) as f_prev:
            bench_prev = json.load(f_prev)
        with open(path_curr) as f_curr:
            bench_curr = json.load(f_curr)
    except OSError as e:
        raise ValueError('Failed reading benchmark file: {e}'.format(e=e))

    try:
        name = bench_curr['provider']
    except KeyError:
        raise ValueError(
            'Current benchmark record file does not contain provider.')
    try:
        prev_name = bench_prev['provider']
    except KeyError:
        raise ValueError(
            'Previous benchmark record file does not contain provider.')
    if name != prev_name:
        raise ValueError(
            'Cloud providers of previous and current benchmark data do not match.')

    if 'fio' not in bench_prev.keys() or 'fio' not in bench_curr.keys():
        raise ValueError('Benchmarks do not both contain fio records.')

    if 'knb' not in bench_prev.keys() or 'knb' not in bench_curr.keys():
        raise ValueError('Benchmarks do not both contain knb records.')

    md_lines = [
        '# {name}'.format(name=name),
        '',
        '<details>',
        '',
        '- Commit of current benchmark: [{ch}](https://github.com/edgelesssys/constellation/commit/{ch})'.format(ch=bench_curr['metadata']['github.sha']),
        '- Commit of previous benchmark: [{ch}](https://github.com/edgelesssys/constellation/commit/{ch})'.format(ch=bench_prev['metadata']['github.sha']),
        '',
        '| Benchmark suite | Metric | Current | Previous | Ratio |',
        '|-|-|-|-|-|',
    ]

    # compare FIO results
    for subtest, metrics in bench_prev['fio'].items():
        for metric in metrics.keys():
            md_lines.append(compare_test('fio', subtest, metric, bench_prev, bench_curr))

    # compare knb results
    for subtest, metrics in bench_prev['knb'].items():
        for metric  in metrics.keys():
            md_lines.append(compare_test('knb', subtest, metric, bench_prev, bench_curr))

    md_lines += ['', '</details>']
    print('\n'.join(md_lines))


def compare_test(test, subtest, metric, bench_prev, bench_curr) -> str:
        if subtest not in bench_curr[test]:
            raise ValueError(
                'Benchmark record from previous benchmark not in current.')
        val_prev = bench_prev[test][subtest][metric]
        val_curr = bench_curr[test][subtest][metric]

        # get unit string or use default API unit string
        unit = UNIT_STR.get(metric, API_UNIT_STR)

        if val_curr == 0 or val_prev == 0:
            ratio = 'N/A'
        else:
            if is_bigger_better(bench_suite=metric):
                ratio_num = val_curr / val_prev
                if ratio_num < ALLOWED_RATIO_DELTA.get(metric, 1):
                    set_failed()
            else:
                ratio_num = val_prev / val_curr
                if ratio_num > ALLOWED_RATIO_DELTA.get(metric, 1):
                    set_failed()
                
            ratio_num = round(ratio_num, 3)
            emoji = PROGRESS[int(ratio_num >= 1)]
            ratio = f'{ratio_num} {emoji}'

        return f'| {subtest} | {metric} ({unit}) | {val_curr} | {val_prev} | {ratio} |'

def set_failed() -> None:
    os.environ['COMPARISON_SUCCESS'] = str(False)

if __name__ == '__main__':
    main()

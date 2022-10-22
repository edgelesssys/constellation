"""Compare the current benchmark data against the previous."""
import os
import json
from typing import Tuple

# Progress indicator icons
PROGRESS = ['⬇️', '⬆️']

# List of benchmarks for which higher numbers are better
BIGGER_BETTER = [
    'net_internode_snd',
    'net_intranode_snd',
    'fio_root_async_R70W30_R',
    'fio_root_async_R70W30_W',
    'fio_root_async_R100W0_R',
    'fio_root_async_R0W100_W',
]

# Lookup for test suite -> unit
UNIT_STR = {
    'net_internode_snd': 'Mbit/s',
    'net_intranode_snd': 'Mbit/s',
    'fio_root_async_R70W30_R': 'MiB/s',
    'fio_root_async_R70W30_W': 'MiB/s',
    'fio_root_async_R100W0_R': 'MiB/s',
    'fio_root_async_R0W100_W': 'MiB/s',
}
# API units are ms, so this is shorter than cluttering the dictionary:
API_UNIT_STR = "ms"


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

    with open(path_prev) as f_prev:
        bench_prev = json.load(f_prev)
    with open(path_curr) as f_curr:
        bench_curr = json.load(f_curr)

    name = next(iter(bench_prev.keys()))
    if name != next(iter(bench_curr.keys())):
        raise ValueError(
            "Cloud providers of previous and current benchmark data don't match.")

    md_lines = [
        '# {name}'.format(name=name),
        '',
        '<details>',
        '',
        '| Benchmark suite | Current | Previous | Ratio |',
        '|-|-|-|-|',
    ]

    for subtest, _ in bench_prev[name].items():
        val_prev = bench_prev[name][subtest]
        val_curr = bench_curr[name][subtest]

        # get unit string or use default API unit string
        unit = UNIT_STR.get(subtest, API_UNIT_STR)

        if val_curr == 0 or val_prev == 0:
            ratio = 'N/A'
        else:
            if is_bigger_better(bench_suite=subtest):
                ratio_num = val_prev / val_curr
            else:
                ratio_num = val_curr / val_prev
            ratio_num = round(ratio_num, 3)
            emoji = PROGRESS[int(ratio_num >= 1)]
            ratio = '{ratio} {emoji}'.format(ratio=ratio_num, emoji=emoji)

        line = '| {test} ({unit}) | {val_curr} | {val_prev} | {ratio} |'.format(
            test=subtest, unit=unit, val_curr=val_curr, val_prev=val_prev, ratio=ratio,
        )
        md_lines.append(line)

    md_lines += ['', '</details>']
    print('\n'.join(md_lines))


if __name__ == '__main__':
    main()

"""Parse logs of K-Bench tests and generate performance graphs."""
import json
import os
from collections import defaultdict
from typing import Tuple

from evaluators import default, fio, network


def configure() -> Tuple[str, str, str, str | None, str]:
    """Read the benchmark data paths.

    Expects ENV vars (required):
    - KBENCH_RESULTS=/path/to/k-bench/out
    - CSP=azure
    - BDIR=benchmarks

    Optional:
    - EXT_NAME=AKS  # Overrides "constellation-$CSP" naming to parse results from managed Kubernetes
    - GITHUB_SHA=ffac5... # Set by GitHub actions, stored in the result JSON.

    Raises TypeError if at least one of them is missing.

    Returns: a tuple of (base_path, csp, out_dir, ext_provider_name).
    """
    base_path = os.environ.get('KBENCH_RESULTS', None)
    csp = os.environ.get('CSP', None)
    out_dir = os.environ.get('BDIR', None)
    if not base_path or not csp or not out_dir:
        raise TypeError(
            'ENV variables KBENCH_RESULTS, CSP, BDIR are required.')

    ext_provider_name = os.environ.get('EXT_NAME', None)
    commit_hash = os.environ.get('GITHUB_SHA', 'N/A')
    return base_path, csp, out_dir, ext_provider_name, commit_hash


def main() -> None:
    """Read and parse the K-Bench tests.

    Write results of the current environment to a JSON file.
    """
    base_path, csp, out_dir, ext_provider_name, commit_hash = configure()

    if ext_provider_name is None:
        # Constellation benchmark.
        ext_provider_name = 'constellation-{csp}'.format(csp=csp)

    # Expect the results in directory:
    # kbench-EXT_PROVIDER_NAME/
    benchmark_path = os.path.join(
        base_path,
        'kbench-{csp}'.format(csp=ext_provider_name),
    )
    tests = {ext_provider_name: benchmark_path}
    out_file_name = '{nm}.json'.format(nm=ext_provider_name)

    if not os.path.exists(benchmark_path):
        raise ValueError(
            'Benchmarks do not exist at {path}.'.format(path=benchmark_path))

    # Parse subtest
    default_results = default.evaluate(tests=tests)
    network_results = network.evaluate(tests=tests)
    fio_results = fio.evaluate(tests=tests)

    combined_results = {'commit': commit_hash, 'subject': ext_provider_name}
    combined_results['kbench'] = {}

    for test in tests:
        combined_results['kbench'].update(default_results[test])
        combined_results['kbench'].update(network_results[test])
        combined_results['kbench'].update(fio_results[test])

    # Write the compact results.
    save_path = os.path.join(out_dir, out_file_name)
    with open(save_path, 'w') as w:
        json.dump(combined_results, fp=w, sort_keys=False, indent=2)


if __name__ == '__main__':
    main()

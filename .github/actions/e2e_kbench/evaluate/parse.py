"""Parse logs of K-Bench tests and generate performance graphs."""
import json
import os
from collections import defaultdict
from typing import Tuple

from evaluators import default, fio, network


def configure() -> Tuple[str, str, str, str | None]:
    """Read the benchmark data paths.

    Expects ENV vars (required):
    - KBENCH_RESULTS=/path/to/k-bench/out
    - CSP=azure
    - BDIR=benchmarks

    Optional:
    - EXT_NAME=AKS  # Overrides "constellation-$CSP" naming, used to parse results from managed Kubernetes

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
    return base_path, csp, out_dir, ext_provider_name


def main() -> None:
    """Read and parse the K-Bench tests.

    Write results of the current environment to a JSON file.
    """
    base_path, csp, out_dir, ext_provider_name = configure()

    if ext_provider_name:
        # Managed Kubernetes benchmark, expect the results in directory:
        # kbench-EXT_NAME/
        benchmark_path = os.path.join(
            base_path,
            "kbench-" + ext_provider_name,
        )
        tests = {ext_provider_name: benchmark_path}
        out_file_name = "{nm}.json".format(nm=ext_provider_name)
    else:
        # Constellation benchmark.
        benchmark_path = os.path.join(
            base_path,
            "kbench-constellation-" + csp,
        )
        tests = {f"constellation-{csp}": benchmark_path}
        out_file_name = "constellation-{csp}.json".format(csp=csp)

    if not os.path.exists(benchmark_path):
        raise Exception(
            f'Path to benchmarks {benchmark_path} does not exist.')

    # Parse subtest
    default_results = default.eval(tests=tests)
    network_results = network.eval(tests=tests)
    fio_results = fio.eval(tests=tests)

    combined_results = defaultdict(dict)

    for test in tests:
        combined_results[test].update(default_results[test])
        combined_results[test].update(network_results[test])
        combined_results[test].update(fio_results[test])

    # Write the compact results.
    save_path = os.path.join(out_dir, out_file_name)
    with open(save_path, 'w') as w:
        json.dump(combined_results, fp=w, sort_keys=False, indent=2)


if __name__ == "__main__":
    main()

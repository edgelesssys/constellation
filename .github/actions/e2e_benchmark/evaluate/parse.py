"""Parse logs of K-Bench tests and generate performance graphs."""
import json
import os
from typing import Tuple
from datetime import datetime

from evaluators import fio, knb


def configure() -> Tuple[str, str, str, str | None, str, str, str, str]:
    """Read the benchmark data paths.

    Expects ENV vars (required):
    - BENCH_RESULTS=/path/to/bench/out
    - CSP=azure
    - BDIR=benchmarks

    Optional:
    - EXT_NAME=AKS  # Overrides "constellation-$CSP" naming to parse results from managed Kubernetes
    - GITHUB_SHA=ffac5... # Set by GitHub actions, stored in the result JSON.

    Raises TypeError if at least one of them is missing.

    Returns: a tuple of (base_path, csp, out_dir, ext_provider_name).
    """
    base_path = os.environ.get('BENCH_RESULTS', None)
    csp = os.environ.get('CSP', None)
    out_dir = os.environ.get('BDIR', None)
    if not base_path or not csp or not out_dir:
        raise TypeError(
            'ENV variables BENCH_RESULTS, CSP, BDIR are required.')

    ext_provider_name = os.environ.get('EXT_NAME', None)
    commit_hash = os.environ.get('GITHUB_SHA', 'N/A')
    commit_ref = os.environ.get('GITHUB_REF_NAME', 'N/A')
    actor = os.environ.get('GITHUB_ACTOR', 'N/A')
    workflow = os.environ.get('GITHUB_WORKFLOW', 'N/A')
    return base_path, csp, out_dir, ext_provider_name, commit_hash, commit_ref, actor, workflow


def main() -> None:
    """Read and parse the K-Bench tests.

    Write results of the current environment to a JSON file.
    """
    base_path, csp, out_dir, ext_provider_name, commit_hash, commit_ref, actor, workflow = configure()

    if ext_provider_name is None:
        # Constellation benchmark.
        ext_provider_name = f'constellation-{csp}'

    # Expect the results in directory:
    fio_path = os.path.join(
        base_path,
        f'fio-{ext_provider_name}.json',
    )
    knb_path = os.path.join(
        base_path,
        f'knb-{ext_provider_name}.json',
    )
    out_file_name = f'{ext_provider_name}.json'

    if not os.path.exists(fio_path) or not os.path.exists(knb_path):
        raise ValueError(
            f'Benchmarks do not exist at {fio_path} or {knb_path}.')

    # Parse subtest
    knb_results = knb.evaluate(knb_path)
    fio_results = fio.evaluate(fio_path)

    combined_results = {'metadata': {
                            'github.sha': commit_hash,
                            'github.ref-name': commit_ref,
                            'github.actor': actor,
                            'github.workflow': workflow,
                            'created': str(datetime.now()),
                        },
                        'provider': ext_provider_name,
                        'fio': {}, 
                        'knb': {}}

    combined_results['knb'].update(knb_results)
    combined_results['fio'].update(fio_results)

    # Write the compact results.
    save_path = os.path.join(out_dir, out_file_name)
    with open(save_path, 'w+') as w:
        json.dump(combined_results, fp=w, sort_keys=False, indent=2)


if __name__ == '__main__':
    main()

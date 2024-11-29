"""Parse logs of K-Bench tests and generate performance graphs."""
import json
import os
from typing import Tuple
from datetime import datetime

from evaluators import fio, knb


def configure() -> Tuple[str, str, str, str, str | None, str, str, str, str]:
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
    attestation_variant = os.environ.get('ATTESTATION_VARIANT', None)
    out_dir = os.environ.get('BDIR', None)
    if not base_path or not csp or not out_dir or not attestation_variant:
        raise TypeError(
            'ENV variables BENCH_RESULTS, CSP, BDIR, ATTESTATION_VARIANT are required.')

    ext_provider_name = os.environ.get('EXT_NAME', None)
    commit_hash = os.environ.get('GITHUB_SHA', 'N/A')
    commit_ref = os.environ.get('GITHUB_REF_NAME', 'N/A')
    actor = os.environ.get('GITHUB_ACTOR', 'N/A')
    workflow = os.environ.get('GITHUB_WORKFLOW', 'N/A')
    return base_path, csp, attestation_variant, out_dir, ext_provider_name, commit_hash, commit_ref, actor, workflow


class BenchmarkParser:
    def __init__(self, base_path, csp, attestation_variant, out_dir, ext_provider_name=None, commit_hash="N/A", commit_ref="N/A", actor="N/A", workflow="N/A"):
        self.base_path = base_path
        self.csp = csp
        self.attestation_variant = attestation_variant
        self.out_dir = out_dir
        self.ext_provider_name = ext_provider_name
        if not self.ext_provider_name:
            self.ext_provider_name = f'constellation-{attestation_variant}'
        self.commit_hash = commit_hash
        self.commit_ref = commit_ref
        self.actor = actor
        self.workflow = workflow

    def parse(self) -> None:
        """Read and parse the K-Bench tests.

        Write results of the current environment to a JSON file.
        """

        # Expect the results in directory:
        fio_path = os.path.join(
            self.base_path,
            f'fio-{self.ext_provider_name}.json',
        )
        knb_path = os.path.join(
            self.base_path,
            f'knb-{self.ext_provider_name}.json',
        )
        out_file_name = f'{self.ext_provider_name}.json'

        if not os.path.exists(fio_path) or not os.path.exists(knb_path):
            raise ValueError(
                f'Benchmarks do not exist at {fio_path} or {knb_path}.')

        # Parse subtest
        knb_results = knb.evaluate(knb_path)
        fio_results = fio.evaluate(fio_path)

        # Get timestamp
        now = datetime.now()
        timestamp = now.strftime("%Y-%m-%dT%H:%M:%S.%fZ")

        combined_results = {'metadata': {
            'github.sha': self.commit_hash,
            'github.ref-name': self.commit_ref,
            'github.actor': self.actor,
            'github.workflow': self.workflow,
        },
            '@timestamp': str(timestamp),
            'provider': self.ext_provider_name,
            'attestationVariant': self.attestation_variant,
            'fio': {},
            'knb': {}}

        combined_results['knb'].update(knb_results)
        combined_results['fio'].update(fio_results)

        # Write the compact results.
        save_path = os.path.join(self.out_dir, out_file_name)
        with open(save_path, 'w+') as w:
            json.dump(combined_results, fp=w, sort_keys=False, indent=2)


def main():
    base_path, csp, attestation_variant, out_dir, ext_provider_name, commit_hash, commit_ref, actor, workflow = configure()
    p = BenchmarkParser(base_path, csp, attestation_variant, out_dir, ext_provider_name,
                        commit_hash, commit_ref, actor, workflow)
    p.parse()


if __name__ == '__main__':
    main()

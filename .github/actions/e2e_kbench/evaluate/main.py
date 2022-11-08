"""Parse logs of K-Bench tests and generate performance graphs."""
import json
import os
from collections import defaultdict

import numpy as np
from evaluators import default, fio, network
from matplotlib import pyplot as plt

BAR_COLOR = '#90FF99'  # Mint Green

# Rotate bar labels by X degrees
LABEL_ROTATE_BY = 30
LABEL_FONTSIZE = 9

# Some lookup dictionaries for x axis
api_suffix = 'ms'
pod_key2header = {
    'pod_create':   'Pod Create',
    'pod_list':     'Pod List',
    'pod_get':      'Pod Get',
    'pod_update':   'Pod Update',
    'pod_delete':   'Pod Delete',
}
svc_key2header = {
    'svc_create':   'Service Create',
    'svc_list':     'Service List',
    'svc_update':   'Service Update',
    'svc_delete':   'Service Delete',
    'svc_get':      'Service Get',
}
depl_key2header = {
    'depl_create':  'Deployment Create',
    'depl_list':    'Deployment List',
    'depl_update':  'Deployment Update',
    'depl_scale':   'Deployment Scale',
    'depl_delete':  'Deployment Delete',
}

fio_suffix = 'MiB/s'
fio_key2header = {
    'fio_root_async_R70W30_R':   'async_R70W30 mix,\n seq. reads',
    'fio_root_async_R70W30_W':   'async_R70W30 mix,\n seq. writes',
    'fio_root_async_R100W0_R':   'async_R100W0 mix,\n seq. reads',
    'fio_root_async_R0W100_W':   'async_R0W100 mix,\n seq. writes',
}

net_suffix = 'Mbit/s'
net_key2header = {
    'net_internode_snd':            f'iperf internode \n send ({net_suffix})',
    'net_intranode_snd':            f'iperf intranode \n send ({net_suffix})',
}


def configure() -> dict:
    """Set the config.

    Raises BaseException if base_path or CSP missing.

    Returns a config dict with the BASE_PATH to the tests
    and the cloud service provider CSP.
    """
    base_path = os.getenv('KBENCH_RESULTS', None)
    if not base_path or not os.path.isdir(base_path):
        raise Exception("Environment variable 'KBENCH_RESULTS' \
needs to point to the K-Bench results root folder")

    csp = os.getenv('CSP', None)
    if not csp:
        raise Exception("Environment variable 'CSP' \
needs to name the cloud service provider.")
    return {'BASE_PATH': base_path, 'CSP': csp}


def bar_chart(data, headers, title='', suffix='', val_label=True, y_log=False):
    """Generate a bar chart from data.

    Args:
        data (list): List of value points.
        headers (list): List of headers (x-axis).
        title (str, optional): The title for the chart. Defaults to "".
        suffix (str, optional): The suffix for values e.g. "MiB/s". Defaults to "".
        val_label (bool, optional): Put a label of the value over the bar chart. Defaults to True.
        y_log (bool, optional): Set the y-axis to a logarithmic scale. Defaults to False.
    Returns:
        fig (matplotlib.pyplot.figure): The pyplot figure
    """
    fig, ax = plt.subplots(figsize=(8, 5))
    fig.patch.set_facecolor('white')
    ax.set_xticks(np.arange(len(headers)))
    ax.set_xticklabels(headers)
    if y_log:
        ax.set_yscale('log')
    bars = ax.bar(headers, data, color=BAR_COLOR, edgecolor='black')
    if val_label:
        ax.bar_label(bars, fmt='%g {suffix}'.format(suffix=suffix))
    plt.setp(ax.get_xticklabels(), fontsize=LABEL_FONTSIZE, rotation=LABEL_ROTATE_BY)
    plt.title(f'{title} ({suffix})')
    plt.tight_layout()
    return fig


def main() -> None:
    """Read, parse and evaluate the K-Bench tests.

    Generate a human-readable table and diagrams.
    """
    config = configure()

    benchmark_path = os.path.join(
        config['BASE_PATH'],
        "kbench-constellation-" + config['CSP'],
    )
    if not os.path.exists(benchmark_path):
        raise Exception(f'Path to benchmarks {benchmark_path} does not exist.')

    tests = {f"constellation-{config['CSP']}": benchmark_path}

    # Execute tests
    default_results = default.eval(tests=tests)
    network_results = network.eval(tests=tests)
    fio_results = fio.eval(tests=tests)

    combined_results = defaultdict(dict)
    for test in tests:
        combined_results[test].update(default_results[test])
        combined_results[test].update(network_results[test])
        combined_results[test].update(fio_results[test])

    # Write the compact results.
    with open('kbench_results.json', 'w') as w:
        json.dump(combined_results, fp=w, sort_keys=False, indent=2)

    # Generate graphs.
    subject = list(combined_results.keys())[0]
    data = combined_results[subject]

    # Combine the evaluation of the Kubernetes API benchmarks
    for i, api in enumerate([pod_key2header, svc_key2header, depl_key2header]):
        api_data = [data[h] for h in api]
        hdrs = api.values()
        bar_chart(data=api_data, headers=hdrs, title="API Latency", suffix=api_suffix)
        plt.savefig(f'api_{i}_perf.png', bbox_inches="tight")

    # Network chart
    net_data = [data[h] for h in net_key2header]
    hdrs = net_key2header.values()
    bar_chart(data=net_data, headers=hdrs, title="Network Throughput", suffix=net_suffix)
    plt.savefig('net_perf.png', bbox_inches="tight")

    # fio chart
    fio_data = [data[h] for h in fio_key2header]
    hdrs = fio_key2header.values()
    bar_chart(data=fio_data, headers=hdrs, title="Storage Throughput", suffix=fio_suffix)
    plt.savefig('storage_perf.png', bbox_inches="tight")


if __name__ == "__main__":
    main()

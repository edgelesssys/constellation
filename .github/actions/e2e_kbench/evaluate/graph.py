"""Generate graphs comparing K-Bench benchmarks across cloud providers and Constellation."""
import json
import os
from collections import defaultdict

import numpy as np
from matplotlib import pyplot as plt

SUBJECTS = [
    'constellation-azure',
    'AKS',
    'constellation-gcp',
    'GKE',
]

LEGEND_NAMES = [
    'Constellation on Azure',
    'AKS',
    'Constellation on GCP',
    'GKE',
]

BAR_COLORS = ['#90FF99', '#929292', '#8B04DD', '#000000']

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
    'net_internode_snd':            'iperf internode \n send ({net_suffix})'.format(net_suffix=net_suffix),
    'net_intranode_snd':            'iperf intranode \n send ({net_suffix})'.format(net_suffix=net_suffix),
}


def configure() -> str:
    """Read the benchmark data paths.

    Expects ENV vars (required):
    - BDIR=benchmarks

    Raises TypeError if at least one of them is missing.

    Returns: out_dir
    """
    out_dir = os.environ.get('BDIR', None)
    if not out_dir:
        raise TypeError(
            'ENV variables BDIR is required.')
    return out_dir


def bar_chart(data, headers, title='', suffix='', val_label=True, y_log=False):
    """Draws a bar chart with multiple bars per data point.

    Args:
        data (dict[str, list]): Benchmark data dictionary: subject -> lists of value points
        headers (list): List of headers (x-axis).
        title (str, optional): The title for the chart. Defaults to "".
        suffix (str, optional): The suffix for values e.g. "MiB/s". Defaults to "".
        val_label (bool, optional): Put a label of the value over the bar chart. Defaults to True.
        y_log (bool, optional): Set the y-axis to a logarithmic scale. Defaults to False.
    Returns:
        fig (matplotlib.pyplot.figure): The pyplot figure
    """
    fig, ax = plt.subplots(figsize=(10, 5))
    fig.patch.set_facecolor('white')

    # Number of bars per group
    n_bars = len(data)

    # The width of a single bar
    bar_width = 0.8 / n_bars

    # List containing handles for the drawn bars, used for the legend
    bars = []

    # Iterate over all data
    for i, values in enumerate(data.values()):
        # The offset in x direction of that bar
        x_offset = (i - n_bars / 2) * bar_width + bar_width / 2

        # Draw a bar for every value of that type
        for x, y in enumerate(values):
            bar = ax.bar(x + x_offset, y, width=bar_width * 0.9,
                         color=BAR_COLORS[i % len(BAR_COLORS)], edgecolor='black')
            if val_label:
                ax.bar_label(bar, padding=1,
                             fmt='%g {suffix}'.format(suffix=suffix))
        # Add a handle to the last drawn bar, which we'll need for the legend
        bars.append(bar[0])
    # Draw legend
    ax.legend(bars, LEGEND_NAMES)
    if y_log:
        ax.set_yscale('log')
    ax.set_xticks(np.arange(len(headers)))
    ax.set_xticklabels(headers)

    plt.setp(ax.get_xticklabels(), fontsize=LABEL_FONTSIZE,
             rotation=LABEL_ROTATE_BY)
    plt.title('{title} ({suffix})'.format(title=title, suffix=suffix))
    plt.tight_layout()
    return fig


def main():
    """Read the files and create diagrams."""
    out_dir = configure()
    combined_results = defaultdict(dict)

    for test in SUBJECTS:
        # Read the previous results
        read_path = os.path.join(
            out_dir, '{subject}.json'.format(subject=test))
        try:
            with open(read_path, 'r') as res_file:
                combined_results[test].update(json.load(res_file))
        except OSError as e:
            raise ValueError(
                'Failed reading {subject} benchmark records: {e}'.format(subject=test, e=e))

    # Combine the evaluation of the Kubernetes API benchmarks
    for i, api in enumerate([pod_key2header, svc_key2header, depl_key2header]):
        api_data = {}
        for s in SUBJECTS:
            points = combined_results[s]['kbench']
            subject_data = [points[h] for h in api]
            api_data[s] = subject_data
        hdrs = list(api.values())
        bar_chart(data=api_data, headers=hdrs,
                  title='API Latency', suffix=api_suffix, y_log=True)

        save_name = os.path.join(out_dir, 'api_{i}_perf.png'.format(i=i))
        plt.savefig(save_name, bbox_inches='tight')

    # Network chart
    net_data = {}
    for s in SUBJECTS:
        points = combined_results[s]['kbench']
        subject_data = [points[h] for h in net_key2header]
        net_data[s] = subject_data
    hdrs = list(net_key2header.values())
    bar_chart(data=net_data, headers=hdrs,
              title='Network Throughput', suffix=net_suffix, y_log=True)
    save_name = os.path.join(out_dir, 'net_perf.png')
    plt.savefig(save_name, bbox_inches='tight')

    # fio chart
    fio_data = {}
    for s in SUBJECTS:
        points = combined_results[s]['kbench']
        subject_data = [points[h] for h in fio_key2header]
        fio_data[s] = subject_data
    hdrs = list(fio_key2header.values())
    bar_chart(data=fio_data, headers=hdrs,
              title='Storage Throughput', suffix=fio_suffix, y_log=True)
    save_name = os.path.join(out_dir, 'storage_perf.png')
    plt.savefig(save_name, bbox_inches='tight')


if __name__ == '__main__':
    main()

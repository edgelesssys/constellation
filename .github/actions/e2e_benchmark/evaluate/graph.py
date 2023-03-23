"""Generate graphs comparing K-Bench benchmarks across cloud providers and Constellation."""
import json
import os
import tempfile
from collections import defaultdict
from pathlib import Path
from urllib import request

import numpy as np
from matplotlib import pyplot as plt
from matplotlib import font_manager as fm


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

FONT_URL = "https://github.com/google/fonts/raw/main/apache/roboto/static/Roboto-Regular.ttf"
FONT_NAME = "Roboto-Regular.ttf"
FONT_SIZE = 13

# Some lookup dictionaries for x axis
fio_iops_unit = 'IOPS'
fio_bw_unit = 'KiB/s'

net_unit = 'Mbit/s'


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


def bar_chart(data, title='', unit='', x_label=''):
    #     """Draws a bar chart with multiple bars per data point.

    #     Args:
    #         data (dict[str, list]): Benchmark data dictionary: subject -> lists of value points
    #         title (str, optional): The title for the chart. Defaults to "".
    #         suffix (str, optional): The suffix for values e.g. "MiB/s". Defaults to "".
    #         x_label (str, optional): The label for the x-axis. Defaults to "".
    #     Returns:
    #         fig (matplotlib.pyplot.figure): The pyplot figure
    #     """

    # Create plot and set configs
    plt.rcdefaults()
    plt.rc('font', family=FONT_NAME, size=FONT_SIZE)
    fig, ax = plt.subplots(figsize=(10, 5))

    # Calculate y positions
    y_pos = np.arange(len(data))

    bars = ax.barh(y_pos, data.values(), align='center', color=BAR_COLORS)

    # Axis formatting
    ax.spines['top'].set_visible(False)
    ax.spines['right'].set_visible(False)
    ax.spines['left'].set_visible(False)
    ax.spines['bottom'].set_color('#DDDDDD')
    ax.tick_params(bottom=False, left=False)
    ax.set_axisbelow(True)
    ax.xaxis.grid(True, color='#EEEEEE')
    ax.yaxis.grid(False)

    # Bar annotations
    for bar in bars:
        ax.text(
            1.03*bar.get_width(),
            bar.get_y() + bar.get_height() / 2,
            f'{bar.get_width():.0f}',
            verticalalignment='center',
        )

    # Set labels and titles
    ax.set_yticks(y_pos, labels=data.keys())
    ax.invert_yaxis()  # labels read top-to-bottom
    ax.set_xlabel(x_label, fontdict={"fontsize": 12})
    if unit != '':
        unit = f"({unit})"
    ax.set_title(f'{title} {unit}', fontdict={
                 "fontsize": 20, 'weight': 'bold'})

    plt.tight_layout()
    # plt.show()
    return fig


def main():
    """ Download and setup fonts"""
    path = Path(tempfile.mkdtemp())
    font_path = path / FONT_NAME
    request.urlretrieve(FONT_URL, font_path)

    font = fm.FontEntry(fname=str(font_path), name=FONT_NAME)
    fm.fontManager.ttflist.append(font)

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

    # Network charts
    # P2P TCP
    net_data = {}
    for s, l in zip(SUBJECTS, LEGEND_NAMES):
        net_data[l] = int(combined_results[s]['knb']['pod2pod']['tcp_bw_mbit'])
    bar_chart(data=net_data,
              title='K8S CNI Benchmark - Pod to Pod - TCP - Bandwidth',
              unit=net_unit,
              x_label=f" TCP Bandwidth in {net_unit} - Higher is better")
    save_name = os.path.join(out_dir, 'benchmark_net_p2p_tcp.png')
    plt.savefig(save_name)

    # P2P TCP
    net_data = {}
    for s, l in zip(SUBJECTS, LEGEND_NAMES):
        net_data[l] = int(combined_results[s]['knb']['pod2pod']['udp_bw_mbit'])
    bar_chart(data=net_data,
              title='K8S CNI Benchmark - Pod to Pod - UDP - Bandwidth',
              unit=net_unit,
              x_label=f" UDP Bandwidth in {net_unit} - Higher is better")
    save_name = os.path.join(out_dir, 'benchmark_net_p2p_udp.png')
    plt.savefig(save_name)

    # P2SVC TCP
    net_data = {}
    for s, l in zip(SUBJECTS, LEGEND_NAMES):
        net_data[l] = int(combined_results[s]['knb']['pod2svc']['tcp_bw_mbit'])
    bar_chart(data=net_data,
              title='K8S CNI Benchmark - Pod to Service - TCP - Bandwidth',
              unit=net_unit,
              x_label=f" TCP Bandwidth in {net_unit} - Higher is better")
    save_name = os.path.join(out_dir, 'benchmark_net_p2svc_tcp.png')
    plt.savefig(save_name)

    # P2SVC UDP
    net_data = {}
    for s, l in zip(SUBJECTS, LEGEND_NAMES):
        net_data[l] = int(combined_results[s]['knb']['pod2svc']['udp_bw_mbit'])
    bar_chart(data=net_data,
              title='K8S CNI Benchmark - Pod to Service - UDP - Bandwidth',
              unit=net_unit,
              x_label=f" UDP Bandwidth in {net_unit} - Higher is better")
    save_name = os.path.join(out_dir, 'benchmark_net_p2svc_udp.png')
    plt.savefig(save_name)

    # FIO chart
    # Read IOPS
    fio_data = {}
    for s, l in zip(SUBJECTS, LEGEND_NAMES):
        fio_data[l] = int(combined_results[s]['fio']['read_iops']['iops'])
    bar_chart(data=fio_data,
              title='FIO Benchmark - Read - IOPS',
              x_label=f" Read {fio_iops_unit} - Higher is better")
    save_name = os.path.join(out_dir, 'benchmark_fio_read_iops.png')
    plt.savefig(save_name)

    # Write IOPS
    fio_data = {}
    for s, l in zip(SUBJECTS, LEGEND_NAMES):
        fio_data[l] = int(combined_results[s]['fio']['write_iops']['iops'])
    bar_chart(data=fio_data,
              title='FIO Benchmark - Write - IOPS',
              x_label=f" Write {fio_iops_unit} - Higher is better")
    save_name = os.path.join(out_dir, 'benchmark_fio_write_iops.png')
    plt.savefig(save_name)

    # Read Bandwidth
    fio_data = {}
    for s, l in zip(SUBJECTS, LEGEND_NAMES):
        fio_data[l] = int(combined_results[s]['fio']['read_bw']['bw_kbytes'])
    bar_chart(data=fio_data,
              title='FIO Benchmark - Read - Bandwidth',
              unit=fio_bw_unit,
              x_label=f" Read Bandwidth in {fio_bw_unit} - Higher is better")
    save_name = os.path.join(out_dir, 'benchmark_fio_read_bw.png')
    plt.savefig(save_name)

    # Write Bandwidth
    fio_data = {}
    for s, l in zip(SUBJECTS, LEGEND_NAMES):
        fio_data[l] = int(combined_results[s]['fio']['write_bw']['bw_kbytes'])
    bar_chart(data=fio_data,
              title='FIO Benchmark - Write - Bandwidth',
              unit=fio_bw_unit,
              x_label=f" Write Bandwidth in {fio_bw_unit} - Higher is better")
    save_name = os.path.join(out_dir, 'benchmark_fio_write_bw.png')
    plt.savefig(save_name)


if __name__ == '__main__':
    main()

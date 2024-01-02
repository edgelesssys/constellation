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

SUBJECTS_AZURE = ['constellation-azure', 'AKS']
SUBJECTS_GCP = ['constellation-gcp', 'GKE']
SUBJECTS_AWS = ['constellation-aws', 'EKS']

LEGEND_NAMES_AZURE = ['Constellation', 'AKS']
LEGEND_NAMES_GCP = ['Constellation', 'GKE']
LEGEND_NAMES_AWS = ['Constellation', 'EKS']


BAR_COLORS = ['#90FF99', '#929292', '#8B04DD', '#000000']

FONT_URL = "https://github.com/openmaptiles/fonts/raw/master/roboto/Roboto-Regular.ttf"
FONT_NAME = "Roboto-Regular.ttf"
FONT_SIZE = 13

# Some lookup dictionaries for x axis
fio_iops_unit = 'IOPS'
fio_bw_unit = 'MiB/s'

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

    for test in SUBJECTS_AZURE+SUBJECTS_GCP+SUBJECTS_AWS:
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
    # P2P TCP + UDP Azure
    net_data = {}
    for s, l in zip(SUBJECTS_AZURE, LEGEND_NAMES_AZURE):
        net_data[l+" - TCP"] = int(combined_results[s]
                                   ['knb']['pod2pod']['tcp_bw_mbit'])
    for s, l in zip(SUBJECTS_AZURE, LEGEND_NAMES_AZURE):
        net_data[l+" - UDP"] = int(combined_results[s]
                                   ['knb']['pod2pod']['udp_bw_mbit'])
    bar_chart(data=net_data,
              title='K8S CNI Benchmark - Pod to Pod - Azure - Bandwidth',
              unit=net_unit,
              x_label=f"Bandwidth in {net_unit} - Higher is better")
    save_name = os.path.join(out_dir, 'benchmark_net_p2p_azure.png')
    plt.savefig(save_name)

    # P2P TCP + UDP GCP
    net_data = {}
    for s, l in zip(SUBJECTS_GCP, LEGEND_NAMES_GCP):
        net_data[l+" - TCP"] = int(combined_results[s]
                                   ['knb']['pod2pod']['tcp_bw_mbit'])
    for s, l in zip(SUBJECTS_GCP, LEGEND_NAMES_GCP):
        net_data[l+" - UDP"] = int(combined_results[s]
                                   ['knb']['pod2pod']['udp_bw_mbit'])
    bar_chart(data=net_data,
              title='K8S CNI Benchmark - Pod to Pod - GCP - Bandwidth',
              unit=net_unit,
              x_label=f"Bandwidth in {net_unit} - Higher is better")
    save_name = os.path.join(out_dir, 'benchmark_net_p2p_gcp.png')
    plt.savefig(save_name)

    # P2P TCP + UDP AWS
    net_data = {}
    for s, l in zip(SUBJECTS_AWS, LEGEND_NAMES_AWS):
        net_data[l+" - TCP"] = int(combined_results[s]
                                   ['knb']['pod2pod']['tcp_bw_mbit'])
    for s, l in zip(SUBJECTS_AWS, LEGEND_NAMES_AWS):
        net_data[l+" - UDP"] = int(combined_results[s]
                                   ['knb']['pod2pod']['udp_bw_mbit'])
    bar_chart(data=net_data,
              title='K8S CNI Benchmark - Pod to Pod - AWS - Bandwidth',
              unit=net_unit,
              x_label=f"Bandwidth in {net_unit} - Higher is better")
    save_name = os.path.join(out_dir, 'benchmark_net_p2p_aws.png')
    plt.savefig(save_name)

    # P2SVC TCP + UDP Azure
    net_data = {}
    for s, l in zip(SUBJECTS_AZURE, LEGEND_NAMES_AZURE):
        net_data[l+" - TCP"] = int(combined_results[s]
                                   ['knb']['pod2svc']['tcp_bw_mbit'])
    for s, l in zip(SUBJECTS_AZURE, LEGEND_NAMES_AZURE):
        net_data[l+" - UDP"] = int(combined_results[s]
                                   ['knb']['pod2svc']['udp_bw_mbit'])
    bar_chart(data=net_data,
              title='K8S CNI Benchmark - Pod to Service - Azure - Bandwidth',
              unit=net_unit,
              x_label=f"Bandwidth in {net_unit} - Higher is better")
    save_name = os.path.join(out_dir, 'benchmark_net_p2svc_azure.png')
    plt.savefig(save_name)

    # P2SVC TCP + UDP GCP
    net_data = {}
    for s, l in zip(SUBJECTS_GCP, LEGEND_NAMES_GCP):
        net_data[l+" - TCP"] = int(combined_results[s]
                                   ['knb']['pod2svc']['tcp_bw_mbit'])
    for s, l in zip(SUBJECTS_GCP, LEGEND_NAMES_GCP):
        net_data[l+" - UDP"] = int(combined_results[s]
                                   ['knb']['pod2svc']['udp_bw_mbit'])
    bar_chart(data=net_data,
              title='K8S CNI Benchmark - Pod to Service - GCP - Bandwidth',
              unit=net_unit,
              x_label=f"Bandwidth in {net_unit} - Higher is better")
    save_name = os.path.join(out_dir, 'benchmark_net_p2svc_gcp.png')
    plt.savefig(save_name)

    # P2SVC TCP + UDP GCP
    net_data = {}
    for s, l in zip(SUBJECTS_AWS, LEGEND_NAMES_AWS):
        net_data[l+" - TCP"] = int(combined_results[s]
                                   ['knb']['pod2svc']['tcp_bw_mbit'])
    for s, l in zip(SUBJECTS_AWS, LEGEND_NAMES_AWS):
        net_data[l+" - UDP"] = int(combined_results[s]
                                   ['knb']['pod2svc']['udp_bw_mbit'])
    bar_chart(data=net_data,
              title='K8S CNI Benchmark - Pod to Service - AWS - Bandwidth',
              unit=net_unit,
              x_label=f"Bandwidth in {net_unit} - Higher is better")
    save_name = os.path.join(out_dir, 'benchmark_net_p2svc_aws.png')
    plt.savefig(save_name)

    # FIO charts

    # IOPS on Azure
    fio_data = {}
    for s, l in zip(SUBJECTS_AZURE, LEGEND_NAMES_AZURE):
        fio_data[l+" - Read"] = int(combined_results[s]
                                    ['fio']['read_iops']['iops'])
    for s, l in zip(SUBJECTS_AZURE, LEGEND_NAMES_AZURE):
        fio_data[l+" - Write"] = int(combined_results[s]
                                     ['fio']['write_iops']['iops'])
    bar_chart(data=fio_data,
              title='FIO Benchmark - Azure - IOPS',
              x_label=f"{fio_iops_unit} - Higher is better")
    save_name = os.path.join(out_dir, 'benchmark_fio_azure_iops.png')
    plt.savefig(save_name)

    # IOPS on GCP
    fio_data = {}
    for s, l in zip(SUBJECTS_GCP, LEGEND_NAMES_GCP):
        fio_data[l+" - Read"] = int(combined_results[s]
                                    ['fio']['read_iops']['iops'])
    for s, l in zip(SUBJECTS_GCP, LEGEND_NAMES_GCP):
        fio_data[l+" - Write"] = int(combined_results[s]
                                     ['fio']['write_iops']['iops'])
    bar_chart(data=fio_data,
              title='FIO Benchmark - GCP - IOPS',
              x_label=f"{fio_iops_unit} - Higher is better")
    save_name = os.path.join(out_dir, 'benchmark_fio_gcp_iops.png')
    plt.savefig(save_name)

    # IOPS on AWS
    fio_data = {}
    for s, l in zip(SUBJECTS_AWS, LEGEND_NAMES_AWS):
        fio_data[l+" - Read"] = int(combined_results[s]
                                    ['fio']['read_iops']['iops'])
    for s, l in zip(SUBJECTS_AWS, LEGEND_NAMES_AWS):
        fio_data[l+" - Write"] = int(combined_results[s]
                                     ['fio']['write_iops']['iops'])
    bar_chart(data=fio_data,
              title='FIO Benchmark - AWS - IOPS',
              x_label=f"{fio_iops_unit} - Higher is better")
    save_name = os.path.join(out_dir, 'benchmark_fio_aws_iops.png')
    plt.savefig(save_name)

    # Bandwidth on Azure
    fio_data = {}
    for s, l in zip(SUBJECTS_AZURE, LEGEND_NAMES_AZURE):
        fio_data[l+" - Read"] = int(combined_results[s]
                                    ['fio']['read_bw']['bw_kbytes'] / 1024)
    for s, l in zip(SUBJECTS_AZURE, LEGEND_NAMES_AZURE):
        fio_data[l+" - Write"] = int(combined_results[s]
                                     ['fio']['write_bw']['bw_kbytes'] / 1024)
    bar_chart(data=fio_data,
              title='FIO Benchmark - Azure - Bandwidth',
              x_label=f"Bandwidth in {fio_bw_unit} - Higher is better")
    save_name = os.path.join(out_dir, 'benchmark_fio_azure_bw.png')
    plt.savefig(save_name)

    # Bandwidth on GCP
    fio_data = {}
    for s, l in zip(SUBJECTS_GCP, LEGEND_NAMES_GCP):
        fio_data[l+" - Read"] = int(combined_results[s]
                                    ['fio']['read_bw']['bw_kbytes'] / 1024)
    for s, l in zip(SUBJECTS_GCP, LEGEND_NAMES_GCP):
        fio_data[l+" - Write"] = int(combined_results[s]
                                     ['fio']['write_bw']['bw_kbytes'] / 1024)
    bar_chart(data=fio_data,
              title='FIO Benchmark - GCP - Bandwidth',
              x_label=f"Bandwidth in {fio_bw_unit} - Higher is better")
    save_name = os.path.join(out_dir, 'benchmark_fio_gcp_bw.png')
    plt.savefig(save_name)

    # Bandwidth on AWS
    fio_data = {}
    for s, l in zip(SUBJECTS_AWS, LEGEND_NAMES_AWS):
        fio_data[l+" - Read"] = int(combined_results[s]
                                    ['fio']['read_bw']['bw_kbytes'] / 1024)
    for s, l in zip(SUBJECTS_AWS, LEGEND_NAMES_AWS):
        fio_data[l+" - Write"] = int(combined_results[s]
                                     ['fio']['write_bw']['bw_kbytes'] / 1024)
    bar_chart(data=fio_data,
              title='FIO Benchmark - AWS - Bandwidth',
              x_label=f"Bandwidth in {fio_bw_unit} - Higher is better")
    save_name = os.path.join(out_dir, 'benchmark_fio_aws_bw.png')
    plt.savefig(save_name)

if __name__ == '__main__':
    main()

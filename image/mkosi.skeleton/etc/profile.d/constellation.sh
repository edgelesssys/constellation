#!/usr/bin/env bash
# Copyright (c) Edgeless Systems GmbH
#
# SPDX-License-Identifier: AGPL-3.0-only

set -euo pipefail
shopt -s inherit_errexit

export PATH=/run/state/bin:${PATH}
export KUBECONFIG=/etc/kubernetes/admin.conf
alias k=kubectl

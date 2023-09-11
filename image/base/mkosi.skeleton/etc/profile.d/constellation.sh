#!/usr/bin/env bash
# Copyright (c) Edgeless Systems GmbH
#
# SPDX-License-Identifier: AGPL-3.0-only

# Note: This script is sourced.

export TERM=linux
export PATH=/run/state/bin:${PATH}
export KUBECONFIG=/etc/kubernetes/admin.conf
alias k=kubectl

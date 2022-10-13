#!/usr/bin/env python
# Copyright (c) Edgeless Systems GmbH
#
# SPDX-License-Identifier: AGPL-3.0-only

# This script calculates the authentihash of a PE / EFI binary.
# Install prerequisites:
#   pip install lief

import sys
import lief

def authentihash(filename):
    pe = lief.parse(filename)
    return pe.authentihash(lief.PE.ALGORITHMS.SHA_256)

if __name__ == '__main__':
    if len(sys.argv) != 2:
        print(f"Usage: {sys.argv[0]} <filename>")
        sys.exit(1)
    print(authentihash(sys.argv[1]).hex())

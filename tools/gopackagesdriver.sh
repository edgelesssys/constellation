#!/usr/bin/env bash
exec bazel run --config linux_amd64_static -- @io_bazel_rules_go//go/tools/gopackagesdriver "${@}"

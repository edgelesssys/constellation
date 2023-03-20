#!/usr/bin/env bash

# Compare licenses of Go dependencies against a whitelist.

set -euo pipefail
shopt -s inherit_errexit

not_allowed() {
  echo "license not allowed for package: ${line}"
  err=1
}

err=0
go mod download

go-licenses csv ./... | {
  while read -r line; do

    pkg=${line%%,*}
    lic=${line##*,}

    case ${lic} in
    Apache-2.0 | BSD-2-Clause | BSD-3-Clause | ISC | MIT) ;;

    MPL-2.0)
      case ${pkg} in
      github.com/siderolabs/talos/pkg/machinery/config/encoder) ;;

      github.com/letsencrypt/boulder) ;;

      github.com/hashicorp/*) ;;

      *)
        not_allowed
        ;;
      esac
      ;;

    AGPL-3.0)
      case ${pkg} in
      github.com/edgelesssys/constellation/v2) ;;

      github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/v2/api/v1alpha1) ;;

      *)
        not_allowed
        ;;
      esac
      ;;

    Unknown)
      case ${pkg} in
      github.com/edgelesssys/go-tdx-qpl/*) ;;

      *)
        not_allowed
        ;;
      esac
      ;;

    *)
      echo "unknown license: ${line}"
      err=1
      ;;
    esac

  done
  exit "${err}"
}

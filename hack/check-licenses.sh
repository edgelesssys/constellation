#!/bin/bash

# Compare licenses of Go dependencies against a whitelist.

set -e -o pipefail

not_allowed() {
  echo "license not allowed for package: $line"
  err=1
}

go mod download

go-licenses csv ./... | {
while read line; do

  pkg=${line%%,*}
  lic=${line##*,}

  case $lic in
    Apache-2.0|BSD-2-Clause|BSD-3-Clause|ISC|MIT)
      ;;

    MPL-2.0)
      case $pkg in
        github.com/talos-systems/talos/pkg/machinery/config/encoder)
          ;;
        github.com/letsencrypt/boulder)
          ;;
        *)
          not_allowed
          ;;
      esac
      ;;

    AGPL-3.0)
      case $pkg in
        github.com/edgelesssys/constellation/v2)
          ;;
        *)
          not_allowed
          ;;
      esac
      ;;

    Unknown)
      case $pkg in
        *)
          not_allowed
          ;;
      esac
      ;;

    *)
      echo "unknown license: $line"
      err=1
      ;;
  esac

done
exit $err
}

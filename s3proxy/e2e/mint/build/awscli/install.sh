#!/bin/bash -e
#
#  Mint (C) 2017-2022 Minio, Inc.
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.
#

die() {
  echo "$*" 1>&2
  exit 1
}

# shellcheck disable=SC2086
ROOTDIR="$(dirname "$(realpath $0)")"
TMPDIR="$(mktemp -d)"

cd "$TMPDIR"

# Download botocore and apply @y4m4's expect 100 continue fix
(git clone --depth 1 -b 1.31.37 https://github.com/boto/botocore &&
  cd botocore &&
  patch -p1 < "$ROOTDIR/expect-100.patch" &&
  python3 -m pip install .) ||
  die "Unable to install botocore.."

# Download and install aws cli
(git clone --depth 1 -b 1.29.37 https://github.com/aws/aws-cli &&
  cd aws-cli &&
  python3 -m pip install .) ||
  die "Unable to install aws-cli.."

# Clean-up
rm -r "$TMPDIR"

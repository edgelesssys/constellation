#!/bin/sh

# The charon binary is not included in the PATH generated by nixery.dev, find it manually.
charon="$(dirname "$(readlink -f "$(command -v charon-systemd)")")/../libexec/ipsec/charon"

"${charon}" &

while ! swanctl --stats > /dev/null 2> /dev/null; do
  sleep 1
done
swanctl --load-all

wait

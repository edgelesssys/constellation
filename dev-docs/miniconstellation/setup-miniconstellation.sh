#!/usr/bin/env bash

echo "Installing Constellation CLI"
curl -LO https://github.com/edgelesssys/constellation/releases/latest/download/constellation-linux-amd64
sudo install constellation-linux-amd64 /usr/local/bin/constellation

echo "Waiting for docker service to be active..."
# Wait at most 20min
count=0
until systemctl is-active docker || [[ ${count} -eq 120 ]]; do
  sleep 10
  count=$((count + 1))
done
if [[ ${count} -eq 120 ]]; then
  echo "Docker service did not come up in time."
  exit 1
fi

# change to workspace
mkdir constellation_workspace
cd constellation_workspace

# takes around 15 minutes
constellation mini up
echo "Cluster creation done."

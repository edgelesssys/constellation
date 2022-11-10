#!/usr/bin/env bash
# source https://learn.microsoft.com/en-us/azure/virtual-machines/linux/no-agent

set -euo pipefail
shopt -s inherit_errexit

attempts=1
until [[ ${attempts} -gt 5 ]]; do
  echo "obtaining goal state - attempt ${attempts}"
  goalstate=$(curl --fail -v -X 'GET' -H "x-ms-agent-name: azure-vm-register" \
    -H "Content-Type: text/xml;charset=utf-8" \
    -H "x-ms-version: 2012-11-30" \
    "http://168.63.129.16/machine/?comp=goalstate")
  if [[ $? -eq 0 ]]; then
    echo "successfully retrieved goal state"
    retrieved_goal_state=true
    break
  fi
  sleep 5
  attempts=$((attempts + 1))
done

if [[ ${retrieved_goal_state} != "true" ]]; then
  echo "failed to obtain goal state - cannot register this VM"
  exit 1
fi

container_id=$(grep ContainerId <<< "${goalstate}" | sed 's/\s*<\/*ContainerId>//g' | sed 's/\r$//')
instance_id=$(grep InstanceId <<< "${goalstate}" | sed 's/\s*<\/*InstanceId>//g' | sed 's/\r$//')

ready_doc=$(
  cat << EOF
<?xml version="1.0" encoding="utf-8"?>
<Health xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema">
  <GoalStateIncarnation>1</GoalStateIncarnation>
  <Container>
    <ContainerId>${container_id}</ContainerId>
    <RoleInstanceList>
      <Role>
        <InstanceId>${instance_id}</InstanceId>
        <Health>
          <State>Ready</State>
        </Health>
      </Role>
    </RoleInstanceList>
  </Container>
</Health>
EOF
)

attempts=1
until [[ ${attempts} -gt 5 ]]; do
  echo "registering with Azure - attempt ${attempts}"
  curl --fail -v -X 'POST' -H "x-ms-agent-name: azure-vm-register" \
    -H "Content-Type: text/xml;charset=utf-8" \
    -H "x-ms-version: 2012-11-30" \
    -d "${ready_doc}" \
    "http://168.63.129.16/machine?comp=health"
  if [[ $? -eq 0 ]]; then
    echo "successfully register with Azure"
    break
  fi
  sleep 5 # sleep to prevent throttling from wire server
done

#!/usr/bin/env bash

set +e
echo "Waiting for SSH server to come online..."

# Wait for SSH to come online, at most 10*30s=5min
COUNT=0
until ssh -i id_rsa -o StrictHostKeyChecking=no adminuser@"${AZURE_VM_IP}" date || [ $COUNT -eq 10 ]; do
  sleep 30
  ((COUNT++))
done

echo "Done waiting."

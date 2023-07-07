#!/usr/bin/env bash

echo "create Terraform resources"

terraform init
terraform apply -auto-approve
terraform output -raw ssh_private_key > id_rsa
chmod 600 id_rsa

azure_vm_ip=$(terraform output -raw public_ip)

echo "::endgroup::"

echo "Waiting for SSH server to come online..."

# Wait for SSH to come online, at most 10*30s=5min
count=0
until ssh -i id_rsa -o StrictHostKeyChecking=no adminuser@"${azure_vm_ip}" date || [[ ${count} -eq 10 ]]; do
  sleep 30
  count=$((count + 1))
done

echo "Done waiting."

echo "Copy prep VM script to remote VM"
scp -i id_rsa ../setup-miniconstellation.sh adminuser@"${azure_vm_ip}":~/setup-miniconstellation.sh

echo "Logging into remote VM"
ssh -i id_rsa adminuser@"${azure_vm_ip}"

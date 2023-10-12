#!/usr/bin/env bash

# Validate inputs and environment variables.
if [[ ! "$1" =~ ^ghcr.io/edgelesssys/constellation/mint:v.*$ ]]; then
  echo "Error: invalid tag, expected input to match pattern '^ghcr.io\/edgelesssys\/constellation\/mint:v*$'"
  exit 1
fi
mint_image=$1

if [[ -z "$KUBECONFIG" ]]; then
  echo "Error: KUBECONFIG environment variable not set"
  exit 1
fi

if [[ -z "$AWS_ACCESS_KEY_ID" ]]; then
  echo "Error: AWS_ACCESS_KEY_ID environment variable not set"
  exit 1
fi

if [[ -z "$AWS_SECRET_ACCESS_KEY" ]]; then
  echo "Error: AWS_SECRET_ACCESS_KEY environment variable not set"
  exit 1
fi

# Wait for the s3proxy pod to be created. kubectl wait can not wait for resources to be created.
start_time=$(date +%s)
timeout=300
while true; do
    if [[ -n "$(kubectl get svc -l app=s3proxy -o jsonpath='{.items[*]}')" ]]; then
        echo "Service with label app=s3proxy found"
        break
    else
        current_time=$(date +%s)
        elapsed_time=$((current_time - start_time))
        if [[ $elapsed_time -ge $timeout ]]; then
            echo "Timeout waiting for service with label app=s3proxy"
            exit 1
        else
            echo "Waiting for service with label app=s3proxy"
            sleep 5
        fi
    fi
done

# Wait until pod becomes ready.
kubectl wait --for=condition=Ready --timeout=2m pod -l app=s3proxy

# Get the CA that signed the s3proxy's TLS certificate.
kubectl get secret s3proxy-tls -o yaml | yq e '.data."ca.crt"' - | base64 -d > s3proxy-ca.crt
if [[ "$?" -ne 0 ]]; then
  echo "Error: failed to get s3proxy-tls secret"
  exit 1
fi

# block for sudoers password so it is not requested for following command, which is put into the background.
sudo -v

port_forward_ip="172.30.0.1"
# sudo -E kubectl port-forward --address "$port_forward_ip" svc/s3proxy-service 443:443 &
sudo -E -b kubectl port-forward --address "$port_forward_ip" svc/s3proxy-service 443:443 > ./port-forward.log 2>&1

# Kill port-forward on exit and print it's output.
trap 'stop_port_forward "$port_forward_ip"' EXIT


# wait for port-forward to be ready
start_time=$(date +%s)
timeout=300
while true; do
    # Check if a connection to the s3proxy can be established. Trust certificates signed by the CA fetched earlier.
    echo | openssl s_client -connect "$port_forward_ip":443 -brief -verify_return_error -CAfile s3proxy-ca.crt
    if [[ "$?" -eq 0 ]]; then
        echo "Port-forward ready"
        break
    else
        current_time=$(date +%s)
        elapsed_time=$((current_time - start_time))
        if [[ $elapsed_time -ge $timeout ]]; then
            echo "Timeout waiting for port-forward"
            exit 1
        else
            echo "Waiting for port-forward"
            sleep 5
        fi
    fi
done

docker run -v $PWD/s3proxy-ca.crt:/etc/ssl/certs/kube-ca.crt \
  --add-host s3.eu-west-1.amazonaws.com:"$port_forward_ip" \
  -e AWS_CA_BUNDLE=/etc/ssl/certs/kube-ca.crt \
  -e SERVER_REGION=eu-west-1 \
  -e SERVER_ENDPOINT=s3.eu-west-1.amazonaws.com:443 \
  -e ACCESS_KEY=$AWS_ACCESS_KEY_ID \
  -e SECRET_KEY=$AWS_SECRET_ACCESS_KEY \
  -e ENABLE_HTTPS=1 \
  "$mint_image" aws-sdk-go versioning

function stop_port_forward() {
  local port_forward_ip="$1"
  sudo pkill -f "kubectl port-forward --address $port_forward_ip svc/s3proxy-service 443:443"
  echo "Output of kubectl port-forward:"
  cat ./port-forward.log
}

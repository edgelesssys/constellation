#!/bin/bash

function terminate_mint() {
  kubectl logs job/mint-deploy
  kubectl delete job mint-deploy
}


if [[ ! "$1" =~ ^ghcr.io/edgelesssys/constellation/mint:v.*$ ]]; then
  echo "Error: invalid tag, expected input to match pattern '^ghcr.io\/edgelesssys\/constellation\/mint:v*$'"
  exit 1
fi

if [[ -z "$KUBECONFIG" ]]; then
  echo "Error: KUBECONFIG environment variable not set"
  exit 1
fi

if [[ -z "$ACCESS_KEY" ]]; then
  echo "Error: ACCESS_KEY environment variable not set"
  exit 1
fi

if [[ -z "$SECRET_KEY" ]]; then
  echo "Error: SECRET_KEY environment variable not set"
  exit 1
fi

# Wait for the s3proxy service to be created. kubectl wait can not wait for resources to be created.
start_time=$(date +%s)
timeout=300
while true; do
    if [[ -n "$(kubectl get svc -l app=s3proxy -o jsonpath='{.items[*]}')" ]]; then
        echo "Service with label app=s3proxy found"
        service_ip=$(kubectl get svc s3proxy-service -o=jsonpath='{.spec.clusterIP}')
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

kubectl delete job mint-deploy --ignore-not-found=true

cat << EOF | kubectl apply -f -
apiVersion: batch/v1
kind: Job
metadata:
  name: mint-deploy
spec:
  template:
    metadata:
      name: mint-deploy
    spec:
      restartPolicy: Never
      hostAliases:
      - ip: "$service_ip"
        hostnames:
        - "s3.eu-west-1.amazonaws.com"
      containers:
      - name: mint
        image: "$1"
        args:
        - "aws-sdk-go"
        - "versioning"
        volumeMounts:
        - name: ca-cert
          mountPath: /etc/ssl/certs/kube-ca.crt
          subPath: kube-ca.crt
        env:
        - name: SERVER_REGION
          value: eu-west-1
        - name: SERVER_ENDPOINT
          value: s3.eu-west-1.amazonaws.com:443
        - name: ENABLE_HTTPS
          value: "1"
        - name: AWS_CA_BUNDLE
          value: /etc/ssl/certs/kube-ca.crt
        - name: ACCESS_KEY
          value: "$ACCESS_KEY"
        - name: SECRET_KEY
          value: "$SECRET_KEY"
      volumes:
      - name: ca-cert
        secret:
            secretName: s3proxy-tls
            items:
            - key: ca.crt
              path: kube-ca.crt
EOF

# Remove job before this script finishes.
trap "terminate_mint" EXIT

# Tests have to complete within 10 minutes, otherwise they have failed.
if kubectl wait --for=condition=complete job/mint-deploy --timeout=600s; then
  echo "Mint tests completed successfully"
  exit 0
else
  echo "Mint tests failed"
  exit 1
fi

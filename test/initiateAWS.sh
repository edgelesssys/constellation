#! /bin/bash

echo "aws_access_key_id = $aws_access_key_id"  >>  /root/.aws/credentials
echo "aws_secret_access_key = $aws_secret_access_key"  >>  /root/.aws/credentials

echo "$GITHUB_TOKEN" >> /root/.netrc


terminate() {
  err=$?
  pkill kubectl
  ./constellation terminate
  exit $err
}
trap terminate ERR


BRANCH="${BRANCH:-main}"
git clone --branch $BRANCH https://github.com/edgelesssys/constellation-coordinator.git
mkdir -p constellation-coordinator/build
cd constellation-coordinator/build
cmake ..
make -j"$(nproc)" cli

./constellation create aws 2 4xlarge -y

echo "Initializing constellation"
./constellation init --privatekey /privatekey

export KUBECONFIG="./admin.conf"
kubectl wait --for=condition=ready --timeout=60s --all nodes

kubectl apply -k github.com/BuoyantIO/emojivoto/kustomize/deployment

echo "Wait for service to be applied"
kubectl wait --for=condition=available --timeout=60s -n emojivoto --all deployments

kubectl -n emojivoto port-forward svc/web-svc 8080:80 &

sleep 5

curl http://localhost:8080

terminate

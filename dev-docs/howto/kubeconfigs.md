# How to create kubeconfigs for users

One of the first things to do after setting up a Constellation cluster is to hand out kubeconfig files to its prospective users.
Adhering to the *principle of least privilege*, it is not advisable to share the admin config with all cluster users.
Instead, users should authenticate individually to the API server, and permissions should be controlled by [RBAC].

Constellation users authenticate to the API server with a client TLS certificate, signed by the Kubernetes CA.
The user's identity and group memberships are taken from the certificates common name and organizations, respectively.
Details can be found in the upstream [authn documentation].

The [`kubeadm` documentation] describes a process for creating new kubeconfigs, but the instructions requires access to a control-plane node, or at least the Kubernetes CA certificate and key.
While the certificates can be extracted, e.g. by spawning a [node debugger pod], we can take a safer road that only requires `kubectl`.
The example script below creates a new kubeconfig for a user and optional group memberships.
It uses the [Kubernetes certificate API] to obtain a user certificate signed by the cluster CA.

[RBAC]: https://kubernetes.io/docs/reference/access-authn-authz/rbac/
[authn documentation]: https://kubernetes.io/docs/reference/access-authn-authz/authentication/#users-in-kubernetes
[`kubeadm` documentation]: https://kubernetes.io/docs/tasks/administer-cluster/kubeadm/kubeadm-certs/#kubeconfig-additional-users
[node debugger pod]: https://kubernetes.io/docs/tasks/debug/debug-cluster/kubectl-node-debug/
[Kubernetes certificate API]: https://kubernetes.io/docs/tasks/tls/managing-tls-in-a-cluster/

```sh
#!/bin/sh
set -eu

if [ $# -lt 2 ]; then
    echo "Usage: $0 username [groupname...]" >&2
    exit 1
fi

user=$1
shift

subj="/CN=${user}"
for g in "$@"; do
  subj="${subj}/O=$g"
done

openssl req -newkey rsa:4096 -out ${user}.csr -keyout ${user}.key -nodes -subj "${subj}"

kubectl apply -f - <<EOF
apiVersion: certificates.k8s.io/v1
kind: CertificateSigningRequest
metadata:
  name: ${user}
spec:
  request: $(base64 -w0 ${user}.csr)
  signerName: kubernetes.io/kube-apiserver-client
  usages:
  - digital signature
  - key encipherment
  - client auth
EOF
kubectl certificate approve ${user}
kubectl wait --for=jsonpath='{.status.certificate}' csr/${user}
kubectl get csr ${user} -o jsonpath='{.status.certificate}' | base64 -d >${user}.pem
kubectl delete csr ${user}

kubectl get cm kube-root-ca.crt -o go-template='{{ index .data "ca.crt" }}' >ca.pem
kubectl get cm kubeadm-config -n kube-system -o=jsonpath="{.data.ClusterConfiguration}" >clusterconfig.yaml
cluster=$(yq .clusterName clusterconfig.yaml)
endpoint=$(yq .controlPlaneEndpoint clusterconfig.yaml)

cat >${user}.conf <<EOF
apiVersion: v1
kind: Config
clusters:
- cluster:
    certificate-authority-data: $(base64 -w0 ca.pem)
    server: https://${endpoint}
  name: ${cluster}
contexts:
- context:
    cluster: ${cluster}
    user: ${user}
  name: ${user}@${cluster}
current-context: ${user}@${cluster}
users:
- name: ${user}
  user:
    client-certificate-data: $(base64 -w0 ${user}.pem)
    client-key-data: $(base64 -w0 ${user}.key)
EOF

env KUBECONFIG=./${user}.conf kubectl auth whoami

rm ca.pem clusterconfig.yaml ${user}.csr ${user}.pem ${user}.key
```

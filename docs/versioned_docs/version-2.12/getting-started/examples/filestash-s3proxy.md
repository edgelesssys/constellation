
# Deploying Filestash

Filestash is a web frontend for different storage backends, including S3.
It's a useful application to showcase s3proxy in action.

1. Deploy s3proxy as described in [Deployment](../../workflows/s3proxy.md#deployment).
2. Create a deployment file for Filestash with one pod:

```sh
cat << EOF > "deployment-filestash.yaml"
apiVersion: apps/v1
kind: Deployment
metadata:
    name: filestash
spec:
    replicas: 1
    selector:
        matchLabels:
            app: filestash
    template:
        metadata:
            labels:
                app: filestash
        spec:
          hostAliases:
          - ip: $(kubectl get svc s3proxy-service -o=jsonpath='{.spec.clusterIP}')
            hostnames:
            - "s3.eu-west-1.amazonaws.com"
          containers:
          - name: filestash
            image: machines/filestash:latest
            ports:
            - containerPort: 8334
            volumeMounts:
            - name: ca-cert
              mountPath: /etc/ssl/certs/kube-ca.crt
              subPath: kube-ca.crt
          volumes:
          - name: ca-cert
            secret:
              secretName: s3proxy-tls
              items:
              - key: ca.crt
                path: kube-ca.crt
EOF
```

The pod spec includes the `hostAliases` key, which adds an entry to the pod's `/etc/hosts`.
The entry forwards all requests for `s3.eu-west-1.amazonaws.com` to the Kubernetes service `s3proxy-service`.
If you followed the s3proxy [Deployment](../../workflows/s3proxy.md#deployment) guide, this service points to a s3proxy pod.

To use other regions than `eu-west-1`, add more entries to `hostAliases` for all regions you require.
Use the same IP for those entries. For example to add `us-east-1` add:

```yaml
- ip: $(kubectl get svc s3proxy-service -o=jsonpath='{.spec.clusterIP}')
  hostnames:
  - "s3.us-east-1.amazonaws.com"
```

The spec also includes a volume mount for the TLS certificate and adds it to the pod's certificate trust store.
The volume is called `ca-cert`.
The key `ca.crt` of that volume is mounted to `/etc/ssl/certs/kube-ca.crt`, which is the default certificate trust store location for that container's OpenSSL library.
Not adding the CA certificate will result in TLS authentication errors.

3. Apply the file: `kubectl apply -f deployment-filestash.yaml`

Afterward, you can use a port forward to access the Filestash pod:
`kubectl port-forward pod/$(kubectl get pod --selector='app=filestash' -o=jsonpath='{.items[*].metadata.name}') 8334:8334`

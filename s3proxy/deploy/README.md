# Deploying s3proxy

**Caution:** Using s3proxy outside Constellation is insecure as the connection between the key management service (KMS) and s3proxy is protected by Constellation's WireGuard VPN.
The VPN is a feature of Constellation and will not be present by default in other environments.

Disclaimer: the following steps will be automated next.

- Run `bazel run //bazel/release:s3proxy_push`
- Set `IMAGE` to the newly built s3proxy image.
- `helm install s3proxy --set awsAccessKeyID="$AWS_ACCESS_KEY_ID" --set awsSecretAccessKey="$AWS_SECRET_ACCESS_KEY" --set image="$IMAGE"  ./s3proxy`

# Deploying Filestash

Filestash is a demo application that can be used to see s3proxy in action.
To deploy Filestash, first deploy s3proxy as described above.
Then run the below commands:

```sh
$ cat << EOF > "deployment-filestash.yaml"
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

$ kubectl apply -f deployment-filestash.yaml
```

Afterwards you can use a port forward to access the Filestash pod:

- `kubectl port-forward pod/$(kubectl get pod --selector='app=filestash' -o=jsonpath='{.items[*].metadata.name}') 8443:8443`


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
            - "s3.us-east-1.amazonaws.com"
            - "s3.us-east-2.amazonaws.com"
            - "s3.us-west-1.amazonaws.com"
            - "s3.us-west-2.amazonaws.com"
            - "s3.eu-north-1.amazonaws.com"
            - "s3.eu-south-1.amazonaws.com"
            - "s3.eu-south-2.amazonaws.com"
            - "s3.eu-west-1.amazonaws.com"
            - "s3.eu-west-2.amazonaws.com"
            - "s3.eu-west-3.amazonaws.com"
            - "s3.eu-central-1.amazonaws.com"
            - "s3.eu-central-2.amazonaws.com"
            - "s3.ap-northeast-1.amazonaws.com"
            - "s3.ap-northeast-2.amazonaws.com"
            - "s3.ap-northeast-3.amazonaws.com"
            - "s3.ap-east-1.amazonaws.com"
            - "s3.ap-southeast-1.amazonaws.com"
            - "s3.ap-southeast-2.amazonaws.com"
            - "s3.ap-southeast-3.amazonaws.com"
            - "s3.ap-southeast-4.amazonaws.com"
            - "s3.ap-south-1.amazonaws.com"
            - "s3.ap-south-2.amazonaws.com"
            - "s3.me-south-1.amazonaws.com"
            - "s3.me-central-1.amazonaws.com"
            - "s3.il-central-1.amazonaws.com"
            - "s3.af-south-1.amazonaws.com"
            - "s3.ca-central-1.amazonaws.com"
            - "s3.sa-east-1.amazonaws.com"
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
The entry forwards all requests for any of the currently defined AWS regions to the Kubernetes service `s3proxy-service`.
If you followed the s3proxy [Deployment](../../workflows/s3proxy.md#deployment) guide, this service points to a s3proxy pod.

The deployment specifies all regions explicitly to prevent accidental data leaks.
If one of your buckets were located in a region that's not part of the `hostAliases` key, traffic towards those buckets would not be redirected to s3proxy.
Similarly, if you want to exclude data for specific regions from going through s3proxy you can remove those regions from the deployment.

The spec also includes a volume mount for the TLS certificate and adds it to the pod's certificate trust store.
The volume is called `ca-cert`.
The key `ca.crt` of that volume is mounted to `/etc/ssl/certs/kube-ca.crt`, which is the default certificate trust store location for that container's OpenSSL library.
Not adding the CA certificate will result in TLS authentication errors.

3. Apply the file: `kubectl apply -f deployment-filestash.yaml`

Afterward, you can use a port forward to access the Filestash pod:
`kubectl port-forward pod/$(kubectl get pod --selector='app=filestash' -o=jsonpath='{.items[*].metadata.name}') 8334:8334`

4. After browsing to `localhost:8443`, Filestash will ask you to set an administrator password.
After setting it, you can directly leave the admin area by clicking the blue cloud symbol in the top left corner.
Subsequently, you can select S3 as storage backend and enter your credentials.
This will bring you to an overview of your buckets.
If you want to deploy Filestash in production, take a look at its [documentation](https://www.filestash.app/docs/).

5. To see the logs of s3proxy intercepting requests made to S3, run: `kubectl logs -f pod/$(kubectl get pod --selector='app=s3proxy' -o=jsonpath='{.items[*].metadata.name}')`
Look out for log messages labeled `intercepting`.
There is one such log message for each message that's encrypted, decrypted, or blocked.

6. Once you have uploaded a file with Filestash, you should be able to view the file in Filestash.
However, if you go to the AWS S3 [Web UI](https://s3.console.aws.amazon.com/s3/home) and download the file you just uploaded in Filestash, you won't be able to read it.
Another way to spot encrypted files without downloading them is to click on a file, scroll to the Metadata section, and look for the header named `x-amz-meta-constellation-encryption`.
This header holds the encrypted data encryption key of the object and is only present on objects that are encrypted by s3proxy.

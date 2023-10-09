# Install s3proxy

Constellation includes a transparent encryption proxy for [AWS S3](https://aws.amazon.com/de/s3/).
s3proxy will encrypt/decrypt objects when you send/retrieve objects to/from S3, without making changes to your application necessary.

## Limitations

:::caution
Using s3proxy outside Constellation is insecure as the connection between the key management service (KMS) and s3proxy is protected by Constellation's WireGuard VPN.
The VPN is a feature of Constellation and won't be present by default in other environments.
:::

These limitations will be removed with future iterations of s3proxy.

- Only `PutObject` and `GetObject` requests are encrypted/decrypted by s3proxy.
By default s3proxy will block requests that may leak data to S3 (e.g. UploadPart).
The `allow-multipart` flag disables request blocking for testing.
Using this flag will leak data if your application uses multipart uploads.
- Using the [Range](https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetObject.html#API_GetObject_RequestSyntax) header on `GetObject` is currently not supported and will result in an error.

If you want to use s3proxy but these limitations stop you from doing so, please consider [opening](https://github.com/edgelesssys/constellation/issues/new?assignees=&labels=&projects=&template=feature_request.yml) an issue.

## Deployment

- `wget https://raw.githubusercontent.com/edgelesssys/constellation/main/s3proxy/deploy/deployment-s3proxy.yaml`
- Replace the values named `replaceme` in `deployment-s3proxy.yaml` with valid AWS credentials. These credentials are used by s3proxy to access your S3 buckets.
- `kubectl apply -f deployment-s3proxy.yaml`

s3proxy is now deployed.
If you want to run a demo application, checkout the [Filestash with s3proxy](../getting-started/examples/filstash-s3proxy.md) example.


## Technical details

### Encryption

s3proxy relies on Google's [Tink Cryptographic Library](https://developers.google.com/tink) to implement cryptographic operations securely.
The used cryptographic primitives are [NIST SP 800 38f](https://nvlpubs.nist.gov/nistpubs/SpecialPublications/NIST.SP.800-38F.pdf) for key wrapping and [AES](https://en.wikipedia.org/wiki/Advanced_Encryption_Standard)-[GCM](https://en.wikipedia.org/wiki/Block_cipher_mode_of_operation#Galois/counter_(GCM)) with 256 bit keys for data encryption.

s3proxy uses [envelope encryption](https://cloud.google.com/kms/docs/envelope-encryption) to encrypt objects.
That means s3proxy uses a key encryption key (KEK) issued by the Constellation KMS to encrypt data encryption keys (DEK).
Each S3 object is encrypted with its own DEK.
The encrypted DEK is then saved as metadata of the encrypted object.
This enables key rotation of the KEK without re-encrypting the data in S3.
The approach also allows access to objects from different locations, as long as each location has access to the KEK.

### Traffic interception

To use s3proxy you have to redirect your outbound AWS S3 traffic to s3proxy.
This can either be done by modifying your client application or by changing the deployment of your application.

The necessary deployment modifications are to add DNS redirection and a trusted TLS certificate to the client's trust store.
DNS redirection can be defined for each pod, allowing you to test s3proxy for one application without changing other applications in the same cluster.
Adding a trusted TLS certificate is necessary as clients communicate with s3proxy via HTTPS.
To have your client application trust s3proxy's TLS certificate, the certificate has to be added to the client's certificate trust store.
The above [deployment example](#deployment) shows how all this can be done using cert-manager and the [hostAliases](https://kubernetes.io/docs/tasks/network/customize-hosts-file-for-pods/) key.

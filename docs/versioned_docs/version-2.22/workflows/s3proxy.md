# Install s3proxy

Constellation includes a transparent client-side encryption proxy for [AWS S3](https://aws.amazon.com/de/s3/) and compatible stores.
s3proxy encrypts objects before sending them to S3 and automatically decrypts them on retrieval, without requiring changes to your application.
With s3proxy, you can use S3 for storage in a confidential way without having to trust the storage provider.

## Limitations

Currently, s3proxy has the following limitations:
- Only `PutObject` and `GetObject` requests are encrypted/decrypted by s3proxy.
By default, s3proxy will block requests that may expose unencrypted data to S3 (e.g. UploadPart).
The `allow-multipart` flag disables request blocking for evaluation purposes.
- Using the [Range](https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetObject.html#API_GetObject_RequestSyntax) header on `GetObject` is currently not supported and will result in an error.

These limitations will be removed with future iterations of s3proxy.
If you want to use s3proxy but these limitations stop you from doing so, consider [opening an issue](https://github.com/edgelesssys/constellation/issues/new?assignees=&labels=&projects=&template=feature_request.yml).

## Deployment

You can add the s3proxy to your Constellation cluster as follows:
1. Add the Edgeless Systems chart repository:
   ```bash
   helm repo add edgeless https://helm.edgeless.systems/stable
   helm repo update
   ```
2. Set ACCESS_KEY and ACCESS_SECRET to valid credentials you want s3proxy to use to interact with S3.
3. Deploy s3proxy:
   ```bash
   helm install s3proxy edgeless/s3proxy --set awsAccessKeyID="$ACCESS_KEY" --set awsSecretAccessKey="$ACCESS_SECRET"
   ```

If you want to run a demo application, check out the [Filestash with s3proxy](../getting-started/examples/filestash-s3proxy.md) example.


## Technical details

### Encryption

s3proxy relies on Google's [Tink Cryptographic Library](https://developers.google.com/tink) to implement cryptographic operations securely.
The used cryptographic primitives are [NIST SP 800 38f](https://nvlpubs.nist.gov/nistpubs/SpecialPublications/NIST.SP.800-38F.pdf) for key wrapping and [AES](https://en.wikipedia.org/wiki/Advanced_Encryption_Standard)-[GCM](https://en.wikipedia.org/wiki/Block_cipher_mode_of_operation#Galois/counter_(GCM)) with 256 bit keys for data encryption.

s3proxy uses [envelope encryption](https://cloud.google.com/kms/docs/envelope-encryption) to encrypt objects.
This means s3proxy uses a key encryption key (KEK) issued by the [KeyService](../architecture/microservices.md#keyservice) to encrypt data encryption keys (DEKs).
Each S3 object is encrypted with its own DEK.
The encrypted DEK is then saved as metadata of the encrypted object.
This enables key rotation of the KEK without re-encrypting the data in S3.
The approach also allows access to objects from different locations, as long as each location has access to the KEK.

### Traffic interception

To use s3proxy, you have to redirect your outbound S3 traffic to s3proxy.
This can either be done by modifying your client application or by changing the deployment of your application.

The necessary deployment modifications are to add DNS redirection and a trusted TLS certificate to the client's trust store.
DNS redirection can be defined for each pod, allowing you to use s3proxy for one application without changing other applications in the same cluster.
Adding a trusted TLS certificate is necessary as clients communicate with s3proxy via HTTPS.
To have your client application trust s3proxy's TLS certificate, the certificate has to be added to the client's certificate trust store.
The [Filestash with s3proxy](../getting-started/examples/filestash-s3proxy.md) example shows how to do this.

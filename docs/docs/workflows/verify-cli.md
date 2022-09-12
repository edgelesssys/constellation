# Verify the CLI

Edgeless Systems uses [sigstore](https://www.sigstore.dev/) to ensure supply-chain security for the Constellation CLI and node images ("artifacts"). sigstore consists of three components: [Cosign](https://docs.sigstore.dev/cosign/overview), [Rekor](https://docs.sigstore.dev/rekor/overview), and Fulcio. Edgeless Systems uses Cosign to sign artifacts. All signatures are automatically uploaded to the public Rekor transparency log, which resides at https://rekor.sigstore.dev/.

:::note
The public key for Edgeless Systems' long-term code-signing key is:
```
-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEf8F1hpmwE+YCFXzjGtaQcrL6XZVT
JmEe5iSLvG1SyQSAew7WdMKF6o9t8e2TFuCkzlOhhlws2OHWbiFZnFWCFw==
-----END PUBLIC KEY-----
```
The public key is also available for download at https://edgeless.systems/es.pub and in the Twitter profile [@EdgelessSystems](https://twitter.com/EdgelessSystems).
:::

The Rekor transparency log is a public append-only ledger that verifies and records signatures and associated metadata. The Rekor transparency log enables everyone to observe the sequence of (software) signatures issued by Edgeless Systems and many other parties. The transparency log allows for the public identification of dubious or malicious signatures.

You should always ensure that (1) your CLI executable was signed with the private key corresponding to the above public key and that (2) there is a corresponding entry in the Rekor transparency log. Both can be done as is described in the following.

:::info
You don't need to verify the Constellation node images. This is done automatically by your CLI and the rest of Constellation.
:::

## Verify the signature

First, [install the Cosign CLI](https://docs.sigstore.dev/cosign/installation). Next, verify the signature that accompanies your CLI executable, for example:

```shell-session
$ cosign verify-blob --key https://edgeless.systems/es.pub --signature constellation-linux-amd64.sig constellation-linux-amd64

Verified OK
```

The above performs an offline verification of the provided public key, signature, and executable. To also verify that a corresponding entry exists in the public Rekor transparency log, add the variable `COSIGN_EXPERIMENTAL=1`:

```shell-session
$ COSIGN_EXPERIMENTAL=1 cosign verify-blob --key https://edgeless.systems/es.pub --signature constellation-linux-amd64.sig constellation-linux-amd64

tlog entry verified with uuid: 0629f03c379219f4ae1b99819fd4c266a39490a338ec24321198ba6ccc16f147 index: 3334047
Verified OK
```

🏁 You now know that your CLI executable was officially released and signed by Edgeless Systems.

## Optional: Manually inspect the transparency log

To further inspect the public Rekor transparency log, [install the Rekor CLI](https://docs.sigstore.dev/rekor/installation). A search for the CLI executable  should give a single UUID. (Note that this UUID contains the UUID from the previous `cosign` command.)

```shell-session
$ rekor-cli search --artifact constellation-linux-amd64

Found matching entries (listed by UUID):
362f8ecba72f43260629f03c379219f4ae1b99819fd4c266a39490a338ec24321198ba6ccc16f147
```

With this UUID you can get the full entry from the transparency log:

```shell-session
$ rekor-cli get --uuid=362f8ecba72f43260629f03c379219f4ae1b99819fd4c266a39490a338ec24321198ba6ccc16f147

LogID: c0d23d6ad406973f9559f3ba2d1ca01f84147d8ffc5b8445c224f98b9591801d
Index: 3334047
IntegratedTime: 2022-08-31T08:36:25Z
UUID: 0629f03c379219f4ae1b99819fd4c266a39490a338ec24321198ba6ccc16f147
Body: {
  "HashedRekordObj": {
    "data": {
      "hash": {
        "algorithm": "sha256",
        "value": "7cdc7a7101b215058264279b8d8f624e4e48b6b42cd54857a5e02daf1a1b014c"
      }
    },
    "signature": {
      "content": "MEYCIQDdL8fuhtFk6ON4b6kW6bvLMXqvw37nm8/UiLcYKjogsAIhAODZCdS1HgHvFJ5KFxT1JZzRN2wPdn3HZsiP0+3q6zsL",
      "publicKey": {
        "content": "LS0tLS1CRUdJTiBQVUJMSUMgS0VZLS0tLS0KTUZrd0V3WUhLb1pJemowQ0FRWUlLb1pJemowREFRY0RRZ0FFZjhGMWhwbXdFK1lDRlh6akd0YVFjckw2WFpWVApKbUVlNWlTTHZHMVN5UVNBZXc3V2RNS0Y2bzl0OGUyVEZ1Q2t6bE9oaGx3czJPSFdiaUZabkZXQ0Z3PT0KLS0tLS1FTkQgUFVCTElDIEtFWS0tLS0tCg=="
      }
    }
  }
}
```
The field `publicKey` should contain Edgeless Systems' public key in Base64 encoding.

You can get an exhaustive list of artifact signatures issued by Edgeless Systems via the following command:

```bash
$ rekor-cli search --public-key https://edgeless.systems/es.pub --pki-format x509
```

Edgeless Systems monitors this list to detect potential unauthorized use of its private key.

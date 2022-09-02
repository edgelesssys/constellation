# constellation-kms-client

This library provides an interface for the key management services used with constellation.
It's intendet for the Constellation CSI Plugins and the CLI.

## KMS

The Cloud KMS is where we store our key encryption key (KEK).
It should be initiated by the CLI and provided with a key release policy.
The CSP Plugin can request to encrypt data encryption keys (DEK) with the DEK to safely store them on persistent memory.
The [kms](../kms/) package interacts with the Cloud KMS APIs.
Currently planned are KMS are:

* AWS KMS
* GCP CKM
* Azure Key Vault


## Storage

Storage is where the CSI Plugin stores the encrypted DEKs.
Currently planned are:

* AWS S3, SSP
* GCP GCS
* Azure Blob

# Credentials

Each Plugin requires credentials to authenticate itself to a CSP.

## Storage

This section covers how credentials are used by the storage plugins.

### AWS S3 Bucket

To use the AWS S3 Bucket plugin, you need to have an existing [AWS account](https://aws.amazon.com/de/premiumsupport/knowledge-center/create-and-activate-aws-account/). 

For authentication, you have to pass a config file to the plugin. The AWS config package lets you automatically fetch the data from the local AWS directory.

#### Passing credentials automatically
You need to store your credentials in your local AWS directory at `$HOME/.aws/`. The AWS config package uses the values from the directory to build a config file, which is used to authenticate the client. The local AWS directory must contain two files:
- `credentials`
```bash
[default]
aws_access_key_id = MyAccessKeyId
aws_secret_access_key = MySecretAccessKey
```
- `config`
```bash
[default]
region = MyRegion
output = json
```
If you have the [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html) installed, you can
initialise the files with the following command:
```bash
aws configure
```

To create the client:
```Go
cfg, err := config.LoadDefaultConfig(context.TODO())
store, err := storage.NewAWSS3Storage(context.TODO(), "bucketName", cfg, func(*s3.Options) {})
```

### Azure Blob Storage

To use the Azure Blob storage plugin, you need to first [create a storage account](https://docs.microsoft.com/en-us/azure/storage/common/storage-account-create?tabs=azure-portal) or give your application access to an existing storage account.

The plugin uses a connection string created for the storage account to authenticate itself to the Azure API.
The connection string can be found in your storage account in the Azure Portal under the "Access Keys" section or with the following Azure CLI command:
```bash
az storage account show-connection-string -g MyResourceGroup -n MyStorageAccount
```

The client will use the specified Blob container if it already exists, or create it first otherwise.

To create the client:
```Go
connectionString := "DefaultEndpointsProtocol=https;AccountName=<myAccountName>;AccountKey=<myAccountKey>;EndpointSuffix=core.windows.net"
store, err := storage.NewAzureStorage(context.TODO(), connectionString, "myContainer", nil)
```

### Google Cloud Storage

To use the Google Cloud Storage plugin, the  Cloud Storage API needs to be enabled in your Google Cloud Account. You can use an existing bucket, create a new bucket yourself, or let the plugin create the bucket on initialization.

When using the Google Cloud APIs, your application will typically [authenticate as a service account](https://cloud.google.com/docs/authentication/production).
You have two options for passing service account credentials to the Storage plugin: (1) Fetching them automatically from the environment or (2) passing them manually in your Go code.

Note that the serivce account requires the following permissions:
* `storage.buckets.create`
* `storage.buckets.get`
* `storage.objects.create`
* `storage.objects.get`
* `storage.objects.update`

#### Finding credentials automatically

If your application is running inside a Google Cloud environment, and you have [attached a service account](https://cloud.google.com/iam/docs/impersonating-service-accounts#attaching-to-resources) to that environment, the Storage Plugin can retrieve credentials for the service account automatically. 

If your application is running in an environment with no service account attached, you can manually attach a [service account key](https://cloud.google.com/iam/docs/service-accounts#service_account_keys) to that environment.
After you [created a service account and stored its access key to file](https://cloud.google.com/docs/authentication/production#create_service_account) you need to set the environment variable `GOOGLE_APPLICATION_CREDENTIALS` to the location of that file.
The Storage Plugin will then be able to automatically load the credentials from there:
```bash
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account-file.json"
```

To create the client:
```Go
store, err := storage.NewGoogleCloudStorage(context.TODO(), "myProject", "myBucket", nil)
```

#### Passing credentials manually

You may also explicitly use your service account file in code.
First, create a service account and key the same way as in [finding credentials automatically](#finding-credentials-automatically).
You can then specify the location of the file in your application code.

To create the client:
```Go
credentialFile := "/path/to/service-account-file.json"
opts := option.WithCredentialsFile(credentialFile)
store, err := storage.NewGoogleCloudStorage(context.TODO(), "myProject", "myBucket", nil, opts)
```

## Key Managment Service

This section covers how credentials are used by the KMS plugins.

### AWS KMS

To use the AWS KMS, you need to have an existing [AWS account](https://aws.amazon.com/de/premiumsupport/knowledge-center/create-and-activate-aws-account/).

For authentication, you have to pass a config file to the plugin. As with the AWS S3 Bucket, you can use the config package to retrieve the data for the config file from your local AWS directory. Follow the steps listed in the [AWS S3 Bucket](#aws-s3-bucket) section on how to build the config file automatically.

To create the client:
```Go
cfg, err := config.LoadDefaultConfig(context.TODO())
awsClient := kms.NewFromConfig(cfg)
client := kmsAws.New(awsClient, nil)
```

### Azure Key Vault

To use the Azure Key Vault, you need to first [create a new key vault](https://docs.microsoft.com/en-us/azure/key-vault/general/quick-create-portal) or use an existing vault.

The implementation uses `NewDefaultAzureCredential` to load credentials. If you application is running on Azure infrastructure, credentials will be loaded using [managed identities](https://docs.microsoft.com/en-us/azure/developer/go/azure-sdk-authentication?tabs=bash#managed-identity). Otherwise you can use [environment variables to configure the client](https://docs.microsoft.com/en-us/azure/developer/go/azure-sdk-authentication?tabs=bash#-option-1-define-environment-variables).

To create the client for Azure KMS:
```Go
client, err := kmsAzure.New(context.TODO(), "myVault", kmsAz.DefaultCloud, nil, nil)
```

To create the client for Azure manged HSM KMS:
```Go
client, err := kmsAzure.NewHSM(context.TODO(), "myHSM", nil, nil)
```

### Google KMS

To use the Google Key Management Service, you need to first enable the KMS API and [create a key ring](https://cloud.google.com/kms/docs/quickstart#key_rings_and_keys) to use with the plugin or use an existing key ring.

Providing credentials to your application for Google's Cloud Key Managment Service happens the same way as described in [Google Cloud Storage](#google-cloud-storage).

Note that the service account used for authentication requires the following permissions:
* `cloudkms.cryptoKeyVersions.create`
* `cloudkms.cryptoKeyVersions.update`
* `cloudkms.cryptoKeyVersions.useToDecrypt`
* `cloudkms.cryptoKeyVersions.useToEncrypt`
* `cloudkms.importJobs.create`
* `cloudkms.importJobs.get`
* `cloudkms.importJobs.useToImport`

Once your credentials are set up you can create your client:

* With automatic credential selection: 

    ```Go
    client, err := kmsGcp.New("myProject", "global", "myKeyRing", nil, kmspb.ProtectionLevel_SOFTWARE)
    ```

* With manual credential selection:

    ```Go
    credentialFile := "/path/to/service-account-file.json"
    opts := option.WithCredentialsFile(credentialFile)
    client, err := kmsGcp.New("myProject", "global", "myKeyRing", nil, kmspb.ProtectionLevel_SOFTWARE, opts)
    ```

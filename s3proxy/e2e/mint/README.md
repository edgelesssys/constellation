# Notes for Constellation

The e2e tests are shipped as a docker container that executes multiple payloads.
Each payload is a S3 client program that sends a list of requests to a configurable S3 server.

Checkout `build.sh`, `upload.sh` and `run.sh` to see how the image is built, uploaded and run.

To locally run e2e tests you will have to:
- deploy s3proxy using the instructions in `../deploy/README.md`.
- Run `build.sh && run.sh`

You don't need to upload a new image manually as the CI will automatically build and upload mint images.

The CI uses the `deploy.sh` script to run the test image as a Kubernetes job inside the cluster.
This is not done for local development so you don't have to upload the image for every change.

# Mint [![Slack](https://slack.minio.io/slack?type=svg)](https://slack.minio.io) [![Docker Pulls](https://img.shields.io/docker/pulls/minio/mint.svg?maxAge=604800)](https://hub.docker.com/r/minio/mint/)

Mint is a testing framework for Minio object server, available as a podman image. It runs correctness, benchmarking and stress tests. Following are the SDKs/tools used in correctness tests.

- awscli
- aws-sdk-go
- aws-sdk-java
- aws-sdk-php
- aws-sdk-ruby
- healthcheck
- mc
- minio-go
- minio-java
- minio-js
- minio-py
- s3cmd
- s3select
- versioning

## Running Mint

Mint is run by `podman run` command which requires Podman to be installed. For Podman installation follow the steps [here](https://podman.io/getting-started/installation#installing-on-linux).

To run Mint with Minio Play server as test target,

```sh
$ podman run -e SERVER_ENDPOINT=play.minio.io:9000 -e ACCESS_KEY=Q3AM3UQ867SPQQA43P2F \
             -e SECRET_KEY=zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG -e ENABLE_HTTPS=1 minio/mint
```

After the tests are run, output is stored in `/mint/log` directory inside the container. To get these logs, use `podman cp` command. For example
```sh
podman cp <container-id>:/mint/log /tmp/logs
```

### Mint environment variables

Below environment variables are required to be passed to the podman container. Supported environment variables:

| Environment variable   | Description                                                                                                                                    | Example                                    |
|:-----------------------|:-----------------------------------------------------------------------------------------------------------------------------------------------|:-------------------------------------------|
| `SERVER_ENDPOINT`      | Endpoint of Minio server in the format `HOST:PORT`; for virtual style `IP:PORT`                                                                | `play.minio.io:9000`                       |
| `ACCESS_KEY`           | Access key for `SERVER_ENDPOINT` credentials                                                                                                   | `Q3AM3UQ867SPQQA43P2F`                     |
| `SECRET_KEY`           | Secret Key for `SERVER_ENDPOINT` credentials                                                                                                   | `zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG` |
| `ENABLE_HTTPS`         | (Optional) Set `1` to indicate to use HTTPS to access `SERVER_ENDPOINT`. Defaults to `0` (HTTP)                                                | `1`                                        |
| `MINT_MODE`            | (Optional) Set mode indicating what category of tests to be run by values `core`, `full`. Defaults to `core`                                   | `full`                                     |
| `DOMAIN`               | (Optional) Value of MINIO_DOMAIN environment variable used in Minio server                                                                     | `myminio.com`                              |
| `ENABLE_VIRTUAL_STYLE` | (Optional) Set `1` to indicate virtual style access . Defaults to `0` (Path style)                                                             | `1`                                        |
| `RUN_ON_FAIL`          | (Optional) Set `1` to indicate execute all tests independent of failures (currently implemented for minio-go and minio-java) . Defaults to `0` | `1`                                        |
| `SERVER_REGION`        | (Optional) Set custom region for region specific tests                                                                                         | `us-west-1`                                |

### Test virtual style access against Minio server

To test Minio server virtual style access with Mint, follow these steps:

- Set a domain in your Minio server using environment variable MINIO_DOMAIN. For example `export MINIO_DOMAIN=myminio.com`.
- Start Minio server.
- Execute Mint against Minio server (with `MINIO_DOMAIN` set to `myminio.com`) using this command
```sh
$ podman run -e "SERVER_ENDPOINT=192.168.86.133:9000" -e "DOMAIN=minio.com"  \
	     -e "ACCESS_KEY=minio" -e "SECRET_KEY=minio123" -e "ENABLE_HTTPS=0" \
	     -e "ENABLE_VIRTUAL_STYLE=1" minio/mint
```

### Mint log format

All test logs are stored in `/mint/log/log.json` as multiple JSON document.  Below is the JSON format for every entry in the log file.

| JSON field | Type     | Description                                                   | Example                                               |
|:-----------|:---------|:--------------------------------------------------------------|:------------------------------------------------------|
| `name`     | _string_ | Testing tool/SDK name                                         | `"aws-sdk-php"`                                       |
| `function` | _string_ | Test function name                                            | `"getBucketLocation ( array $params = [] )"`          |
| `args`     | _object_ | (Optional) Key/Value map of arguments passed to test function | `{"Bucket":"aws-sdk-php-bucket-20341"}`               |
| `duration` | _int_    | Time taken in milliseconds to run the test                    | `384`                                                 |
| `status`   | _string_ | one of `PASS`, `FAIL` or `NA`                                 | `"PASS"`                                              |
| `alert`    | _string_ | (Optional) Alert message indicating test failure              | `"I/O error on create file"`                          |
| `message`  | _string_ | (Optional) Any log message                                    | `"validating checksum of downloaded object"`          |
| `error`    | _string_ | Detailed error message including stack trace on status `FAIL` | `"Error executing \"CompleteMultipartUpload\" on ...` |

## For Developers

### Running Mint development code

After making changes to Mint source code a local podman image can be built/run by

```sh
$ podman build -t minio/mint . -f Dockerfile
$ podman run -e SERVER_ENDPOINT=play.minio.io:9000 -e ACCESS_KEY=Q3AM3UQ867SPQQA43P2F \
             -e SECRET_KEY=zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG \
             -e ENABLE_HTTPS=1 -e MINT_MODE=full minio/mint:latest
```


### Adding tests with new tool/SDK

Below are the steps need to be followed

- Create new app directory under [build](https://github.com/minio/mint/tree/master/build) and [run/core](https://github.com/minio/mint/tree/master/run/core) directories.
- Create `install.sh` which does installation of required tool/SDK under app directory.
- Any build and install time dependencies should be added to [install-packages.list](https://github.com/minio/mint/blob/master/install-packages.list).
- Build time dependencies should be added to [remove-packages.list](https://github.com/minio/mint/blob/master/remove-packages.list) for removal to have clean Mint podman image.
- Add `run.sh` in app directory under `run/core` which execute actual tests.

#### Test data
Tests may use pre-created data set to perform various object operations on Minio server.  Below data files are available under `/mint/data` directory.

| File name        | Size    |
|:-----------------|:--------|
| datafile-0-b     | 0B      |
| datafile-1-b     | 1B      |
| datafile-1-kB    | 1KiB    |
| datafile-10-kB   | 10KiB   |
| datafile-33-kB   | 33KiB   |
| datafile-100-kB  | 100KiB  |
| datafile-1-MB    | 1MiB    |
| datafile-1.03-MB | 1.03MiB |
| datafile-5-MB    | 5MiB    |
| datafile-6-MB    | 6MiB    |
| datafile-10-MB   | 10MiB   |
| datafile-11-MB   | 11MiB   |
| datafile-65-MB   | 65MiB   |
| datafile-129-MB  | 129MiB  |

### Updating SDKs/binaries in the image

In many cases, updating the SDKs or binaries in the image is just a matter of making a commit updating the corresponding version in this repo. However, in some cases, e.g. when `mc` needs to be updated (the latest `mc` is pulled in during each mint image build), a sort of "dummy" commit is required as an image rebuild must be triggered. Note that an empty commit does not appear to trigger the image rebuild in the Docker Hub.

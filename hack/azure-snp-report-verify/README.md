# Obtain current Azure SNP ID key digest & firmware versions

On Azure, Constellation verifies that the SNP attestation report contains Azure's ID key digest.
Additionally, some firmware security version numbers (SVNs) are validated.
Currently, the only way to verify the digest's origin is to perform guest attestation with the help of the Microsoft Azure Attestation (MAA) service.
There's a [sample](https://github.com/Azure/confidential-computing-cvm-guest-attestation) on how to do this, but it's not straightforward.
So we created tooling to make things easier.

Perform the following steps to get the ID key digest & firmware versions:

1. Create an Ubuntu CVM on Azure with secure boot enabled and ssh into it.
2. Run
   ```
   docker run --rm --privileged -v/sys/kernel/security:/sys/kernel/security ghcr.io/edgelesssys/constellation/azure-snp-reporter
   ```
   This executes the guest attestation and prints the JWT received from the MAA. (It's the long base64 blob.)
3. Copy the JWT and run **on your local trusted machine**:
   ```
   go run verify.go <jwt>
   ```
   On success it prints the ID key digest and relevant firmware SVNs.

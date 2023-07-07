/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
# Virtual Trusted Platform Module (vTPM)

This package provides functions to interact with a vTPM.
It also implements the low level TPM attestation and verification logic of Constellation's TPM attestation workflow.

Code that directly interacts with the TPM goes here.

# vTPM components

For attestation we make use of multiple vTPM features:

  - Endorsement Key

    Asymmetric key used to establish trust in other keys issued by the TPM or used directly for attestation. The private part never leaves the TPM, while the public part, referred to as Endorsement Public Key (EPK), is available to remote parties. The TPM can issue new keys, signed by its endorsement key, which can then be verified by a remote party using the EPK.

  - Endorsement Public Key Certificate (EPKC)

    A Certificate signed by the TPM manufacturer verifying the authenticity of the EPK. The public key of the Certificate is the EPK.

  - Event Log

    A log of events over the boot process.

  - [Platform Control Register (PCR)]

    Registers holding measurements of software and configuration data. PCR values are not directly written, but updated: a new value is the digest of the old value concatenated with the to be added data.
    Contents of the PCRs can be signed for attestation. Providing proof to a remote party about software running on the system.

# Attestation flow

1. The VM boots and writes its measured software state to the PCRs.

2. The PCRs are hashed and signed by the EPK.

3. An attestation statement is created, containing the EPK, the original PCR values, the hashed PCRs, the signature, and the event log.

4. A remote party establishes trust in the TPMs EPK by verifying its EPKC with the TPM manufactures CA certificate chain.

5. The remote party verifies the signature was created by the TPM, and the hash matches the PCRs.

6. The remote party reads the event log and verifies measuring the event log results in the given PCR values

7. The software state is now verified, the only thing left to do is to decide if the state is good or not. This is done by comparing the given PCR values to a set of expected PCR values.

[Platform Control Register (PCR)]: https://link.springer.com/chapter/10.1007/978-1-4302-6584-9_12
*/
package vtpm

import (
	"io"

	"github.com/google/go-tpm/legacy/tpm2"
)

// TPMOpenFunc opens a TPM device.
type TPMOpenFunc func() (io.ReadWriteCloser, error)

// OpenVTPM opens the vTPM at `TPMPath`.
func OpenVTPM() (io.ReadWriteCloser, error) {
	return tpm2.OpenTPM()
}

type nopTPM struct{}

// OpenNOPTPM returns a NOP io.ReadWriteCloser that can be used as a TPM.
func OpenNOPTPM() (io.ReadWriteCloser, error) {
	return &nopTPM{}, nil
}

func (t nopTPM) Read(p []byte) (int, error) {
	return len(p), nil
}

func (t nopTPM) Write(p []byte) (int, error) {
	return len(p), nil
}

func (t nopTPM) Close() error {
	return nil
}

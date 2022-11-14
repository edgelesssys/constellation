/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package vtpm

import (
	"errors"

	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/google/go-tpm/tpm2"
)

// MarkNodeAsBootstrapped marks a node as initialized by extending PCRs.
func MarkNodeAsBootstrapped(openTPM TPMOpenFunc, clusterID []byte) error {
	tpm, err := openTPM()
	if err != nil {
		return err
	}
	defer tpm.Close()

	// clusterID is used to uniquely identify this running instance of Constellation
	return tpm2.PCREvent(tpm, measurements.PCRIndexClusterID, clusterID)
}

// IsNodeBootstrapped checks if a node is already bootstrapped by reading PCRs.
func IsNodeBootstrapped(openTPM TPMOpenFunc) (bool, error) {
	tpm, err := openTPM()
	if err != nil {
		return false, err
	}
	defer tpm.Close()

	idxClusterID := int(measurements.PCRIndexClusterID)
	pcrs, err := tpm2.ReadPCRs(tpm, tpm2.PCRSelection{
		Hash: tpm2.AlgSHA256,
		PCRs: []int{idxClusterID},
	})
	if err != nil {
		return false, err
	}
	if len(pcrs[idxClusterID]) == 0 {
		return false, errors.New("cluster ID PCR does not exist")
	}
	return pcrInitialized(pcrs[idxClusterID]), nil

	/* Old code that will be reenabled in the future
	idxOwner := int(PCRIndexOwnerID)
	idxCluster := int(PCRIndexClusterID)
	selection := tpm2.PCRSelection{
		Hash: tpm2.AlgSHA256,
		PCRs: []int{idxOwner, idxCluster},
	}

	pcrs, err := tpm2.ReadPCRs(tpm, selection)
	if err != nil {
		return false, err
	}

	if len(pcrs[idxOwner]) == 0 {
		return false, errors.New("owner ID PCR does not exist")
	}
	if len(pcrs[idxCluster]) == 0 {
		return false, errors.New("cluster ID PCR does not exist")
	}

	ownerInitialized := pcrInitialized(pcrs[idxOwner])
	clusterInitialized := pcrInitialized(pcrs[idxCluster])

	if ownerInitialized == clusterInitialized {
		return ownerInitialized && clusterInitialized, nil
	}
	ownerState := "not initialized"
	if ownerInitialized {
		ownerState = "initialized"
	}
	clusterState := "not initialized"
	if clusterInitialized {
		clusterState = "initialized"
	}
	return false, fmt.Errorf("PCRs %v and %v are not consistent: PCR[%v]=%v (%v), PCR[%v]=%v (%v)", idxOwner, idxCluster, idxOwner, pcrs[idxOwner], ownerState, idxCluster, pcrs[idxCluster], clusterState)
	*/
}

// pcrInitialized checks if a PCR value is set to a non-zero value.
func pcrInitialized(pcr []byte) bool {
	for _, b := range pcr {
		if b != 0 {
			return true
		}
	}
	return false
}

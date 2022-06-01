package vtpm

import (
	"errors"
	"fmt"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpmutil"
)

const (
	// PCRIndexOwnerID is a PCR we extend to mark the node as initialized.
	// The value used to extend is derived from Constellation's master key.
	PCRIndexOwnerID = tpmutil.Handle(11)
	// PCRIndexClusterID is a PCR we extend to mark the node as initialized.
	// The value used to extend is a random generated 32 Byte value.
	PCRIndexClusterID = tpmutil.Handle(12)
)

// MarkNodeAsInitialized marks a node as initialized by extending PCRs.
func MarkNodeAsInitialized(openTPM TPMOpenFunc, ownerID, clusterID []byte) error {
	tpm, err := openTPM()
	if err != nil {
		return err
	}
	defer tpm.Close()

	// ownerID is used to identify the Constellation as belonging to a specific master key
	if err := tpm2.PCREvent(tpm, PCRIndexOwnerID, ownerID); err != nil {
		return err
	}
	// clusterID is used to uniquely identify this running instance of Constellation
	return tpm2.PCREvent(tpm, PCRIndexClusterID, clusterID)
}

// IsNodeInitialized checks if a node is already initialized by reading PCRs.
func IsNodeInitialized(openTPM TPMOpenFunc) (bool, error) {
	tpm, err := openTPM()
	if err != nil {
		return false, err
	}
	defer tpm.Close()

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

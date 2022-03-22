package vtpm

import (
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

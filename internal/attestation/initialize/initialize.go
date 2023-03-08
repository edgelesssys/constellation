/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package initialize implements functions to mark a node as initialized in the context of cluster attestation.
// This is done by measuring the cluster ID using the available CC technology.
package initialize

import (
	"bytes"
	"errors"
	"io"

	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/tdx"
	tdxapi "github.com/edgelesssys/go-tdx-qpl/tdx"
	"github.com/google/go-tpm/tpm2"
)

// MarkNodeAsBootstrapped marks a node as initialized by extending PCRs.
// clusterID is used to uniquely identify this running instance of Constellation.
func MarkNodeAsBootstrapped(openDevice func() (io.ReadWriteCloser, error), clusterID []byte) error {
	device, err := openDevice()
	if err != nil {
		return err
	}
	defer device.Close()

	// The TDX device is of type *os.File, while the TPM device may be
	// *os.File or an emulated device over a unix socket.
	// Therefore, we can't simply use a type switch here,
	// since the TPM may implement the same methods as the TDX device
	if handle, ok := tdx.IsTDXDevice(device); ok {
		return tdxMarkNodeAsBootstrapped(handle, clusterID)
	}
	return tpmMarkNodeAsBootstrapped(device, clusterID)
}

// IsNodeBootstrapped checks if a node is already bootstrapped by reading PCRs.
func IsNodeBootstrapped(openDevice func() (io.ReadWriteCloser, error)) (bool, error) {
	device, err := openDevice()
	if err != nil {
		return false, err
	}
	defer device.Close()

	// The TDX device is of type *os.File, while the TPM device may be
	// *os.File or an emulated device over a unix socket.
	// Therefore, we can't simply use a type switch here,
	// since the TPM may implement the same methods as the TDX device
	if handle, ok := tdx.IsTDXDevice(device); ok {
		return tdxIsNodeBootstrapped(handle)
	}
	return tpmIsNodeBootstrapped(device)
}

func tdxIsNodeBootstrapped(handle tdx.Device) (bool, error) {
	tdMeasure, err := tdxapi.ReadMeasurements(handle)
	if err != nil {
		return false, err
	}
	return measurementInitialized(tdMeasure[measurements.TDXIndexClusterID][:]), nil
}

func tpmIsNodeBootstrapped(tpm io.ReadWriteCloser) (bool, error) {
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
	return measurementInitialized(pcrs[idxClusterID]), nil

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

func tdxMarkNodeAsBootstrapped(handle tdx.Device, clusterID []byte) error {
	return tdxapi.ExtendRTMR(handle, clusterID, measurements.RTMRIndexClusterID)
}

func tpmMarkNodeAsBootstrapped(tpm io.ReadWriteCloser, clusterID []byte) error {
	return tpm2.PCREvent(tpm, measurements.PCRIndexClusterID, clusterID)
}

// measurementInitialized checks if a PCR value is set to a non-zero value.
func measurementInitialized(measurement []byte) bool {
	return !bytes.Equal(measurement, bytes.Repeat([]byte{0x00}, len(measurement)))
}

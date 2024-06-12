/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package provider

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/attestation/idkeydigest"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/encoding"
)

// naming schema:
// convertFromTf<type> : convert a terraform struct to a constellation struct
// convertToTf<type> : convert a constellation struct to a terraform struct
// terraform struct: used to parse the terraform state
// constellation struct: used to call the constellation API

// convertFromTfAttestationCfg converts the related terraform struct to a constellation attestation config.
func convertFromTfAttestationCfg(tfAttestation attestationAttribute, attestationVariant variant.Variant) (config.AttestationCfg, error) {
	c11nMeasurements := make(measurements.M)
	for strIdx, v := range tfAttestation.Measurements {
		idx, err := strconv.ParseUint(strIdx, 10, 32)
		if err != nil {
			return nil, err
		}
		expectedBt, err := hex.DecodeString(v.Expected)
		if err != nil {
			return nil, err
		}
		var valOption measurements.MeasurementValidationOption
		switch v.WarnOnly {
		case true:
			valOption = measurements.WarnOnly
		case false:
			valOption = measurements.Enforce
		}
		c11nMeasurements[uint32(idx)] = measurements.Measurement{
			Expected:      expectedBt,
			ValidationOpt: valOption,
		}
	}

	var attestationConfig config.AttestationCfg
	switch attestationVariant {
	case variant.AzureSEVSNP{}:
		firmwareCfg, err := convertFromTfFirmwareCfg(tfAttestation.AzureSNPFirmwareSignerConfig)
		if err != nil {
			return nil, fmt.Errorf("converting firmware signer config: %w", err)
		}

		var rootKey config.Certificate
		if err := json.Unmarshal([]byte(tfAttestation.AMDRootKey), &rootKey); err != nil {
			return nil, fmt.Errorf("unmarshalling root key: %w", err)
		}

		attestationConfig = &config.AzureSEVSNP{
			Measurements:         c11nMeasurements,
			BootloaderVersion:    newVersion(tfAttestation.BootloaderVersion),
			TEEVersion:           newVersion(tfAttestation.TEEVersion),
			SNPVersion:           newVersion(tfAttestation.SNPVersion),
			MicrocodeVersion:     newVersion(tfAttestation.MicrocodeVersion),
			FirmwareSignerConfig: firmwareCfg,
			AMDRootKey:           rootKey,
		}
	case variant.AWSSEVSNP{}:
		var rootKey config.Certificate
		if err := json.Unmarshal([]byte(tfAttestation.AMDRootKey), &rootKey); err != nil {
			return nil, fmt.Errorf("unmarshalling root key: %w", err)
		}

		attestationConfig = &config.AWSSEVSNP{
			Measurements:      c11nMeasurements,
			BootloaderVersion: newVersion(tfAttestation.BootloaderVersion),
			TEEVersion:        newVersion(tfAttestation.TEEVersion),
			SNPVersion:        newVersion(tfAttestation.SNPVersion),
			MicrocodeVersion:  newVersion(tfAttestation.MicrocodeVersion),
			AMDRootKey:        rootKey,
		}
	case variant.AzureTDX{}:
		var rootKey config.Certificate
		if err := json.Unmarshal([]byte(tfAttestation.TDX.IntelRootKey), &rootKey); err != nil {
			return nil, fmt.Errorf("unmarshalling root key: %w", err)
		}
		teeTCBSVN, err := hex.DecodeString(tfAttestation.TDX.TEETCBSVN)
		if err != nil {
			return nil, fmt.Errorf("decoding tee_tcb_svn: %w", err)
		}
		qeVendorID, err := hex.DecodeString(tfAttestation.TDX.QEVendorID)
		if err != nil {
			return nil, fmt.Errorf("decoding qe_vendor_id: %w", err)
		}
		mrSeam, err := hex.DecodeString(tfAttestation.TDX.MRSeam)
		if err != nil {
			return nil, fmt.Errorf("decoding mr_seam: %w", err)
		}
		xfam, err := hex.DecodeString(tfAttestation.TDX.XFAM)
		if err != nil {
			return nil, fmt.Errorf("decoding xfam: %w", err)
		}

		attestationConfig = &config.AzureTDX{
			Measurements: c11nMeasurements,
			QESVN:        newVersion(tfAttestation.TDX.QESVN),
			PCESVN:       newVersion(tfAttestation.TDX.PCESVN),
			TEETCBSVN:    newVersion(encoding.HexBytes(teeTCBSVN)),
			QEVendorID:   newVersion(encoding.HexBytes(qeVendorID)),
			MRSeam:       mrSeam,
			XFAM:         newVersion(encoding.HexBytes(xfam)),
			IntelRootKey: rootKey,
		}
	case variant.GCPSEVES{}:
		attestationConfig = &config.GCPSEVES{
			Measurements: c11nMeasurements,
		}
	case variant.GCPSEVSNP{}:
		attestationConfig = &config.GCPSEVSNP{
			Measurements: c11nMeasurements,
		}
	case variant.QEMUVTPM{}:
		attestationConfig = &config.QEMUVTPM{
			Measurements: c11nMeasurements,
		}
	default:
		return nil, fmt.Errorf("unknown attestation variant: %s", attestationVariant)
	}
	return attestationConfig, nil
}

// convertToTfAttestationCfg converts the constellation attestation config to the related terraform structs.
func convertToTfAttestation(attVar variant.Variant, latestVersions attestationconfigapi.Entry) (tfAttestation attestationAttribute, err error) {
	tfAttestation = attestationAttribute{
		Variant: attVar.String(),
	}

	switch attVar {
	case variant.AWSSEVSNP{}:
		certStr, err := certAsString(config.DefaultForAWSSEVSNP().AMDRootKey)
		if err != nil {
			return tfAttestation, err
		}
		tfAttestation.AMDRootKey = certStr
		tfAttestation.BootloaderVersion = latestVersions.Bootloader
		tfAttestation.TEEVersion = latestVersions.TEE
		tfAttestation.SNPVersion = latestVersions.SNP
		tfAttestation.MicrocodeVersion = latestVersions.Microcode

	case variant.GCPSEVSNP{}:
		certStr, err := certAsString(config.DefaultForGCPSEVSNP().AMDRootKey)
		if err != nil {
			return tfAttestation, err
		}
		tfAttestation.AMDRootKey = certStr
		tfAttestation.BootloaderVersion = latestVersions.Bootloader
		tfAttestation.TEEVersion = latestVersions.TEE
		tfAttestation.SNPVersion = latestVersions.SNP
		tfAttestation.MicrocodeVersion = latestVersions.Microcode

	case variant.AzureSEVSNP{}:
		certStr, err := certAsString(config.DefaultForAzureSEVSNP().AMDRootKey)
		if err != nil {
			return tfAttestation, err
		}
		tfAttestation.AMDRootKey = certStr
		tfAttestation.BootloaderVersion = latestVersions.Bootloader
		tfAttestation.TEEVersion = latestVersions.TEE
		tfAttestation.SNPVersion = latestVersions.SNP
		tfAttestation.MicrocodeVersion = latestVersions.Microcode

		firmwareCfg := config.DefaultForAzureSEVSNP().FirmwareSignerConfig
		tfFirmwareCfg, err := convertToTfFirmwareCfg(firmwareCfg)
		if err != nil {
			return tfAttestation, err
		}
		tfAttestation.AzureSNPFirmwareSignerConfig = tfFirmwareCfg

	case variant.AzureTDX{}:
		certStr, err := certAsString(config.DefaultForAzureTDX().IntelRootKey)
		if err != nil {
			return tfAttestation, err
		}
		tfAttestation.TDX.IntelRootKey = certStr
		tfAttestation.TDX.PCESVN = latestVersions.PCESVN
		tfAttestation.TDX.QESVN = latestVersions.QESVN
		tfAttestation.TDX.TEETCBSVN = hex.EncodeToString(latestVersions.TEETCBSVN[:])
		tfAttestation.TDX.QEVendorID = hex.EncodeToString(latestVersions.QEVendorID[:])
		tfAttestation.TDX.XFAM = hex.EncodeToString(latestVersions.XFAM[:])

	case variant.GCPSEVES{}, variant.QEMUVTPM{}:
		// no additional fields
	default:
		return tfAttestation, fmt.Errorf("unknown attestation variant: %s", attVar)
	}
	return tfAttestation, nil
}

func certAsString(cert config.Certificate) (string, error) {
	certBytes, err := cert.MarshalJSON()
	if err != nil {
		return "", err
	}
	return string(certBytes), nil
}

// convertToTfFirmwareCfg converts the constellation firmware config to the terraform struct.
func convertToTfFirmwareCfg(firmwareCfg config.SNPFirmwareSignerConfig) (azureSnpFirmwareSignerConfigAttribute, error) {
	keyDigestAny, err := firmwareCfg.AcceptedKeyDigests.MarshalYAML()
	if err != nil {
		return azureSnpFirmwareSignerConfigAttribute{}, err
	}
	keyDigest, ok := keyDigestAny.([]string)
	if !ok {
		return azureSnpFirmwareSignerConfigAttribute{}, fmt.Errorf("reading Accepted Key Digests: could not convert %T to []string", keyDigestAny)
	}
	return azureSnpFirmwareSignerConfigAttribute{
		AcceptedKeyDigests: keyDigest,
		EnforcementPolicy:  firmwareCfg.EnforcementPolicy.String(),
		MAAURL:             firmwareCfg.MAAURL,
	}, nil
}

// convertFromTfFirmwareCfg converts the terraform struct to a constellation firmware config.
func convertFromTfFirmwareCfg(tfFirmwareCfg azureSnpFirmwareSignerConfigAttribute) (config.SNPFirmwareSignerConfig, error) {
	keyDigests, err := idkeydigest.UnmarshalHexString(tfFirmwareCfg.AcceptedKeyDigests)
	if err != nil {
		return config.SNPFirmwareSignerConfig{}, err
	}
	return config.SNPFirmwareSignerConfig{
		AcceptedKeyDigests: keyDigests,
		EnforcementPolicy:  idkeydigest.EnforcePolicyFromString(tfFirmwareCfg.EnforcementPolicy),
		MAAURL:             tfFirmwareCfg.MAAURL,
	}, nil
}

// convertToTfMeasurements converts the constellation measurements to the terraform struct.
func convertToTfMeasurements(m measurements.M) map[string]measurementAttribute {
	tfMeasurements := map[string]measurementAttribute{}
	for key, value := range m {
		keyStr := strconv.FormatUint(uint64(key), 10)
		tfMeasurements[keyStr] = measurementAttribute{
			Expected: hex.EncodeToString(value.Expected),
			WarnOnly: bool(value.ValidationOpt),
		}
	}
	return tfMeasurements
}

func newVersion[T uint8 | uint16 | encoding.HexBytes](v T) config.AttestationVersion[T] {
	return config.AttestationVersion[T]{
		Value: v,
	}
}

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
)

// naming schema:
// convertFromTf<type> : convert a terraform struct to a constellation struct
// convertToTf<type> : convert a constellation struct to a terraform struct
// terraform struct: used to parse the terraform state
// constellation struct: used to call the constellation API

// convertFromTfAttestationCfg converts the related terraform struct to a constellation attestation config.
func convertFromTfAttestationCfg(tfAttestation attestation, attestationVariant variant.Variant) (config.AttestationCfg, error) {
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

	var rootKey config.Certificate
	if err := json.Unmarshal([]byte(tfAttestation.AMDRootKey), &rootKey); err != nil {
		return nil, fmt.Errorf("unmarshalling root key: %w", err)
	}

	var attestationConfig config.AttestationCfg
	switch attestationVariant {
	case variant.AzureSEVSNP{}:
		firmwareCfg, err := convertFromTfFirmwareCfg(tfAttestation.AzureSNPFirmwareSignerConfig)
		if err != nil {
			return nil, fmt.Errorf("converting firmware signer config: %w", err)
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
		attestationConfig = &config.AWSSEVSNP{
			Measurements:      c11nMeasurements,
			BootloaderVersion: newVersion(tfAttestation.BootloaderVersion),
			TEEVersion:        newVersion(tfAttestation.TEEVersion),
			SNPVersion:        newVersion(tfAttestation.SNPVersion),
			MicrocodeVersion:  newVersion(tfAttestation.MicrocodeVersion),
			AMDRootKey:        rootKey,
		}
	case variant.GCPSEVES{}:
		attestationConfig = &config.GCPSEVES{
			Measurements: c11nMeasurements,
		}
	default:
		return nil, fmt.Errorf("unknown attestation variant: %s", attestationVariant)
	}
	return attestationConfig, nil
}

// convertToTfAttestationCfg converts the constellation attestation config to the related terraform structs.
func convertToTfAttestation(attVar variant.Variant, snpVersions attestationconfigapi.SEVSNPVersionAPI) (tfAttestation attestation, err error) {
	tfAttestation = attestation{
		Variant:           attVar.String(),
		BootloaderVersion: snpVersions.Bootloader,
		TEEVersion:        snpVersions.TEE,
		SNPVersion:        snpVersions.SNP,
		MicrocodeVersion:  snpVersions.Microcode,
	}

	switch attVar {
	case variant.AWSSEVSNP{}:
		certStr, err := certAsString(config.DefaultForAWSSEVSNP().AMDRootKey)
		if err != nil {
			return tfAttestation, err
		}
		tfAttestation.AMDRootKey = certStr

	case variant.AzureSEVSNP{}:
		certStr, err := certAsString(config.DefaultForAzureSEVSNP().AMDRootKey)
		if err != nil {
			return tfAttestation, err
		}
		tfAttestation.AMDRootKey = certStr

		firmwareCfg := config.DefaultForAzureSEVSNP().FirmwareSignerConfig
		tfFirmwareCfg, err := convertToTfFirmwareCfg(firmwareCfg)
		if err != nil {
			return tfAttestation, err
		}
		tfAttestation.AzureSNPFirmwareSignerConfig = tfFirmwareCfg
	case variant.GCPSEVES{}:
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
func convertToTfFirmwareCfg(firmwareCfg config.SNPFirmwareSignerConfig) (azureSnpFirmwareSignerConfig, error) {
	keyDigestAny, err := firmwareCfg.AcceptedKeyDigests.MarshalYAML()
	if err != nil {
		return azureSnpFirmwareSignerConfig{}, err
	}
	keyDigest, ok := keyDigestAny.([]string)
	if !ok {
		return azureSnpFirmwareSignerConfig{}, fmt.Errorf("reading Accepted Key Digests: could not convert %T to []string", keyDigestAny)
	}
	return azureSnpFirmwareSignerConfig{
		AcceptedKeyDigests: keyDigest,
		EnforcementPolicy:  firmwareCfg.EnforcementPolicy.String(),
		MAAURL:             firmwareCfg.MAAURL,
	}, nil
}

// convertFromTfFirmwareCfg converts the terraform struct to a constellation firmware config.
func convertFromTfFirmwareCfg(tfFirmwareCfg azureSnpFirmwareSignerConfig) (config.SNPFirmwareSignerConfig, error) {
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
func convertToTfMeasurements(m measurements.M) map[string]measurement {
	tfMeasurements := map[string]measurement{}
	for key, value := range m {
		keyStr := strconv.FormatUint(uint64(key), 10)
		tfMeasurements[keyStr] = measurement{
			Expected: hex.EncodeToString(value.Expected),
			WarnOnly: bool(value.ValidationOpt),
		}
	}
	return tfMeasurements
}

type extraMicroservices struct {
	CSIDriver bool `tfsdk:"csi_driver"`
}

type measurement struct {
	Expected string `tfsdk:"expected"`
	WarnOnly bool   `tfsdk:"warn_only"`
}

type attestation struct {
	BootloaderVersion            uint8                        `tfsdk:"bootloader_version"`
	TEEVersion                   uint8                        `tfsdk:"tee_version"`
	SNPVersion                   uint8                        `tfsdk:"snp_version"`
	MicrocodeVersion             uint8                        `tfsdk:"microcode_version"`
	AMDRootKey                   string                       `tfsdk:"amd_root_key"`
	AzureSNPFirmwareSignerConfig azureSnpFirmwareSignerConfig `tfsdk:"azure_firmware_signer_config"`
	Variant                      string                       `tfsdk:"variant"`
	Measurements                 map[string]measurement       `tfsdk:"measurements"`
}

type azureSnpFirmwareSignerConfig struct {
	AcceptedKeyDigests []string `tfsdk:"accepted_key_digests"`
	EnforcementPolicy  string   `tfsdk:"enforcement_policy"`
	MAAURL             string   `tfsdk:"maa_url"`
}

func newVersion(v uint8) config.AttestationVersion {
	return config.AttestationVersion{
		Value: v,
	}
}

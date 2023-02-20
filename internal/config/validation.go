/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/compatibility"
	"github.com/edgelesssys/constellation/v2/internal/config/instancetypes"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/edgelesssys/constellation/v2/internal/versionsapi"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"go.uber.org/multierr"
	"golang.org/x/mod/semver"
)

// DisplayValidationErrors shows all validation errors inside configError as one formatted string.
func DisplayValidationErrors(errWriter io.Writer, configError error) error {
	errs := multierr.Errors(configError)
	if errs != nil {
		fmt.Fprintln(errWriter, "Problems validating config file:")
		for _, err := range errs {
			fmt.Fprintln(errWriter, "\t"+err.Error())
		}
		fmt.Fprintln(errWriter, "Fix the invalid entries or generate a new configuration using `constellation config generate`")
		return errors.New("invalid configuration")
	}
	return nil
}

func registerInvalidK8sVersionError(ut ut.Translator) error {
	return ut.Add("supported_k8s_version", "{0} specifies an unsupported Kubernetes version. {1}", true)
}

func translateInvalidK8sVersionError(ut ut.Translator, fe validator.FieldError) string {
	validVersions := make([]string, len(versions.VersionConfigs))
	i := 0
	for k := range versions.VersionConfigs {
		validVersions[i] = string(k)
		i++
	}
	validVersionsSorted := semver.ByVersion(validVersions)
	sort.Sort(validVersionsSorted)

	if len(validVersionsSorted) == 0 {
		t, _ := ut.T("supported_k8s_version", fe.Field(), "No valid versions available. This should never happen")
		return t
	}
	maxVersion := validVersionsSorted[len(validVersionsSorted)-1]
	minVersion := validVersionsSorted[0]

	var errorMsg string
	configured, ok := fe.Value().(string)
	if !ok {
		errorMsg = fmt.Sprintf("The configured version is not a valid string. Supported versions: %s", strings.Join(validVersionsSorted, " "))
		t, _ := ut.T("supported_k8s_version", fe.Field(), errorMsg)
		return t
	}

	configured = compatibility.EnsurePrefixV(configured)
	switch {
	case !semver.IsValid(configured):
		errorMsg = "The configured version is not a valid semantic version"
	case semver.Compare(configured, minVersion) == -1:
		errorMsg = fmt.Sprintf("The configured version %s is older than the oldest version supported by this CLI: %s", configured, minVersion)
	case semver.Compare(configured, maxVersion) == 1:
		errorMsg = fmt.Sprintf("The configured version %s is newer than the newest version supported by this CLI: %s", configured, maxVersion)
	}

	errorMsg = errorMsg + fmt.Sprintf("Supported versions: %s", strings.Join(validVersionsSorted, " "))

	t, _ := ut.T("supported_k8s_version", fe.Field(), errorMsg)

	return t
}

func validateAWSInstanceType(fl validator.FieldLevel) bool {
	return validInstanceTypeForProvider(fl.Field().String(), false, cloudprovider.AWS)
}

func validateAzureInstanceType(fl validator.FieldLevel) bool {
	azureConfig := fl.Parent().Interface().(AzureConfig)
	var acceptNonCVM bool
	if azureConfig.ConfidentialVM != nil {
		// This is the inverse of the config value (acceptNonCVMs is true if confidentialVM is false).
		// We could make the validator the other way around, but this should be an explicit bypass rather than checking if CVMs are "allowed".
		acceptNonCVM = !*azureConfig.ConfidentialVM
	}
	return validInstanceTypeForProvider(fl.Field().String(), acceptNonCVM, cloudprovider.Azure)
}

func validateGCPInstanceType(fl validator.FieldLevel) bool {
	return validInstanceTypeForProvider(fl.Field().String(), false, cloudprovider.GCP)
}

// validateProvider checks if zero or more than one providers are defined in the config.
func validateProvider(sl validator.StructLevel) {
	provider := sl.Current().Interface().(ProviderConfig)
	providerCount := 0

	if provider.AWS != nil {
		providerCount++
	}
	if provider.Azure != nil {
		providerCount++
	}
	if provider.GCP != nil {
		providerCount++
	}
	if provider.QEMU != nil {
		providerCount++
	}

	if providerCount < 1 {
		sl.ReportError(provider, "Provider", "Provider", "no_provider", "")
	} else if providerCount > 1 {
		sl.ReportError(provider, "Provider", "Provider", "more_than_one_provider", "")
	}
}

func registerTranslateAWSInstanceTypeError(ut ut.Translator) error {
	return ut.Add("aws_instance_type", fmt.Sprintf("{0} must be an instance from one of the following families types with size xlarge or higher: %v", instancetypes.AWSSupportedInstanceFamilies), true)
}

func translateAWSInstanceTypeError(ut ut.Translator, fe validator.FieldError) string {
	t, _ := ut.T("aws_instance_type", fe.Field())

	return t
}

func registerTranslateGCPInstanceTypeError(ut ut.Translator) error {
	return ut.Add("gcp_instance_type", fmt.Sprintf("{0} must be one of %v", instancetypes.GCPInstanceTypes), true)
}

func translateGCPInstanceTypeError(ut ut.Translator, fe validator.FieldError) string {
	t, _ := ut.T("gcp_instance_type", fe.Field())

	return t
}

// Validation translation functions for Provider errors.
func registerNoProviderError(ut ut.Translator) error {
	return ut.Add("no_provider", "{0}: No provider has been defined (requires either Azure, GCP or QEMU)", true)
}

func translateNoProviderError(ut ut.Translator, fe validator.FieldError) string {
	t, _ := ut.T("no_provider", fe.Field())

	return t
}

func registerMoreThanOneProviderError(ut ut.Translator) error {
	return ut.Add("more_than_one_provider", "{0}: Only one provider can be defined ({1} are defined)", true)
}

func (c *Config) translateMoreThanOneProviderError(ut ut.Translator, fe validator.FieldError) string {
	definedProviders := make([]string, 0)

	// c.Provider should not be nil as Provider would need to be defined for the validation to fail in this place.
	if c.Provider.AWS != nil {
		definedProviders = append(definedProviders, "AWS")
	}
	if c.Provider.Azure != nil {
		definedProviders = append(definedProviders, "Azure")
	}
	if c.Provider.GCP != nil {
		definedProviders = append(definedProviders, "GCP")
	}
	if c.Provider.QEMU != nil {
		definedProviders = append(definedProviders, "QEMU")
	}

	// Show single string if only one other provider is defined, show list with brackets if multiple are defined.
	t, _ := ut.T("more_than_one_provider", fe.Field(), strings.Join(definedProviders, ", "))

	return t
}

func validInstanceTypeForProvider(insType string, acceptNonCVM bool, provider cloudprovider.Provider) bool {
	switch provider {
	case cloudprovider.AWS:
		return checkIfAWSInstanceTypeIsValid(insType)
	case cloudprovider.Azure:
		if acceptNonCVM {
			for _, instanceType := range instancetypes.AzureTrustedLaunchInstanceTypes {
				if insType == instanceType {
					return true
				}
			}
		} else {
			for _, instanceType := range instancetypes.AzureCVMInstanceTypes {
				if insType == instanceType {
					return true
				}
			}
		}
		return false
	case cloudprovider.GCP:
		for _, instanceType := range instancetypes.GCPInstanceTypes {
			if insType == instanceType {
				return true
			}
		}
		return false
	default:
		return false
	}
}

// checkIfAWSInstanceTypeIsValid checks if an AWS instance type passed as user input is in one of the instance families supporting NitroTPM.
func checkIfAWSInstanceTypeIsValid(userInput string) bool {
	// Check if user or code does anything weird and tries to pass multiple strings as one
	if strings.Contains(userInput, " ") {
		return false
	}
	if strings.Contains(userInput, ",") {
		return false
	}
	if strings.Contains(userInput, ";") {
		return false
	}

	splitInstanceType := strings.Split(userInput, ".")

	if len(splitInstanceType) != 2 {
		return false
	}

	userDefinedFamily := splitInstanceType[0]
	userDefinedSize := splitInstanceType[1]

	// Check if instace type has at least 4 vCPUs (= contains "xlarge" in its name)
	hasEnoughVCPUs := strings.Contains(userDefinedSize, "xlarge")
	if !hasEnoughVCPUs {
		return false
	}

	// Now check if the user input is a supported family
	// Note that we cannot directly use the family split from the Graviton check above, as some instances are directly specified by their full name and not just the family in general
	for _, supportedFamily := range instancetypes.AWSSupportedInstanceFamilies {
		supportedFamilyLowercase := strings.ToLower(supportedFamily)
		if userDefinedFamily == supportedFamilyLowercase {
			return true
		}
	}

	return false
}

// Validation translation functions for Azure & GCP instance type errors.
func registerTranslateAzureInstanceTypeError(ut ut.Translator) error {
	return ut.Add("azure_instance_type", "{0} must be one of {1}", true)
}

func (c *Config) translateAzureInstanceTypeError(ut ut.Translator, fe validator.FieldError) string {
	// Suggest trusted launch VMs if confidential VMs have been specifically disabled
	var t string
	if c.Provider.Azure != nil && c.Provider.Azure.ConfidentialVM != nil && !*c.Provider.Azure.ConfidentialVM {
		t, _ = ut.T("azure_instance_type", fe.Field(), fmt.Sprintf("%v", instancetypes.AzureTrustedLaunchInstanceTypes))
	} else {
		t, _ = ut.T("azure_instance_type", fe.Field(), fmt.Sprintf("%v", instancetypes.AzureCVMInstanceTypes))
	}

	return t
}

func validateNoPlaceholder(fl validator.FieldLevel) bool {
	return len(getPlaceholderEntries(fl.Field().Interface().(Measurements))) == 0
}

func registerContainsPlaceholderError(ut ut.Translator) error {
	return ut.Add("no_placeholders", "{0} placeholder values (repeated 1234...)", true)
}

func translateContainsPlaceholderError(ut ut.Translator, fe validator.FieldError) string {
	placeholders := getPlaceholderEntries(fe.Value().(Measurements))
	msg := fmt.Sprintf("Measurements %v contain", placeholders)
	if len(placeholders) == 1 {
		msg = fmt.Sprintf("Measurement %v contains", placeholders)
	}

	t, _ := ut.T("no_placeholders", msg)
	return t
}

func getPlaceholderEntries(m Measurements) []uint32 {
	var placeholders []uint32
	placeholder := measurements.PlaceHolderMeasurement()

	for idx, measurement := range m {
		if bytes.Equal(measurement.Expected[:], placeholder.Expected[:]) {
			placeholders = append(placeholders, idx)
		}
	}

	return placeholders
}

func (c *Config) validateK8sVersion(fl validator.FieldLevel) bool {
	// TODO: v2.7: do not create extendedVersion variable and directly validate field from fl.
	// This patch is for compatibility with configs from v2.5 only. Configs specifying k8s
	// the version as MAJOR.MINOR automatically get extended with the respective patch version.
	configVersion := compatibility.EnsurePrefixV(fl.Field().String())
	if !semver.IsValid(configVersion) {
		return false
	}

	extendedVersion := K8sVersionFromMajorMinor(semver.MajorMinor(configVersion))
	if extendedVersion == "" {
		return false
	}

	valid := versions.IsSupportedK8sVersion(extendedVersion)
	if !valid {
		return false
	}

	c.KubernetesVersion = extendedVersion
	return true
}

// K8sVersionFromMajorMinor takes a semver in format MAJOR.MINOR
// and returns the version in format MAJOR.MINOR.PATCH with the
// supported patch version as PATCH.
func K8sVersionFromMajorMinor(version string) string {
	switch version {
	case semver.MajorMinor(string(versions.V1_24)):
		return string(versions.V1_24)
	case semver.MajorMinor(string(versions.V1_25)):
		return string(versions.V1_25)
	case semver.MajorMinor(string(versions.V1_26)):
		return string(versions.V1_26)
	default:
		return ""
	}
}

func registerVersionCompatibilityError(ut ut.Translator) error {
	return ut.Add("version_compatibility", "{0} specifies an invalid version: {1}", true)
}

func translateVersionCompatibilityError(ut ut.Translator, fe validator.FieldError) string {
	err := validateVersionCompatibilityHelper(fe.Field(), fe.Value().(string))
	var msg string

	switch {
	case errors.Is(err, compatibility.ErrSemVer):
		msg = fmt.Sprintf("configured version (%s) does not adhere to SemVer syntax", fe.Value().(string))
	case errors.Is(err, compatibility.ErrMajorMismatch):
		msg = fmt.Sprintf("the CLI's major version (%s) has to match your configured major version (%s)", constants.VersionInfo, fe.Value().(string))
	case errors.Is(err, compatibility.ErrMinorDrift):
		msg = fmt.Sprintf("the CLI's minor version (%s) and the configured version (%s) are more than one minor version apart", constants.VersionInfo, fe.Value().(string))
	case errors.Is(err, compatibility.ErrOutdatedCLI):
		msg = fmt.Sprintf("the CLI's version (%s) is older than the configured version (%s)", constants.VersionInfo, fe.Value().(string))
	default:
		msg = err.Error()
	}

	t, _ := ut.T("version_compatibility", fe.Field(), msg)

	return t
}

// Check that the validated field and the CLI version are not more than one minor version apart.
func validateVersionCompatibility(fl validator.FieldLevel) bool {
	if err := validateVersionCompatibilityHelper(fl.FieldName(), fl.Field().String()); err != nil {
		return false
	}

	return true
}

func validateVersionCompatibilityHelper(fieldName string, configuredVersion string) error {
	if fieldName == "Image" {
		imageVersion, err := versionsapi.NewVersionFromShortPath(configuredVersion, versionsapi.VersionKindImage)
		if err != nil {
			return err
		}
		configuredVersion = imageVersion.Version
	}

	if fieldName == "MicroserviceVersion" {
		cliVersion := compatibility.EnsurePrefixV(constants.VersionInfo)
		serviceVersion := compatibility.EnsurePrefixV(configuredVersion)
		if semver.Compare(cliVersion, serviceVersion) == -1 {
			return fmt.Errorf("the CLI's version (%s) is older than the configured version (%s)", cliVersion, serviceVersion)
		}
	}

	return compatibility.BinaryWith(configuredVersion)
}

func returnsTrue(fl validator.FieldLevel) bool {
	return true
}

// validateUpgradeConfig prints a warning to STDERR and validates the field successfully.
func validateUpgradeConfig(sl validator.StructLevel) {
	config, ok := sl.Current().Interface().(UpgradeConfig)
	if !ok {
		sl.ReportError(config, "upgrade", "UpgradeConfig", "malformed_upgrade_config", "")
	}
	// A valid `upgrade` section will always have a non-nil Measurements map.
	// Only print a warning if the user actually specified an upgrade section.
	if config.Image == "" && config.CSP == cloudprovider.Unknown && config.Measurements == nil {
		return
	}
	fmt.Fprintln(os.Stderr, "WARNING: the config key `upgrade` will be deprecated in an upcoming version. Please check the documentation for more information.")
}

func registerValidateNameError(ut ut.Translator) error {
	return ut.Add("validate_name", "{0} must be no more than {1} characters long", true)
}

func (c *Config) translateValidateNameError(ut ut.Translator, fe validator.FieldError) string {
	var t string
	if c.Provider.AWS != nil {
		t, _ = ut.T("validate_name", fe.Field(), strconv.Itoa(constants.AWSConstellationNameLength))
	} else {
		t, _ = ut.T("validate_name", fe.Field(), strconv.Itoa(constants.ConstellationNameLength))
	}

	return t
}

// validateName makes sure the name of the constellation is not too long.
// Since this value may differ between providers, we can't simply use built-in validation.
// This also allows us to eventually add more validation rules for constellation names if necessary.
func (c *Config) validateName(fl validator.FieldLevel) bool {
	// TODO: v2.7: remove fallback to defaultName and make name a required field
	if c.Name == "" {
		fmt.Fprintln(os.Stderr, "WARNING: the config key `name` is not set. This key will be required in the next version. Falling back to default name.")
		c.Name = defaultName
	}

	if c.Provider.AWS != nil {
		return len(fl.Field().String()) <= constants.AWSConstellationNameLength
	}
	return len(fl.Field().String()) <= constants.ConstellationNameLength
}

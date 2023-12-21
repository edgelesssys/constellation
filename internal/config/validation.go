/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"golang.org/x/mod/semver"

	"github.com/edgelesssys/constellation/v2/internal/api/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/compatibility"
	"github.com/edgelesssys/constellation/v2/internal/config/disktypes"
	"github.com/edgelesssys/constellation/v2/internal/config/instancetypes"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/role"
	consemver "github.com/edgelesssys/constellation/v2/internal/semver"
	"github.com/edgelesssys/constellation/v2/internal/versions"
)

// ValidationError occurs when the validation of a config fails.
// It contains a list of errors that occurred during validation.
type ValidationError struct {
	validationErrMsgs []string
}

func (e *ValidationError) Error() string {
	return "invalid configuration"
}

// LongMessage prints the errors that occurred during validation in a verbose and user friendly way.
func (e *ValidationError) LongMessage() string {
	msg := "Problems validating config file:\n"
	for _, ve := range e.validationErrMsgs {
		msg += fmt.Sprintf("\t%s\n", ve)
	}
	msg += "Fix the invalid entries or generate a new configuration using `constellation config generate`"
	return msg
}

func (e *ValidationError) messagesCount() int {
	return len(e.validationErrMsgs)
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
		errorMsg = "The configured version is not a valid semantic version\n"
	case semver.Compare(configured, minVersion) == -1:
		errorMsg = fmt.Sprintf("The configured version %s is older than the oldest version supported by this CLI: %s\n", configured, minVersion)
	case semver.Compare(configured, maxVersion) == 1:
		errorMsg = fmt.Sprintf("The configured version %s is newer than the newest version supported by this CLI: %s\n", configured, maxVersion)
	}

	errorMsg = errorMsg + fmt.Sprintf("Supported versions: %s", strings.Join(validVersionsSorted, " "))

	t, _ := ut.T("supported_k8s_version", fe.Field(), errorMsg)

	return t
}

func validateAWSRegionField(fl validator.FieldLevel) bool {
	return ValidateAWSRegion(fl.Field().String())
}

func validateAWSZoneField(fl validator.FieldLevel) bool {
	return ValidateAWSZone(fl.Field().String())
}

func validateAzureZoneField(fl validator.FieldLevel) bool {
	return ValidateAzureZone(fl.Field().String())
}

func validateGCPZoneField(fl validator.FieldLevel) bool {
	return ValidateGCPZone(fl.Field().String())
}

func validateOpenStackRegionField(fl validator.FieldLevel) bool {
	return ValidateOpenStackRegion(fl.Field().String())
}

// ValidateAWSZone validates that the zone is in the correct format.
func ValidateAWSZone(zone string) bool {
	awsZoneRegex := regexp.MustCompile(`^\w+-\w+-[1-9][abc]$`)
	return awsZoneRegex.MatchString(zone)
}

// ValidateAzureZone validates that the zone is in the correct format.
func ValidateAzureZone(zone string) bool {
	azureZoneRegex := regexp.MustCompile(`^$|^\d+(?:,\d+)*$`)
	return azureZoneRegex.MatchString(zone)
}

// ValidateGCPZone validates that the zone is in the correct format.
func ValidateGCPZone(zone string) bool {
	gcpZoneRegex := regexp.MustCompile(`^[\w-]+-[a-z]$`)
	return gcpZoneRegex.MatchString(zone)
}

// ValidateOpenStackRegion validates that the region is in the correct format.
func ValidateOpenStackRegion(region string) bool {
	return len(region) > 0
}

// ValidateAWSRegion validates that the region is in the correct format.
func ValidateAWSRegion(region string) bool {
	awsRegionRegex := regexp.MustCompile(`^\w+-\w+-[1-9]$`)
	return awsRegionRegex.MatchString(region)
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
	if provider.OpenStack != nil {
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

func validateAttestation(sl validator.StructLevel) {
	attestation := sl.Current().Interface().(AttestationConfig)
	attestationCount := 0

	if attestation.AWSSEVSNP != nil {
		attestationCount++
	}
	if attestation.AWSNitroTPM != nil {
		attestationCount++
	}
	if attestation.AzureSEVSNP != nil {
		attestationCount++
	}
	if attestation.AzureTrustedLaunch != nil {
		attestationCount++
	}
	if attestation.GCPSEVES != nil {
		attestationCount++
	}
	if attestation.QEMUVTPM != nil {
		attestationCount++
	}

	if attestationCount < 1 {
		sl.ReportError(attestation, "Attestation", "Attestation", "no_attestation", "")
	} else if attestationCount > 1 {
		sl.ReportError(attestation, "Attestation", "Attestation", "more_than_one_attestation", "")
	}
}

func validateNodeGroups(sl validator.StructLevel) {
	nodeGroups := sl.Current().Interface().(Config).NodeGroups
	defaultControlPlaneGroup, hasDefaultControlPlaneGroup := nodeGroups[constants.DefaultControlPlaneGroupName]
	defaultWorkerGroup, hasDefaultWorkerGroup := nodeGroups[constants.DefaultWorkerGroupName]

	if !hasDefaultControlPlaneGroup {
		sl.ReportError(nodeGroups, "NodeGroups", "NodeGroups", "no_default_control_plane_group", "")
	}
	if !hasDefaultWorkerGroup {
		sl.ReportError(nodeGroups, "NodeGroups", "NodeGroups", "no_default_worker_group", "")
	}

	if hasDefaultControlPlaneGroup {
		if defaultControlPlaneGroup.Role != role.ControlPlane.TFString() {
			sl.ReportError(nodeGroups, "NodeGroups", "NodeGroups", "control_plane_group_role_mismatch", "")
		}
		if defaultControlPlaneGroup.InitialCount < 1 {
			sl.ReportError(nodeGroups, "NodeGroups", "NodeGroups", "control_plane_group_initial_count", "")
		}
	}
	if hasDefaultWorkerGroup {
		if defaultWorkerGroup.Role != role.Worker.TFString() {
			sl.ReportError(nodeGroups, "NodeGroups", "NodeGroups", "worker_group_role_mismatch", "")
		}
	}
}

func translateNoAttestationError(ut ut.Translator, fe validator.FieldError) string {
	t, _ := ut.T("no_attestation", fe.Field())

	return t
}

func registerNoAttestationError(ut ut.Translator) error {
	return ut.Add("no_attestation", "{0}: No attestation has been defined (requires either awsSEVSNP, awsNitroTPM, azureSEVSNP, azureTrustedLaunch, gcpSEVES, or qemuVTPM)", true)
}

func translateNoDefaultControlPlaneGroupError(ut ut.Translator, fe validator.FieldError) string {
	t, _ := ut.T("no_default_control_plane_group", fe.Field())

	return t
}

func registerNoDefaultControlPlaneGroupError(ut ut.Translator) error {
	return ut.Add("no_default_control_plane_group", "{0}: No default control plane group (control_plane_default) has been defined", true)
}

func translateNoDefaultWorkerGroupError(ut ut.Translator, fe validator.FieldError) string {
	t, _ := ut.T("no_default_worker_group", fe.Field())

	return t
}

func registerNoDefaultWorkerGroupError(ut ut.Translator) error {
	return ut.Add("no_default_worker_group", "{0}: No default worker group (worker_default) has been defined", true)
}

func translateControlPlaneGroupInitialCountError(ut ut.Translator, fe validator.FieldError) string {
	t, _ := ut.T("control_plane_group_initial_count", fe.Field())

	return t
}

func registerControlPlaneGroupInitialCountError(ut ut.Translator) error {
	return ut.Add("control_plane_group_initial_count", "{0}: The default control plane group (control_plane_default) must have at least one node", true)
}

func translateControlPlaneGroupRoleMismatchError(ut ut.Translator, fe validator.FieldError) string {
	t, _ := ut.T("control_plane_group_role_mismatch", fe.Field())

	return t
}

func registerControlPlaneGroupRoleMismatchError(ut ut.Translator) error {
	return ut.Add("control_plane_group_role_mismatch", "{0}: The default control plane group (control_plane_default) must have the \"control-plane\" role", true)
}

func translateWorkerGroupRoleMismatchError(ut ut.Translator, fe validator.FieldError) string {
	t, _ := ut.T("worker_group_role_mismatch", fe.Field())

	return t
}

func registerWorkerGroupRoleMismatchError(ut ut.Translator) error {
	return ut.Add("worker_group_role_mismatch", "{0}: The default worker group (worker_default) must have the \"worker\" role", true)
}

func registerValidZoneError(ut ut.Translator) error {
	return ut.Add("valid_zone", "{0}: has invalid format: {1}", true)
}

func (c *Config) translateValidZoneError(ut ut.Translator, fe validator.FieldError) string {
	var t string
	switch c.GetProvider() {
	case cloudprovider.AWS:
		t, _ = ut.T("valid_zone", fe.Field(), "field must be of format eu-central-1a")
	case cloudprovider.Azure:
		t, _ = ut.T("valid_zone", fe.Field(), "field must be a comma separated list of zones: 1,2,3 (or empty for all zones)")
	case cloudprovider.GCP:
		t, _ = ut.T("valid_zone", fe.Field(), "field must be of format europe-west3-b")
	default:
		t, _ = ut.T("valid_zone", fe.Field(), "field must not be empty")
	}
	return t
}

func registerAWSRegionError(ut ut.Translator) error {
	return ut.Add("aws_region", "{0}: has invalid format: {1}", true)
}

func translateAWSRegionError(ut ut.Translator, fe validator.FieldError) string {
	t, _ := ut.T("aws_region", fe.Field(), "field must be of format eu-central-1")

	return t
}

func translateAWSZoneError(ut ut.Translator, fe validator.FieldError) string {
	t, _ := ut.T("aws_zone", fe.Field(), "field must be of format eu-central-1a")

	return t
}

func registerAWSZoneError(ut ut.Translator) error {
	return ut.Add("aws_zone", "{0}: has invalid format: {1}", true)
}

func registerMoreThanOneAttestationError(ut ut.Translator) error {
	return ut.Add("more_than_one_attestation", "{0}: Only one attestation can be defined ({1} are defined)", true)
}

func (c *Config) translateMoreThanOneAttestationError(ut ut.Translator, fe validator.FieldError) string {
	definedAttestations := make([]string, 0)

	if c.Attestation.AWSNitroTPM != nil {
		definedAttestations = append(definedAttestations, "AWSNitroTPM")
	}
	if c.Attestation.AWSSEVSNP != nil {
		definedAttestations = append(definedAttestations, "AWSSEVSNP")
	}
	if c.Attestation.AzureSEVSNP != nil {
		definedAttestations = append(definedAttestations, "AzureSEVSNP")
	}
	if c.Attestation.AzureTrustedLaunch != nil {
		definedAttestations = append(definedAttestations, "AzureTrustedLaunch")
	}
	if c.Attestation.GCPSEVES != nil {
		definedAttestations = append(definedAttestations, "GCPSEVES")
	}
	if c.Attestation.QEMUVTPM != nil {
		definedAttestations = append(definedAttestations, "QEMUVTPM")
	}

	t, _ := ut.T("more_than_one_attestation", fe.Field(), strings.Join(definedAttestations, ", "))

	return t
}

func registerTranslateDiskTypeError(ut ut.Translator) error {
	return ut.Add("disk_type", "{0} must be one of {1}", true)
}

func (c *Config) translateDiskTypeError(ut ut.Translator, fe validator.FieldError) string {
	var supported []string
	switch c.GetProvider() {
	case cloudprovider.AWS:
		supported = disktypes.AWSDiskTypes
	case cloudprovider.Azure:
		supported = disktypes.AzureDiskTypes
	case cloudprovider.GCP:
		supported = disktypes.GCPDiskTypes
	}

	t, _ := ut.T("disk_type", fe.Field(), fmt.Sprintf("%v", supported))
	return t
}

func (c *Config) registerTranslateInstanceTypeError(ut ut.Translator) error {
	switch c.GetProvider() {
	case cloudprovider.AWS:
		return registerTranslateAWSInstanceTypeError(ut)
	case cloudprovider.Azure:
		return registerTranslateAzureInstanceTypeError(ut)
	case cloudprovider.GCP:
		return registerTranslateGCPInstanceTypeError(ut)
	}
	return ut.Add("instance_type", "{0} is an invalid instance type", true)
}

func (c *Config) translateInstanceTypeError(ut ut.Translator, fe validator.FieldError) string {
	switch c.GetProvider() {
	case cloudprovider.AWS:
		return c.translateAWSInstanceTypeError(ut, fe)
	case cloudprovider.Azure:
		return c.translateAzureInstanceTypeError(ut, fe)
	case cloudprovider.GCP:
		return translateGCPInstanceTypeError(ut, fe)
	}
	t, _ := ut.T("instance_type", fe.Field())

	return t
}

func registerTranslateAWSInstanceTypeError(ut ut.Translator) error {
	return ut.Add("instance_type", "{0} must be an instance from one of the following families types with size xlarge or higher: {1}", true)
}

func (c *Config) translateAWSInstanceTypeError(ut ut.Translator, fe validator.FieldError) string {
	var t string

	attestVariant := c.GetAttestationConfig().GetVariant()

	instances := instancetypes.AWSSNPSupportedInstanceFamilies
	if attestVariant.Equal(variant.AWSNitroTPM{}) {
		instances = instancetypes.AWSSupportedInstanceFamilies
	}

	t, _ = ut.T("instance_type", fe.Field(), fmt.Sprintf("%v", instances))

	return t
}

func registerTranslateGCPInstanceTypeError(ut ut.Translator) error {
	return ut.Add("instance_type", fmt.Sprintf("{0} must be one of %v", instancetypes.GCPInstanceTypes), true)
}

func translateGCPInstanceTypeError(ut ut.Translator, fe validator.FieldError) string {
	t, _ := ut.T("instance_type", fe.Field())

	return t
}

// Validation translation functions for Azure & GCP instance type errors.
func registerTranslateAzureInstanceTypeError(ut ut.Translator) error {
	return ut.Add("instance_type", "{0} must be one of {1}", true)
}

func (c *Config) translateAzureInstanceTypeError(ut ut.Translator, fe validator.FieldError) string {
	// Suggest trusted launch VMs if confidential VMs have been specifically disabled
	var t string

	attestVariant := c.GetAttestationConfig().GetVariant()

	instances := instancetypes.AzureCVMInstanceTypes
	if attestVariant.Equal(variant.AzureTrustedLaunch{}) {
		instances = instancetypes.AzureTrustedLaunchInstanceTypes
	}

	t, _ = ut.T("instance_type", fe.Field(), fmt.Sprintf("%v", instances))

	return t
}

// Validation translation functions for Provider errors.
func registerNoProviderError(ut ut.Translator) error {
	return ut.Add("no_provider", "{0}: No provider has been defined (requires either Azure, GCP, OpenStack or QEMU)", true)
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
	if c.Provider.OpenStack != nil {
		definedProviders = append(definedProviders, "OpenStack")
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
		return isSupportedAWSInstanceType(insType, acceptNonCVM)
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
	case cloudprovider.OpenStack, cloudprovider.QEMU:
		return true
	default:
		return false
	}
}

// isSupportedAWSInstanceType checks if an AWS instance type passed as user input is in one of the supported instance types.
func isSupportedAWSInstanceType(userInput string, acceptNonCVM bool) bool {
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

	instances := instancetypes.AWSSNPSupportedInstanceFamilies
	if acceptNonCVM {
		instances = instancetypes.AWSSupportedInstanceFamilies
	}

	// Now check if the user input is a supported family
	// Note that we cannot directly use the family split from the Graviton check above, as some instances are directly specified by their full name and not just the family in general
	for _, supportedFamily := range instances {
		supportedFamilyLowercase := strings.ToLower(supportedFamily)
		if userDefinedFamily == supportedFamilyLowercase {
			return true
		}
	}

	return false
}

func validateNoPlaceholder(fl validator.FieldLevel) bool {
	return len(getPlaceholderEntries(fl.Field().Interface().(measurements.M))) == 0
}

// validateMeasurement acts like validateNoPlaceholder, but is used for the measurements.Measurement type.
func validateMeasurement(sl validator.StructLevel) {
	measurement := sl.Current().Interface().(measurements.Measurement)
	actual := measurement.Expected
	placeHolder := measurements.PlaceHolderMeasurement(measurements.PCRMeasurementLength).Expected
	if bytes.Equal(actual, placeHolder) {
		sl.ReportError(measurement, "launchMeasurement", "launchMeasurement", "no_placeholders", "")
	}
}

func registerContainsPlaceholderError(ut ut.Translator) error {
	return ut.Add("no_placeholders", "{0} placeholder values (repeated 1234...)", true)
}

func translateContainsPlaceholderError(ut ut.Translator, fe validator.FieldError) string {
	var msg string
	switch fe.Field() {
	case "launchMeasurement":
		msg = "launchMeasurement contains"
	case "measurements":
		placeholders := getPlaceholderEntries(fe.Value().(measurements.M))
		msg = fmt.Sprintf("measurements %v contain", placeholders)
		if len(placeholders) == 1 {
			msg = fmt.Sprintf("measurement %v contains", placeholders)
		}
	}

	t, _ := ut.T("no_placeholders", msg)
	return t
}

func getPlaceholderEntries(m measurements.M) []uint32 {
	var placeholders []uint32
	placeholderTDX := measurements.PlaceHolderMeasurement(measurements.TDXMeasurementLength)
	placeholderTPM := measurements.PlaceHolderMeasurement(measurements.PCRMeasurementLength)

	for idx, measurement := range m {
		if bytes.Equal(measurement.Expected, placeholderTDX.Expected) ||
			bytes.Equal(measurement.Expected, placeholderTPM.Expected) {
			placeholders = append(placeholders, idx)
		}
	}

	return placeholders
}

// validateK8sVersion does not check the patch version.
func (c *Config) validateK8sVersion(fl validator.FieldLevel) bool {
	_, err := versions.NewValidK8sVersion(compatibility.EnsurePrefixV(fl.Field().String()), false)
	return err == nil
}

func registerImageCompatibilityError(ut ut.Translator) error {
	return ut.Add("image_compatibility", "{0} specifies an invalid version: {1}", true)
}

// Check that the validated field and the CLI version are not more than one minor version apart.
func validateVersionCompatibility(fl validator.FieldLevel) bool {
	binaryVersion := constants.BinaryVersion()
	if err := validateImageCompatibilityHelper(binaryVersion, fl.FieldName(), fl.Field().String()); err != nil {
		return false
	}

	return true
}

func validateImageCompatibilityHelper(binaryVersion consemver.Semver, fieldName, configuredVersion string) error {
	if fieldName == "image" {
		imageVersion, err := versionsapi.NewVersionFromShortPath(configuredVersion, versionsapi.VersionKindImage)
		if err != nil {
			return err
		}
		configuredVersion = imageVersion.Version()
	}

	return compatibility.BinaryWith(binaryVersion.String(), configuredVersion)
}

func translateImageCompatibilityError(ut ut.Translator, fe validator.FieldError) string {
	binaryVersion := constants.BinaryVersion()
	err := validateImageCompatibilityHelper(binaryVersion, fe.Field(), fe.Value().(string))

	msg := msgFromCompatibilityError(err, binaryVersion.String(), fe.Value().(string))

	t, _ := ut.T("image_compatibility", fe.Field(), msg)

	return t
}

// msgFromCompatibilityError translates compatibility errors into user-facing error messages.
func msgFromCompatibilityError(err error, binaryVersion, fieldValue string) string {
	switch {
	case errors.Is(err, compatibility.ErrSemVer):
		return fmt.Sprintf("configured version (%s) does not adhere to SemVer syntax", fieldValue)
	case errors.Is(err, compatibility.ErrMajorMismatch):
		return fmt.Sprintf("the CLI's major version (%s) has to match your configured major version (%s). Use --force to ignore the version mismatch.", binaryVersion, fieldValue)
	case errors.Is(err, compatibility.ErrMinorDrift):
		return fmt.Sprintf("the CLI's minor version (%s) and the configured version (%s) are more than one minor version apart. Use --force to ignore the version mismatch.", binaryVersion, fieldValue)
	case errors.Is(err, compatibility.ErrOutdatedCLI):
		return fmt.Sprintf("the CLI's version (%s) is older than the configured version (%s). Use --force to ignore the version mismatch.", binaryVersion, fieldValue)
	default:
		return err.Error()
	}
}

// ValidateMicroserviceVersion checks that the version of the microservice is compatible with the binary version.
func ValidateMicroserviceVersion(binaryVersion, version consemver.Semver) error {
	// Major versions always have to match.
	if binaryVersion.Major() != version.Major() {
		return compatibility.ErrMajorMismatch
	}
	// Allow newer CLIs (for upgrades), but dissallow newer service versions.
	if binaryVersion.Compare(version) == -1 {
		return compatibility.ErrOutdatedCLI
	}
	// Abort if minor version drift between CLI and versionA value is greater than 1.
	if binaryVersion.Minor()-version.Minor() > 1 {
		return compatibility.ErrMinorDrift
	}

	return nil
}

func returnsTrue(_ validator.FieldLevel) bool {
	return true
}

func (c *Config) registerValidateNameError(ut ut.Translator) error {
	if c.Provider.AWS != nil {
		return ut.Add("validate_name", "{0} must be no more than {1} characters long and contain only lowercase letters", true)
	}
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
	if c.Provider.AWS != nil {
		name := fl.Field().String()
		return strings.ToLower(name) == name && len(name) <= constants.AWSConstellationNameLength
	}
	return len(fl.Field().String()) <= constants.ConstellationNameLength
}

func warnDeprecated(fl validator.FieldLevel) bool {
	fmt.Fprintf(
		os.Stderr,
		"WARNING: The config key %q is deprecated and will be removed in an upcoming version.\n",
		fl.FieldName(),
	)
	return true
}

func (c *Config) validateNodeGroupZoneField(fl validator.FieldLevel) bool {
	switch c.GetProvider() {
	case cloudprovider.AWS:
		return validateAWSZoneField(fl)
	case cloudprovider.Azure:
		return validateAzureZoneField(fl)
	case cloudprovider.GCP:
		return validateGCPZoneField(fl)
	case cloudprovider.OpenStack:
		return validateOpenStackRegionField(fl)
	case cloudprovider.QEMU:
		// QEMU does not use zones
		return true
	case cloudprovider.Unknown:
		return true
	}
	// this indicates we are missing a case for a new provider
	return false
}

func (c *Config) validateInstanceType(fl validator.FieldLevel) bool {
	acceptNonCVM := false

	if c.GetAttestationConfig().GetVariant().Equal(variant.AzureTrustedLaunch{}) {
		acceptNonCVM = true
	}
	if c.GetAttestationConfig().GetVariant().Equal(variant.AWSNitroTPM{}) {
		acceptNonCVM = true
	}

	return validInstanceTypeForProvider(fl.Field().String(), acceptNonCVM, c.GetProvider())
}

func (c *Config) validateStateDiskTypeField(fl validator.FieldLevel) bool {
	switch c.GetProvider() {
	case cloudprovider.AWS:
		return validateAWSStateDiskField(fl)
	case cloudprovider.Azure:
		return validateAzureStateDiskField(fl)
	case cloudprovider.GCP:
		return validateGCPStateDiskField(fl)
	case cloudprovider.OpenStack:
		return validateOpenStackStateDiskField(fl)
	case cloudprovider.QEMU:
		return validateQEMUStateDiskField(fl)
	case cloudprovider.Unknown:
		return true

	}
	// this indicates we are missing a case for a new provider
	return false
}

func validateAWSStateDiskField(fl validator.FieldLevel) bool {
	gotDiskField := fl.Field().String()
	for _, diskType := range disktypes.AWSDiskTypes {
		if gotDiskField == diskType {
			return true
		}
	}
	return false
}

func validateAzureStateDiskField(fl validator.FieldLevel) bool {
	gotDiskField := fl.Field().String()
	for _, diskType := range disktypes.AzureDiskTypes {
		if gotDiskField == diskType {
			return true
		}
	}
	return false
}

func validateGCPStateDiskField(fl validator.FieldLevel) bool {
	gotDiskField := fl.Field().String()
	for _, diskType := range disktypes.GCPDiskTypes {
		if gotDiskField == diskType {
			return true
		}
	}
	return false
}

func validateOpenStackStateDiskField(fl validator.FieldLevel) bool {
	return len(fl.Field().String()) > 0
}

func validateQEMUStateDiskField(_ validator.FieldLevel) bool {
	return true
}

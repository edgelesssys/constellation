/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/attestation/attestation"
	azpolicy "github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"

	"github.com/edgelesssys/constellation/v2/cli/internal/clusterid"
	"github.com/edgelesssys/constellation/v2/cli/internal/image"
	"github.com/edgelesssys/constellation/v2/cli/internal/libvirt"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/variant"
)

// Creator creates cloud resources.
type Creator struct {
	out                io.Writer
	image              imageFetcher
	newTerraformClient func(ctx context.Context) (terraformClient, error)
	newLibvirtRunner   func() libvirtRunner
	newRawDownloader   func() rawDownloader
	policyPatcher      PolicyPatcher
}

// NewCreator creates a new creator.
func NewCreator(out io.Writer) *Creator {
	return &Creator{
		out:   out,
		image: image.New(),
		newTerraformClient: func(ctx context.Context) (terraformClient, error) {
			return terraform.New(ctx, constants.TerraformWorkingDir)
		},
		newLibvirtRunner: func() libvirtRunner {
			return libvirt.New()
		},
		newRawDownloader: func() rawDownloader {
			return image.NewDownloader()
		},
		policyPatcher: policyPatcher{},
	}
}

// CreateOptions are the options for creating a Constellation cluster.
type CreateOptions struct {
	Provider          cloudprovider.Provider
	Config            *config.Config
	InsType           string
	ControlPlaneCount int
	WorkerCount       int
	image             string
	TFLogLevel        terraform.LogLevel
}

// Create creates the handed amount of instances and all the needed resources.
func (c *Creator) Create(ctx context.Context, opts CreateOptions) (clusterid.File, error) {
	image, err := c.image.FetchReference(ctx, opts.Config)
	if err != nil {
		return clusterid.File{}, fmt.Errorf("fetching image reference: %w", err)
	}
	opts.image = image

	switch opts.Provider {
	case cloudprovider.AWS:
		cl, err := c.newTerraformClient(ctx)
		if err != nil {
			return clusterid.File{}, err
		}
		defer cl.RemoveInstaller()
		return c.createAWS(ctx, cl, opts)
	case cloudprovider.GCP:
		cl, err := c.newTerraformClient(ctx)
		if err != nil {
			return clusterid.File{}, err
		}
		defer cl.RemoveInstaller()
		return c.createGCP(ctx, cl, opts)
	case cloudprovider.Azure:
		cl, err := c.newTerraformClient(ctx)
		if err != nil {
			return clusterid.File{}, err
		}
		defer cl.RemoveInstaller()
		return c.createAzure(ctx, cl, opts)
	case cloudprovider.OpenStack:
		cl, err := c.newTerraformClient(ctx)
		if err != nil {
			return clusterid.File{}, err
		}
		defer cl.RemoveInstaller()
		return c.createOpenStack(ctx, cl, opts)
	case cloudprovider.QEMU:
		if runtime.GOARCH != "amd64" || runtime.GOOS != "linux" {
			return clusterid.File{}, fmt.Errorf("creation of a QEMU based Constellation is not supported for %s/%s", runtime.GOOS, runtime.GOARCH)
		}
		cl, err := c.newTerraformClient(ctx)
		if err != nil {
			return clusterid.File{}, err
		}
		defer cl.RemoveInstaller()
		lv := c.newLibvirtRunner()
		qemuOpts := qemuCreateOptions{
			source:        image,
			CreateOptions: opts,
		}
		return c.createQEMU(ctx, cl, lv, qemuOpts)
	default:
		return clusterid.File{}, fmt.Errorf("unsupported cloud provider: %s", opts.Provider)
	}
}

func (c *Creator) createAWS(ctx context.Context, cl terraformClient, opts CreateOptions) (idFile clusterid.File, retErr error) {
	vars := terraform.AWSClusterVariables{
		CommonVariables: terraform.CommonVariables{
			Name:               opts.Config.Name,
			CountControlPlanes: opts.ControlPlaneCount,
			CountWorkers:       opts.WorkerCount,
			StateDiskSizeGB:    opts.Config.StateDiskSizeGB,
		},
		StateDiskType:          opts.Config.Provider.AWS.StateDiskType,
		Region:                 opts.Config.Provider.AWS.Region,
		Zone:                   opts.Config.Provider.AWS.Zone,
		InstanceType:           opts.InsType,
		AMIImageID:             opts.image,
		IAMProfileControlPlane: opts.Config.Provider.AWS.IAMProfileControlPlane,
		IAMProfileWorkerNodes:  opts.Config.Provider.AWS.IAMProfileWorkerNodes,
		Debug:                  opts.Config.IsDebugCluster(),
	}

	if err := cl.PrepareWorkspace(path.Join("terraform", strings.ToLower(cloudprovider.AWS.String())), &vars); err != nil {
		return clusterid.File{}, err
	}

	defer rollbackOnError(c.out, &retErr, &rollbackerTerraform{client: cl}, opts.TFLogLevel)
	tfOutput, err := cl.CreateCluster(ctx, opts.TFLogLevel)
	if err != nil {
		return clusterid.File{}, err
	}

	return clusterid.File{
		CloudProvider: cloudprovider.AWS,
		InitSecret:    []byte(tfOutput.Secret),
		IP:            tfOutput.IP,
		UID:           tfOutput.UID,
	}, nil
}

func (c *Creator) createGCP(ctx context.Context, cl terraformClient, opts CreateOptions) (idFile clusterid.File, retErr error) {
	vars := terraform.GCPClusterVariables{
		CommonVariables: terraform.CommonVariables{
			Name:               opts.Config.Name,
			CountControlPlanes: opts.ControlPlaneCount,
			CountWorkers:       opts.WorkerCount,
			StateDiskSizeGB:    opts.Config.StateDiskSizeGB,
		},
		Project:         opts.Config.Provider.GCP.Project,
		Region:          opts.Config.Provider.GCP.Region,
		Zone:            opts.Config.Provider.GCP.Zone,
		CredentialsFile: opts.Config.Provider.GCP.ServiceAccountKeyPath,
		InstanceType:    opts.InsType,
		StateDiskType:   opts.Config.Provider.GCP.StateDiskType,
		ImageID:         opts.image,
		Debug:           opts.Config.IsDebugCluster(),
	}

	if err := cl.PrepareWorkspace(path.Join("terraform", strings.ToLower(cloudprovider.GCP.String())), &vars); err != nil {
		return clusterid.File{}, err
	}

	defer rollbackOnError(c.out, &retErr, &rollbackerTerraform{client: cl}, opts.TFLogLevel)
	tfOutput, err := cl.CreateCluster(ctx, opts.TFLogLevel)
	if err != nil {
		return clusterid.File{}, err
	}

	return clusterid.File{
		CloudProvider: cloudprovider.GCP,
		InitSecret:    []byte(tfOutput.Secret),
		IP:            tfOutput.IP,
		UID:           tfOutput.UID,
	}, nil
}

func (c *Creator) createAzure(ctx context.Context, cl terraformClient, opts CreateOptions) (idFile clusterid.File, retErr error) {
	vars := terraform.AzureClusterVariables{
		CommonVariables: terraform.CommonVariables{
			Name:               opts.Config.Name,
			CountControlPlanes: opts.ControlPlaneCount,
			CountWorkers:       opts.WorkerCount,
			StateDiskSizeGB:    opts.Config.StateDiskSizeGB,
		},
		Location:             opts.Config.Provider.Azure.Location,
		ResourceGroup:        opts.Config.Provider.Azure.ResourceGroup,
		UserAssignedIdentity: opts.Config.Provider.Azure.UserAssignedIdentity,
		InstanceType:         opts.InsType,
		StateDiskType:        opts.Config.Provider.Azure.StateDiskType,
		ImageID:              opts.image,
		SecureBoot:           *opts.Config.Provider.Azure.SecureBoot,
		CreateMAA:            opts.Config.GetAttestationConfig().GetVariant().Equal(variant.AzureSEVSNP{}),
		Debug:                opts.Config.IsDebugCluster(),
	}

	vars.ConfidentialVM = opts.Config.GetAttestationConfig().GetVariant().Equal(variant.AzureSEVSNP{})

	vars = normalizeAzureURIs(vars)

	if err := cl.PrepareWorkspace(path.Join("terraform", strings.ToLower(cloudprovider.Azure.String())), &vars); err != nil {
		return clusterid.File{}, err
	}

	defer rollbackOnError(c.out, &retErr, &rollbackerTerraform{client: cl}, opts.TFLogLevel)
	tfOutput, err := cl.CreateCluster(ctx, opts.TFLogLevel)
	if err != nil {
		return clusterid.File{}, err
	}

	if vars.CreateMAA {
		// Patch the attestation policy to allow the cluster to boot while having secure boot disabled.
		if err := c.policyPatcher.Patch(ctx, tfOutput.AttestationURL); err != nil {
			return clusterid.File{}, err
		}
	}

	return clusterid.File{
		CloudProvider:  cloudprovider.Azure,
		IP:             tfOutput.IP,
		InitSecret:     []byte(tfOutput.Secret),
		UID:            tfOutput.UID,
		AttestationURL: tfOutput.AttestationURL,
	}, nil
}

// PolicyPatcher interacts with Azure to update the attestation policy.
type PolicyPatcher interface {
	Patch(ctx context.Context, attestationURL string) error
}

type policyPatcher struct{}

// Patch updates the attestation policy to the base64-encoded attestation policy JWT for the given attestation URL.
// https://learn.microsoft.com/en-us/azure/attestation/author-sign-policy#next-steps
func (p policyPatcher) Patch(ctx context.Context, attestationURL string) error {
	// hacky way to update the MAA attestation policy. This should be changed as soon as either the Terraform provider supports it
	// or the Go SDK gets updated to a recent API version.
	// https://github.com/hashicorp/terraform-provider-azurerm/issues/20804
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return fmt.Errorf("retrieving default Azure credentials: %w", err)
	}
	token, err := cred.GetToken(ctx, azpolicy.TokenRequestOptions{
		Scopes: []string{"https://attest.azure.net/.default"},
	})
	if err != nil {
		return fmt.Errorf("retrieving token from default Azure credentials: %w", err)
	}

	client := attestation.NewPolicyClient()

	// azureGuest is the id for the "Azure VM" attestation type. Other types are documented here:
	// https://learn.microsoft.com/en-us/rest/api/attestation/policy/set
	req, err := client.SetPreparer(ctx, attestationURL, "azureGuest", p.encodeAttestationPolicy())
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.Token))
	if err != nil {
		return fmt.Errorf("preparing request: %w", err)
	}

	resp, err := client.Send(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("updating attestation policy: unexpected status code: %s", resp.Status)
	}

	return nil
}

// encodeAttestationPolicy encodes the base64-encoded attestation policy in the JWS format specified here:
// https://learn.microsoft.com/en-us/azure/attestation/author-sign-policy#creating-the-policy-file-in-json-web-signature-format
func (p policyPatcher) encodeAttestationPolicy() string {
	const policy = `
                version= 1.0;
                authorizationrules
                {
                    [type=="x-ms-azurevm-default-securebootkeysvalidated", value==false] => deny();
                    [type=="x-ms-azurevm-debuggersdisabled", value==false] => deny();
                    // The line below was edited by the Constellation CLI. Do not edit manually.
                    //[type=="secureboot", value==false] => deny();
                    [type=="x-ms-azurevm-signingdisabled", value==false] => deny();
                    [type=="x-ms-azurevm-dbvalidated", value==false] => deny();
                    [type=="x-ms-azurevm-dbxvalidated", value==false] => deny();
                    => permit();
                };
                issuancerules
                {
                };`
	encodedPolicy := base64.RawURLEncoding.EncodeToString([]byte(policy))
	const header = `{"alg":"none"}`
	payload := fmt.Sprintf(`{"AttestationPolicy":"%s"}`, encodedPolicy)

	encodedHeader := base64.RawURLEncoding.EncodeToString([]byte(header))
	encodedPayload := base64.RawURLEncoding.EncodeToString([]byte(payload))

	return fmt.Sprintf("%s.%s.", encodedHeader, encodedPayload)
}

// The azurerm Terraform provider enforces its own convention of case sensitivity for Azure URIs which Azure's API itself does not enforce or, even worse, actually returns.
// Let's go loco with case insensitive Regexp here and fix the user input here to be compliant with this arbitrary design decision.
var (
	caseInsensitiveSubscriptionsRegexp          = regexp.MustCompile(`(?i)\/subscriptions\/`)
	caseInsensitiveResourceGroupRegexp          = regexp.MustCompile(`(?i)\/resourcegroups\/`)
	caseInsensitiveProvidersRegexp              = regexp.MustCompile(`(?i)\/providers\/`)
	caseInsensitiveUserAssignedIdentitiesRegexp = regexp.MustCompile(`(?i)\/userassignedidentities\/`)
	caseInsensitiveMicrosoftManagedIdentity     = regexp.MustCompile(`(?i)\/microsoft.managedidentity\/`)
	caseInsensitiveCommunityGalleriesRegexp     = regexp.MustCompile(`(?i)\/communitygalleries\/`)
	caseInsensitiveImagesRegExp                 = regexp.MustCompile(`(?i)\/images\/`)
	caseInsensitiveVersionsRegExp               = regexp.MustCompile(`(?i)\/versions\/`)
)

func normalizeAzureURIs(vars terraform.AzureClusterVariables) terraform.AzureClusterVariables {
	vars.UserAssignedIdentity = caseInsensitiveSubscriptionsRegexp.ReplaceAllString(vars.UserAssignedIdentity, "/subscriptions/")
	vars.UserAssignedIdentity = caseInsensitiveResourceGroupRegexp.ReplaceAllString(vars.UserAssignedIdentity, "/resourceGroups/")
	vars.UserAssignedIdentity = caseInsensitiveProvidersRegexp.ReplaceAllString(vars.UserAssignedIdentity, "/providers/")
	vars.UserAssignedIdentity = caseInsensitiveUserAssignedIdentitiesRegexp.ReplaceAllString(vars.UserAssignedIdentity, "/userAssignedIdentities/")
	vars.UserAssignedIdentity = caseInsensitiveMicrosoftManagedIdentity.ReplaceAllString(vars.UserAssignedIdentity, "/Microsoft.ManagedIdentity/")
	vars.ImageID = caseInsensitiveCommunityGalleriesRegexp.ReplaceAllString(vars.ImageID, "/communityGalleries/")
	vars.ImageID = caseInsensitiveImagesRegExp.ReplaceAllString(vars.ImageID, "/images/")
	vars.ImageID = caseInsensitiveVersionsRegExp.ReplaceAllString(vars.ImageID, "/versions/")

	return vars
}

func (c *Creator) createOpenStack(ctx context.Context, cl terraformClient, opts CreateOptions) (idFile clusterid.File, retErr error) {
	// TODO: Remove this once OpenStack is supported.
	if os.Getenv("CONSTELLATION_OPENSTACK_DEV") != "1" {
		return clusterid.File{}, errors.New("OpenStack isn't supported yet")
	}
	if _, hasOSAuthURL := os.LookupEnv("OS_AUTH_URL"); !hasOSAuthURL && opts.Config.Provider.OpenStack.Cloud == "" {
		return clusterid.File{}, errors.New(
			"neither environment variable OS_AUTH_URL nor cloud name for \"clouds.yaml\" is set. OpenStack authentication requires a set of " +
				"OS_* environment variables that are typically sourced into the current shell with an openrc file " +
				"or a cloud name for \"clouds.yaml\". " +
				"See https://docs.openstack.org/openstacksdk/latest/user/config/configuration.html for more information",
		)
	}

	vars := terraform.OpenStackClusterVariables{
		CommonVariables: terraform.CommonVariables{
			Name:               opts.Config.Name,
			CountControlPlanes: opts.ControlPlaneCount,
			CountWorkers:       opts.WorkerCount,
			StateDiskSizeGB:    opts.Config.StateDiskSizeGB,
		},
		Cloud:                   opts.Config.Provider.OpenStack.Cloud,
		AvailabilityZone:        opts.Config.Provider.OpenStack.AvailabilityZone,
		FlavorID:                opts.Config.Provider.OpenStack.FlavorID,
		FloatingIPPoolID:        opts.Config.Provider.OpenStack.FloatingIPPoolID,
		StateDiskType:           opts.Config.Provider.OpenStack.StateDiskType,
		ImageURL:                opts.image,
		DirectDownload:          *opts.Config.Provider.OpenStack.DirectDownload,
		OpenstackUserDomainName: opts.Config.Provider.OpenStack.UserDomainName,
		OpenstackUsername:       opts.Config.Provider.OpenStack.Username,
		OpenstackPassword:       opts.Config.Provider.OpenStack.Password,
		Debug:                   opts.Config.IsDebugCluster(),
	}

	if err := cl.PrepareWorkspace(path.Join("terraform", strings.ToLower(cloudprovider.OpenStack.String())), &vars); err != nil {
		return clusterid.File{}, err
	}

	defer rollbackOnError(c.out, &retErr, &rollbackerTerraform{client: cl}, opts.TFLogLevel)
	tfOutput, err := cl.CreateCluster(ctx, opts.TFLogLevel)
	if err != nil {
		return clusterid.File{}, err
	}

	return clusterid.File{
		CloudProvider: cloudprovider.OpenStack,
		IP:            tfOutput.IP,
		InitSecret:    []byte(tfOutput.Secret),
		UID:           tfOutput.UID,
	}, nil
}

type qemuCreateOptions struct {
	source string
	CreateOptions
}

func (c *Creator) createQEMU(ctx context.Context, cl terraformClient, lv libvirtRunner, opts qemuCreateOptions) (idFile clusterid.File, retErr error) {
	qemuRollbacker := &rollbackerQEMU{client: cl, libvirt: lv, createdWorkspace: false}
	defer rollbackOnError(c.out, &retErr, qemuRollbacker, opts.TFLogLevel)

	// TODO: render progress bar
	downloader := c.newRawDownloader()
	imagePath, err := downloader.Download(ctx, c.out, false, opts.source, opts.Config.Image)
	if err != nil {
		return clusterid.File{}, fmt.Errorf("download raw image: %w", err)
	}

	libvirtURI := opts.Config.Provider.QEMU.LibvirtURI
	libvirtSocketPath := "."

	switch {
	// if no libvirt URI is specified, start a libvirt container
	case libvirtURI == "":
		if err := lv.Start(ctx, opts.Config.Name, opts.Config.Provider.QEMU.LibvirtContainerImage); err != nil {
			return clusterid.File{}, err
		}
		libvirtURI = libvirt.LibvirtTCPConnectURI

	// socket for system URI should be in /var/run/libvirt/libvirt-sock
	case libvirtURI == "qemu:///system":
		libvirtSocketPath = "/var/run/libvirt/libvirt-sock"

	// socket for session URI should be in /run/user/<uid>/libvirt/libvirt-sock
	case libvirtURI == "qemu:///session":
		libvirtSocketPath = fmt.Sprintf("/run/user/%d/libvirt/libvirt-sock", os.Getuid())

	// if a unix socket is specified we need to parse the URI to get the socket path
	case strings.HasPrefix(libvirtURI, "qemu+unix://"):
		unixURI, err := url.Parse(strings.TrimPrefix(libvirtURI, "qemu+unix://"))
		if err != nil {
			return clusterid.File{}, err
		}
		libvirtSocketPath = unixURI.Query().Get("socket")
		if libvirtSocketPath == "" {
			return clusterid.File{}, fmt.Errorf("socket path not specified in qemu+unix URI: %s", libvirtURI)
		}
	}

	metadataLibvirtURI := libvirtURI
	if libvirtSocketPath != "." {
		metadataLibvirtURI = "qemu:///system"
	}

	vars := terraform.QEMUVariables{
		CommonVariables: terraform.CommonVariables{
			Name:               opts.Config.Name,
			CountControlPlanes: opts.ControlPlaneCount,
			CountWorkers:       opts.WorkerCount,
			StateDiskSizeGB:    opts.Config.StateDiskSizeGB,
		},
		LibvirtURI:         libvirtURI,
		LibvirtSocketPath:  libvirtSocketPath,
		ImagePath:          imagePath,
		ImageFormat:        opts.Config.Provider.QEMU.ImageFormat,
		CPUCount:           opts.Config.Provider.QEMU.VCPUs,
		MemorySizeMiB:      opts.Config.Provider.QEMU.Memory,
		MetadataAPIImage:   opts.Config.Provider.QEMU.MetadataAPIImage,
		MetadataLibvirtURI: metadataLibvirtURI,
		NVRAM:              opts.Config.Provider.QEMU.NVRAM,
		Firmware:           opts.Config.Provider.QEMU.Firmware,
	}

	if err := cl.PrepareWorkspace(path.Join("terraform", strings.ToLower(cloudprovider.QEMU.String())), &vars); err != nil {
		return clusterid.File{}, err
	}

	// Allow rollback of QEMU Terraform workspace from this point on
	qemuRollbacker.createdWorkspace = true

	tfOutput, err := cl.CreateCluster(ctx, opts.TFLogLevel)
	if err != nil {
		return clusterid.File{}, err
	}

	return clusterid.File{
		CloudProvider: cloudprovider.QEMU,
		InitSecret:    []byte(tfOutput.Secret),
		IP:            tfOutput.IP,
		UID:           tfOutput.UID,
	}, nil
}

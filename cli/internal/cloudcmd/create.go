/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"

	"github.com/edgelesssys/constellation/v2/cli/internal/clusterid"
	"github.com/edgelesssys/constellation/v2/cli/internal/libvirt"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/imagefetcher"
)

// Creator creates cloud resources.
type Creator struct {
	out                io.Writer
	image              imageFetcher
	newTerraformClient func(ctx context.Context) (terraformClient, error)
	newLibvirtRunner   func() libvirtRunner
	newRawDownloader   func() rawDownloader
	policyPatcher      policyPatcher
}

// NewCreator creates a new creator.
func NewCreator(out io.Writer) *Creator {
	return &Creator{
		out:   out,
		image: imagefetcher.New(),
		newTerraformClient: func(ctx context.Context) (terraformClient, error) {
			return terraform.New(ctx, constants.TerraformWorkingDir)
		},
		newLibvirtRunner: func() libvirtRunner {
			return libvirt.New()
		},
		newRawDownloader: func() rawDownloader {
			return imagefetcher.NewDownloader()
		},
		policyPatcher: NewAzurePolicyPatcher(),
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
	provider := opts.Config.GetProvider()
	attestationVariant := opts.Config.GetAttestationConfig().GetVariant()
	region := opts.Config.GetRegion()
	image, err := c.image.FetchReference(ctx, provider, attestationVariant, opts.Config.Image, region)
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

		vars := awsTerraformVars(opts.Config, opts.image, &opts.ControlPlaneCount, &opts.WorkerCount)

		return c.createAWS(ctx, cl, vars, opts.TFLogLevel)
	case cloudprovider.GCP:
		cl, err := c.newTerraformClient(ctx)
		if err != nil {
			return clusterid.File{}, err
		}
		defer cl.RemoveInstaller()

		vars := gcpTerraformVars(opts.Config, opts.image, &opts.ControlPlaneCount, &opts.WorkerCount)

		return c.createGCP(ctx, cl, vars, opts.TFLogLevel)
	case cloudprovider.Azure:
		cl, err := c.newTerraformClient(ctx)
		if err != nil {
			return clusterid.File{}, err
		}
		defer cl.RemoveInstaller()

		vars := azureTerraformVars(opts.Config, opts.image, &opts.ControlPlaneCount, &opts.WorkerCount)

		return c.createAzure(ctx, cl, vars, opts.TFLogLevel)
	case cloudprovider.OpenStack:
		cl, err := c.newTerraformClient(ctx)
		if err != nil {
			return clusterid.File{}, err
		}
		defer cl.RemoveInstaller()

		vars := openStackTerraformVars(opts.Config, opts.image, &opts.ControlPlaneCount, &opts.WorkerCount)

		return c.createOpenStack(ctx, cl, vars, opts)
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

func (c *Creator) createAWS(ctx context.Context, cl terraformClient, vars terraform.ClusterVariables, loglevel terraform.LogLevel) (idFile clusterid.File, retErr error) {
	tfOutput, err := runTerraformCreate(ctx, cl, cloudprovider.AWS, vars, c.out, loglevel)
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

func (c *Creator) createGCP(ctx context.Context, cl terraformClient, vars terraform.ClusterVariables, loglevel terraform.LogLevel) (idFile clusterid.File, retErr error) {
	tfOutput, err := runTerraformCreate(ctx, cl, cloudprovider.GCP, vars, c.out, loglevel)
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

func (c *Creator) createAzure(ctx context.Context, cl terraformClient, vars terraform.ClusterVariables, loglevel terraform.LogLevel) (idFile clusterid.File, retErr error) {
	tfOutput, err := runTerraformCreate(ctx, cl, cloudprovider.Azure, vars, c.out, loglevel)
	if err != nil {
		return clusterid.File{}, err
	}

	if vars.GetCreateMAA() != nil && *vars.GetCreateMAA() {
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

// policyPatcher interacts with the CSP (currently only applies for Azure) to update the attestation policy.
type policyPatcher interface {
	Patch(ctx context.Context, attestationURL string) error
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

func normalizeAzureURIs(vars *terraform.AzureClusterVariables) *terraform.AzureClusterVariables {
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

func (c *Creator) createOpenStack(ctx context.Context, cl terraformClient, vars terraform.Variables, opts CreateOptions) (idFile clusterid.File, retErr error) {
	// TODO(malt3): Remove this once OpenStack is supported.
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

	tfOutput, err := runTerraformCreate(ctx, cl, cloudprovider.OpenStack, vars, c.out, opts.TFLogLevel)
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

func runTerraformCreate(ctx context.Context, cl terraformClient, provider cloudprovider.Provider, vars terraform.Variables, outWriter io.Writer, loglevel terraform.LogLevel) (output terraform.CreateOutput, retErr error) {
	if err := cl.PrepareWorkspace(path.Join("terraform", strings.ToLower(provider.String())), vars); err != nil {
		return terraform.CreateOutput{}, err
	}

	defer rollbackOnError(outWriter, &retErr, &rollbackerTerraform{client: cl}, loglevel)
	tfOutput, err := cl.CreateCluster(ctx, loglevel)
	if err != nil {
		return terraform.CreateOutput{}, err
	}

	return tfOutput, nil
}

type qemuCreateOptions struct {
	source string
	CreateOptions
}

func (c *Creator) createQEMU(ctx context.Context, cl terraformClient, lv libvirtRunner, opts qemuCreateOptions) (idFile clusterid.File, retErr error) {
	qemuRollbacker := &rollbackerQEMU{client: cl, libvirt: lv, createdWorkspace: false}
	defer rollbackOnError(c.out, &retErr, qemuRollbacker, opts.TFLogLevel)

	// TODO(malt3): render progress bar
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
			return clusterid.File{}, fmt.Errorf("start libvirt container: %w", err)
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

	vars := qemuTerraformVars(opts.Config, imagePath, &opts.ControlPlaneCount, &opts.WorkerCount, libvirtURI, libvirtSocketPath, metadataLibvirtURI)

	if opts.Config.Provider.QEMU.Firmware != "" {
		vars.Firmware = toPtr(opts.Config.Provider.QEMU.Firmware)
	}

	if err := cl.PrepareWorkspace(path.Join("terraform", strings.ToLower(cloudprovider.QEMU.String())), vars); err != nil {
		return clusterid.File{}, fmt.Errorf("prepare workspace: %w", err)
	}

	// Allow rollback of QEMU Terraform workspace from this point on
	qemuRollbacker.createdWorkspace = true

	tfOutput, err := cl.CreateCluster(ctx, opts.TFLogLevel)
	if err != nil {
		return clusterid.File{}, fmt.Errorf("create cluster: %w", err)
	}

	return clusterid.File{
		CloudProvider: cloudprovider.QEMU,
		InitSecret:    []byte(tfOutput.Secret),
		IP:            tfOutput.IP,
		UID:           tfOutput.UID,
	}, nil
}

func toPtr[T any](v T) *T {
	return &v
}

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package terraform handles creation/destruction of a Constellation cluster using Terraform.

Since Terraform does not provide a stable Go API, we use the `terraform-exec` package to interact with Terraform.

The Terraform templates are located in the "terraform" subdirectory. The templates are embedded into the CLI binary using `go:embed`.
On use the relevant template is extracted to the working directory and the user customized variables are written to a `terraform.tfvars` file.
*/
package terraform

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/hashicorp/go-version"
	install "github.com/hashicorp/hc-install"
	"github.com/hashicorp/hc-install/fs"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/hc-install/src"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/spf13/afero"
)

const (
	tfVersion         = ">= 1.4.6"
	terraformVarsFile = "terraform.tfvars"
)

// ErrTerraformWorkspaceExistsWithDifferentVariables is returned when existing Terraform files differ from the version the CLI wants to extract.
var ErrTerraformWorkspaceExistsWithDifferentVariables = errors.New("creating cluster: a Terraform workspace already exists with different variables")

// Client manages interaction with Terraform.
type Client struct {
	tf tfInterface

	manualStateMigrations []StateMigration
	file                  file.Handler
	workingDir            string
	remove                func()
}

// New sets up a new Client for Terraform.
func New(ctx context.Context, workingDir string) (*Client, error) {
	file := file.NewHandler(afero.NewOsFs())
	if err := file.MkdirAll(workingDir); err != nil {
		return nil, err
	}
	tf, remove, err := GetExecutable(ctx, workingDir)
	if err != nil {
		return nil, err
	}

	return &Client{
		tf:         tf,
		remove:     remove,
		file:       file,
		workingDir: workingDir,
	}, nil
}

// WithManualStateMigration adds a manual state migration to the Client.
func (c *Client) WithManualStateMigration(migration StateMigration) *Client {
	c.manualStateMigrations = append(c.manualStateMigrations, migration)
	return c
}

// ShowIAM reads the state of Constellation IAM resources from Terraform.
func (c *Client) ShowIAM(ctx context.Context, provider cloudprovider.Provider) (IAMOutput, error) {
	tfState, err := c.tf.Show(ctx)
	if err != nil {
		return IAMOutput{}, err
	}

	switch provider {
	case cloudprovider.GCP:
		saKeyOutputRaw, ok := tfState.Values.Outputs["sa_key"]
		if !ok {
			return IAMOutput{}, errors.New("no service account key output found")
		}
		saKeyOutput, ok := saKeyOutputRaw.Value.(string)
		if !ok {
			return IAMOutput{}, errors.New("invalid type in service account key output: not a string")
		}
		return IAMOutput{
			GCP: GCPIAMOutput{
				SaKey: saKeyOutput,
			},
		}, nil
	case cloudprovider.Azure:
		subscriptionIDRaw, ok := tfState.Values.Outputs["subscription_id"]
		if !ok {
			return IAMOutput{}, errors.New("no subscription id output found")
		}
		subscriptionIDOutput, ok := subscriptionIDRaw.Value.(string)
		if !ok {
			return IAMOutput{}, errors.New("invalid type in subscription id output: not a string")
		}
		tenantIDRaw, ok := tfState.Values.Outputs["tenant_id"]
		if !ok {
			return IAMOutput{}, errors.New("no tenant id output found")
		}
		tenantIDOutput, ok := tenantIDRaw.Value.(string)
		if !ok {
			return IAMOutput{}, errors.New("invalid type in tenant id output: not a string")
		}
		uamiIDRaw, ok := tfState.Values.Outputs["uami_id"]
		if !ok {
			return IAMOutput{}, errors.New("no UAMI id output found")
		}
		uamiIDOutput, ok := uamiIDRaw.Value.(string)
		if !ok {
			return IAMOutput{}, errors.New("invalid type in UAMI id output: not a string")
		}
		return IAMOutput{
			Azure: AzureIAMOutput{
				SubscriptionID: subscriptionIDOutput,
				TenantID:       tenantIDOutput,
				UAMIID:         uamiIDOutput,
			},
		}, nil
	case cloudprovider.AWS:
		controlPlaneProfileRaw, ok := tfState.Values.Outputs["control_plane_instance_profile"]
		if !ok {
			return IAMOutput{}, errors.New("no control plane instance profile output found")
		}
		controlPlaneProfileOutput, ok := controlPlaneProfileRaw.Value.(string)
		if !ok {
			return IAMOutput{}, errors.New("invalid type in control plane instance profile output: not a string")
		}
		workerNodeProfileRaw, ok := tfState.Values.Outputs["worker_nodes_instance_profile"]
		if !ok {
			return IAMOutput{}, errors.New("no worker node instance profile output found")
		}
		workerNodeProfileOutput, ok := workerNodeProfileRaw.Value.(string)
		if !ok {
			return IAMOutput{}, errors.New("invalid type in worker node instance profile output: not a string")
		}
		return IAMOutput{
			AWS: AWSIAMOutput{
				ControlPlaneInstanceProfile: controlPlaneProfileOutput,
				WorkerNodeInstanceProfile:   workerNodeProfileOutput,
			},
		}, nil
	default:
		return IAMOutput{}, errors.New("unsupported cloud provider")
	}
}

// ShowCluster reads the state of Constellation cluster resources from Terraform.
func (c *Client) ShowCluster(ctx context.Context, provider cloudprovider.Provider) (ApplyOutput, error) {
	tfState, err := c.tf.Show(ctx)
	if err != nil {
		return ApplyOutput{}, fmt.Errorf("terraform show: %w", err)
	}

	ipOutput, ok := tfState.Values.Outputs["ip"]
	if !ok {
		return ApplyOutput{}, errors.New("no IP output found")
	}
	ip, ok := ipOutput.Value.(string)
	if !ok {
		return ApplyOutput{}, errors.New("invalid type in IP output: not a string")
	}

	apiServerCertSANsOutput, ok := tfState.Values.Outputs["api_server_cert_sans"]
	if !ok {
		return ApplyOutput{}, errors.New("no api_server_cert_sans output found")
	}
	apiServerCertSANsUntyped, ok := apiServerCertSANsOutput.Value.([]any)
	if !ok {
		return ApplyOutput{}, fmt.Errorf("invalid type in api_server_cert_sans output: %s is not a list of elements", apiServerCertSANsOutput.Type.FriendlyName())
	}
	apiServerCertSANs, err := toStringSlice(apiServerCertSANsUntyped)
	if err != nil {
		return ApplyOutput{}, fmt.Errorf("convert api_server_cert_sans output: %w", err)
	}

	secretOutput, ok := tfState.Values.Outputs["initSecret"]
	if !ok {
		return ApplyOutput{}, errors.New("no initSecret output found")
	}
	secret, ok := secretOutput.Value.(string)
	if !ok {
		return ApplyOutput{}, errors.New("invalid type in initSecret output: not a string")
	}

	uidOutput, ok := tfState.Values.Outputs["uid"]
	if !ok {
		return ApplyOutput{}, errors.New("no uid output found")
	}
	uid, ok := uidOutput.Value.(string)
	if !ok {
		return ApplyOutput{}, errors.New("invalid type in uid output: not a string")
	}

	res := ApplyOutput{
		IP:                ip,
		APIServerCertSANs: apiServerCertSANs,
		Secret:            secret,
		UID:               uid,
	}
	// TODO add provider
	switch provider {
	case cloudprovider.GCP:
		gcpProjectOutput, ok := tfState.Values.Outputs["project"]
		if ok {
			gcpProject, ok := gcpProjectOutput.Value.(string)
			if !ok {
				return ApplyOutput{}, errors.New("invalid type in project output: not a string")
			}
			cidrNodesOutput, ok := tfState.Values.Outputs["ip_cidr_nodes"]
			if !ok {
				return ApplyOutput{}, errors.New("no ip_cidr_nodes output found")
			}
			cidrNodes, ok := cidrNodesOutput.Value.(string)
			if !ok {
				return ApplyOutput{}, errors.New("invalid type in ip_cidr_nodes output: not a string")
			}
			cidrPodsOutput, ok := tfState.Values.Outputs["ip_cidr_pods"]
			if !ok {
				return ApplyOutput{}, errors.New("no ip_cidr_pods output found")
			}
			cidrPods, ok := cidrPodsOutput.Value.(string)
			if !ok {
				return ApplyOutput{}, errors.New("invalid type in ip_cidr_pods output: not a string")
			}
			res.GCP = &GCPApplyOutput{
				ProjectID:  gcpProject,
				IPCidrNode: cidrNodes,
				IPCidrPod:  cidrPods,
			}
		}
	case cloudprovider.Azure:
		var attestationURL string
		if ok {
			if attestationURLOutput, ok := tfState.Values.Outputs["attestationURL"]; ok {
				if attestationURLString, ok := attestationURLOutput.Value.(string); ok {
					attestationURL = attestationURLString
				}
			}
		}

		azureUAMIOutput, ok := tfState.Values.Outputs["user_assigned_identity"]
		if !ok {
			return ApplyOutput{}, errors.New("no user_assigned_identity output found")
		}
		azureUAMI, ok := azureUAMIOutput.Value.(string)
		if !ok {
			return ApplyOutput{}, errors.New("invalid type in user_assigned_identity output: not a string")
		}

		rgOutput, ok := tfState.Values.Outputs["resource_group"]
		if !ok {
			return ApplyOutput{}, errors.New("no resource_group output found")
		}
		rg, ok := rgOutput.Value.(string)
		if !ok {
			return ApplyOutput{}, errors.New("invalid type in resource_group output: not a string")
		}

		subscriptionOutput, ok := tfState.Values.Outputs["subscription_id"]
		if !ok {
			return ApplyOutput{}, errors.New("no subscription_id output found")
		}
		subscriptionID, ok := subscriptionOutput.Value.(string)
		if !ok {
			return ApplyOutput{}, errors.New("invalid type in subscription_id output: not a string")
		}

		networkSGNameOutput, ok := tfState.Values.Outputs["network_security_group_name"]
		if !ok {
			return ApplyOutput{}, errors.New("no network_security_group_name output found")
		}
		networkSGName, ok := networkSGNameOutput.Value.(string)
		if !ok {
			return ApplyOutput{}, errors.New("invalid type in network_security_group_name output: not a string")
		}
		loadBalancerNameOutput, ok := tfState.Values.Outputs["loadbalancer_name"]
		if !ok {
			return ApplyOutput{}, errors.New("no loadbalancer_name output found")
		}
		loadBalancerName, ok := loadBalancerNameOutput.Value.(string)
		if !ok {
			return ApplyOutput{}, errors.New("invalid type in loadbalancer_name output: not a string")
		}
		res.Azure = &AzureApplyOutput{
			ResourceGroup:            rg,
			SubscriptionID:           subscriptionID,
			UserAssignedIdentity:     azureUAMI,
			NetworkSecurityGroupName: networkSGName,
			LoadBalancerName:         loadBalancerName,
			AttestationURL:           attestationURL,
		}
	}
	return res, nil
}

// PrepareWorkspace prepares a Terraform workspace for a Constellation cluster.
func (c *Client) PrepareWorkspace(path string, vars Variables) error {
	if err := prepareWorkspace(path, c.file, c.workingDir); err != nil {
		return fmt.Errorf("prepare workspace: %w", err)
	}

	return c.writeVars(vars)
}

// PrepareUpgradeWorkspace prepares a Terraform workspace for a Constellation version upgrade.
// It copies the Terraform state from the old working dir and the embedded Terraform files into the new working dir.
func (c *Client) PrepareUpgradeWorkspace(path, oldWorkingDir, newWorkingDir, backupDir string, vars Variables) error {
	if err := prepareUpgradeWorkspace(path, c.file, oldWorkingDir, newWorkingDir, backupDir); err != nil {
		return fmt.Errorf("prepare upgrade workspace: %w", err)
	}

	return c.writeVars(vars)
}

// PrepareIAMUpgradeWorkspace prepares a Terraform workspace for a Constellation IAM upgrade.
func PrepareIAMUpgradeWorkspace(file file.Handler, path, oldWorkingDir, newWorkingDir, backupDir string) error {
	if err := prepareUpgradeWorkspace(path, file, oldWorkingDir, newWorkingDir, backupDir); err != nil {
		return fmt.Errorf("prepare upgrade workspace: %w", err)
	}
	// copy the vars file from the old working dir to the new working dir
	if err := file.CopyFile(filepath.Join(oldWorkingDir, terraformVarsFile), filepath.Join(newWorkingDir, terraformVarsFile)); err != nil {
		return fmt.Errorf("copying vars file: %w", err)
	}
	return nil
}

// CreateCluster creates a Constellation cluster using Terraform.
func (c *Client) CreateCluster(ctx context.Context, provider cloudprovider.Provider, logLevel LogLevel) (ApplyOutput, error) {
	if err := c.setLogLevel(logLevel); err != nil {
		return ApplyOutput{}, fmt.Errorf("set terraform log level %s: %w", logLevel.String(), err)
	}

	if err := c.tf.Init(ctx); err != nil {
		return ApplyOutput{}, fmt.Errorf("terraform init: %w", err)
	}

	if err := c.applyManualStateMigrations(ctx); err != nil {
		return ApplyOutput{}, fmt.Errorf("apply manual state migrations: %w", err)
	}

	if err := c.tf.Apply(ctx); err != nil {
		return ApplyOutput{}, fmt.Errorf("terraform apply: %w", err)
	}

	return c.ShowCluster(ctx, provider)
}

// ApplyOutput contains the Terraform output values of a cluster creation
// or apply operation.
type ApplyOutput struct {
	IP                string
	APIServerCertSANs []string
	Secret            string
	UID               string
	GCP               *GCPApplyOutput
	Azure             *AzureApplyOutput
}

// AzureApplyOutput contains the Terraform output values of a terraform apply operation on Microsoft Azure.
type AzureApplyOutput struct {
	ResourceGroup            string
	SubscriptionID           string
	NetworkSecurityGroupName string
	LoadBalancerName         string
	UserAssignedIdentity     string
	// AttestationURL is the URL of the attestation provider.
	AttestationURL string
}

// GCPApplyOutput contains the Terraform output values of a terraform apply operation on GCP.
type GCPApplyOutput struct {
	ProjectID  string
	IPCidrNode string
	IPCidrPod  string
}

// IAMOutput contains the output information of the Terraform IAM operations.
type IAMOutput struct {
	GCP   GCPIAMOutput
	Azure AzureIAMOutput
	AWS   AWSIAMOutput
}

// GCPIAMOutput contains the output information of the Terraform IAM operation on GCP.
type GCPIAMOutput struct {
	SaKey string
}

// AzureIAMOutput contains the output information of the Terraform IAM operation on Microsoft Azure.
type AzureIAMOutput struct {
	SubscriptionID string
	TenantID       string
	UAMIID         string
}

// AWSIAMOutput contains the output information of the Terraform IAM operation on GCP.
type AWSIAMOutput struct {
	ControlPlaneInstanceProfile string
	WorkerNodeInstanceProfile   string
}

// ApplyIAMConfig creates an IAM configuration using Terraform.
func (c *Client) ApplyIAMConfig(ctx context.Context, provider cloudprovider.Provider, logLevel LogLevel) (IAMOutput, error) {
	if err := c.setLogLevel(logLevel); err != nil {
		return IAMOutput{}, fmt.Errorf("set terraform log level %s: %w", logLevel.String(), err)
	}

	if err := c.tf.Init(ctx); err != nil {
		return IAMOutput{}, err
	}

	if err := c.tf.Apply(ctx); err != nil {
		return IAMOutput{}, err
	}
	return c.ShowIAM(ctx, provider)
}

// Plan determines the diff that will be applied by Terraform. The plan output is written to the planFile.
// If there is a diff, the returned bool is true. Otherwise, it is false.
func (c *Client) Plan(ctx context.Context, logLevel LogLevel, planFile string) (bool, error) {
	if err := c.setLogLevel(logLevel); err != nil {
		return false, fmt.Errorf("set terraform log level %s: %w", logLevel.String(), err)
	}

	if err := c.tf.Init(ctx); err != nil {
		return false, fmt.Errorf("terraform init: %w", err)
	}

	if err := c.applyManualStateMigrations(ctx); err != nil {
		return false, fmt.Errorf("apply manual state migrations: %w", err)
	}

	opts := []tfexec.PlanOption{
		tfexec.Out(planFile),
	}
	return c.tf.Plan(ctx, opts...)
}

// ShowPlan formats the diff in planFilePath and writes it to the specified output.
func (c *Client) ShowPlan(ctx context.Context, logLevel LogLevel, planFilePath string, output io.Writer) error {
	if err := c.setLogLevel(logLevel); err != nil {
		return fmt.Errorf("set terraform log level %s: %w", logLevel.String(), err)
	}

	planResult, err := c.tf.ShowPlanFileRaw(ctx, planFilePath)
	if err != nil {
		return fmt.Errorf("terraform show plan: %w", err)
	}

	_, err = output.Write([]byte(planResult))
	if err != nil {
		return fmt.Errorf("write plan output: %w", err)
	}

	return nil
}

// Destroy destroys Terraform-created cloud resources.
func (c *Client) Destroy(ctx context.Context, logLevel LogLevel) error {
	if err := c.setLogLevel(logLevel); err != nil {
		return fmt.Errorf("set terraform log level %s: %w", logLevel.String(), err)
	}

	if err := c.tf.Init(ctx); err != nil {
		return fmt.Errorf("terraform init: %w", err)
	}
	return c.tf.Destroy(ctx)
}

// RemoveInstaller removes the Terraform installer, if it was downloaded for this command.
func (c *Client) RemoveInstaller() {
	c.remove()
}

// CleanUpWorkspace removes terraform files from the current directory.
func (c *Client) CleanUpWorkspace() error {
	return cleanUpWorkspace(c.file, c.workingDir)
}

// GetExecutable returns a Terraform executable either from the local filesystem,
// or downloads the latest version fulfilling the version constraint.
func GetExecutable(ctx context.Context, workingDir string) (terraform *tfexec.Terraform, remove func(), err error) {
	inst := install.NewInstaller()

	version, err := version.NewConstraint(tfVersion)
	if err != nil {
		return nil, nil, err
	}

	downloadVersion := &releases.LatestVersion{
		Product:     product.Terraform,
		Constraints: version,
	}
	localVersion := &fs.Version{
		Product:     product.Terraform,
		Constraints: version,
	}

	execPath, err := inst.Ensure(ctx, []src.Source{localVersion, downloadVersion})
	if err != nil {
		return nil, nil, err
	}

	tf, err := tfexec.NewTerraform(workingDir, execPath)

	return tf, func() { _ = inst.Remove(context.Background()) }, err
}

// applyManualStateMigrations applies manual state migrations that are not handled by Terraform due to missing features.
// Each migration is expected to be idempotent.
// This is a temporary solution until we can remove the need for manual state migrations.
func (c *Client) applyManualStateMigrations(ctx context.Context) error {
	for _, migration := range c.manualStateMigrations {
		if err := migration.Hook(ctx, c.tf); err != nil {
			return fmt.Errorf("apply manual state migration %s: %w", migration.DisplayName, err)
		}
	}
	return nil
	// expects to be run on initialized workspace
	// and only works for AWS
	// if migration fails, we expect to either be on a different CSP or that the migration has already been applied
	// and we can continue
	// c.tf.StateMv(ctx, "module.control_plane.aws_iam_role.this", "module.control_plane.aws_iam_role.control_plane")
}

// writeVars tries to write the Terraform variables file or, if it exists, checks if it is the same as we are expecting.
func (c *Client) writeVars(vars Variables) error {
	if vars == nil {
		return errors.New("creating cluster: vars is nil")
	}

	pathToVarsFile := filepath.Join(c.workingDir, terraformVarsFile)
	if err := c.file.Write(pathToVarsFile, []byte(vars.String())); errors.Is(err, afero.ErrFileExists) {
		// If a variables file already exists, check if it's the same as we're expecting, so we can continue using it.
		varsContent, err := c.file.Read(pathToVarsFile)
		if err != nil {
			return fmt.Errorf("read variables file: %w", err)
		}
		if vars.String() != string(varsContent) {
			return ErrTerraformWorkspaceExistsWithDifferentVariables
		}
	} else if err != nil {
		return fmt.Errorf("write variables file: %w", err)
	}

	return nil
}

// setLogLevel sets the log level for Terraform.
func (c *Client) setLogLevel(logLevel LogLevel) error {
	if logLevel.String() != "" {
		if err := c.tf.SetLog(logLevel.String()); err != nil {
			return fmt.Errorf("set log level %s: %w", logLevel.String(), err)
		}
		if err := c.tf.SetLogPath(filepath.Join("..", constants.TerraformLogFile)); err != nil {
			return fmt.Errorf("set log path: %w", err)
		}
	}
	return nil
}

// StateMigration is a manual state migration that is not handled by Terraform due to missing features.
// TODO(AB#3248): Remove this after we can assume that all existing clusters have been migrated.
type StateMigration struct {
	DisplayName string
	Hook        func(ctx context.Context, tfClient TFMigrator) error
}

func toStringSlice(in []any) ([]string, error) {
	out := make([]string, len(in))
	for i, v := range in {
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("invalid type in list: item at index %v of list is not a string", i)
		}
		out[i] = s
	}
	return out, nil
}

type tfInterface interface {
	Apply(context.Context, ...tfexec.ApplyOption) error
	Destroy(context.Context, ...tfexec.DestroyOption) error
	Init(context.Context, ...tfexec.InitOption) error
	Show(context.Context, ...tfexec.ShowOption) (*tfjson.State, error)
	Plan(ctx context.Context, opts ...tfexec.PlanOption) (bool, error)
	ShowPlanFileRaw(ctx context.Context, planPath string, opts ...tfexec.ShowOption) (string, error)
	SetLog(level string) error
	SetLogPath(path string) error
	TFMigrator
}

// TFMigrator is an interface for manual terraform state migrations (terraform state mv).
// TODO(AB#3248): Remove this after we can assume that all existing clusters have been migrated.
type TFMigrator interface {
	StateMv(ctx context.Context, src, dst string, opts ...tfexec.StateMvCmdOption) error
}

package cmd

import (
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/edgelesssys/constellation/bootstrapper/util"
	"github.com/edgelesssys/constellation/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/cli/internal/proto"
	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/state"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var diskUUIDRegexp = regexp.MustCompile("^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$")

// NewRecoverCmd returns a new cobra.Command for the recover command.
func NewRecoverCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recover",
		Short: "Recover a completely stopped Constellation cluster",
		Long: "Recover a Constellation cluster by sending a recovery key to an instance in the boot stage." +
			"\nThis is only required if instances restart without other instances available for bootstrapping.",
		Args: cobra.ExactArgs(0),
		RunE: runRecover,
	}
	cmd.Flags().StringP("endpoint", "e", "", "endpoint of the instance, passed as HOST[:PORT] (required)")
	must(cmd.MarkFlagRequired("endpoint"))
	cmd.Flags().String("disk-uuid", "", "disk UUID of the encrypted state disk (required)")
	must(cmd.MarkFlagRequired("disk-uuid"))
	cmd.Flags().String("master-secret", constants.MasterSecretFilename, "path to base64-encoded master secret")
	return cmd
}

func runRecover(cmd *cobra.Command, args []string) error {
	fileHandler := file.NewHandler(afero.NewOsFs())
	recoveryClient := &proto.KeyClient{}
	defer recoveryClient.Close()
	return recover(cmd, fileHandler, recoveryClient)
}

func recover(cmd *cobra.Command, fileHandler file.Handler, recoveryClient recoveryClient) error {
	flags, err := parseRecoverFlags(cmd, fileHandler)
	if err != nil {
		return err
	}

	var stat state.ConstellationState
	if err := fileHandler.ReadJSON(constants.StateFilename, &stat); err != nil {
		return err
	}

	provider := cloudprovider.FromString(stat.CloudProvider)

	config, err := readConfig(cmd.OutOrStdout(), fileHandler, flags.configPath, provider)
	if err != nil {
		return fmt.Errorf("reading and validating config: %w", err)
	}

	validators, err := cloudcmd.NewValidators(provider, config)
	if err != nil {
		return err
	}
	cmd.Print(validators.WarningsIncludeInit())

	if err := recoveryClient.Connect(flags.endpoint, validators.V()); err != nil {
		return err
	}

	diskKey, err := deriveStateDiskKey(flags.masterSecret, flags.diskUUID)
	if err != nil {
		return err
	}

	if err := recoveryClient.PushStateDiskKey(cmd.Context(), diskKey); err != nil {
		return err
	}

	cmd.Println("Pushed recovery key.")
	return nil
}

func parseRecoverFlags(cmd *cobra.Command, fileHandler file.Handler) (recoverFlags, error) {
	endpoint, err := cmd.Flags().GetString("endpoint")
	if err != nil {
		return recoverFlags{}, fmt.Errorf("parsing endpoint argument: %w", err)
	}
	endpoint, err = validateEndpoint(endpoint, constants.BootstrapperPort)
	if err != nil {
		return recoverFlags{}, fmt.Errorf("validating endpoint argument: %w", err)
	}

	diskUUID, err := cmd.Flags().GetString("disk-uuid")
	if err != nil {
		return recoverFlags{}, fmt.Errorf("parsing disk-uuid argument: %w", err)
	}
	if match := diskUUIDRegexp.MatchString(diskUUID); !match {
		return recoverFlags{}, errors.New("flag '--disk-uuid' isn't a valid LUKS UUID")
	}
	diskUUID = strings.ToLower(diskUUID)

	masterSecretPath, err := cmd.Flags().GetString("master-secret")
	if err != nil {
		return recoverFlags{}, fmt.Errorf("parsing master-secret path argument: %w", err)
	}
	masterSecret, err := readMasterSecret(fileHandler, masterSecretPath)
	if err != nil {
		return recoverFlags{}, fmt.Errorf("reading the master secret from file %s: %w", masterSecretPath, err)
	}

	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return recoverFlags{}, fmt.Errorf("parsing config path argument: %w", err)
	}

	return recoverFlags{
		endpoint:     endpoint,
		diskUUID:     diskUUID,
		masterSecret: masterSecret,
		configPath:   configPath,
	}, nil
}

type recoverFlags struct {
	endpoint     string
	diskUUID     string
	masterSecret []byte
	configPath   string
}

// readMasterSecret reads a base64 encoded master secret from file.
func readMasterSecret(fileHandler file.Handler, filename string) ([]byte, error) {
	// Try to read the base64 secret from file
	encodedSecret, err := fileHandler.Read(filename)
	if err != nil {
		return nil, err
	}
	decoded, err := base64.StdEncoding.DecodeString(string(encodedSecret))
	if err != nil {
		return nil, err
	}
	return decoded, nil
}

// deriveStateDiskKey derives a state disk key from a master secret and a disk UUID.
func deriveStateDiskKey(masterKey []byte, diskUUID string) ([]byte, error) {
	return util.DeriveKey(masterKey, []byte("Constellation"), []byte("key"+diskUUID), constants.StateDiskKeyLength)
}

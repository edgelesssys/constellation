/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"

	"github.com/edgelesssys/constellation/v2/debugd/internal/debugd"
	"github.com/edgelesssys/constellation/v2/debugd/internal/debugd/logcollector"
	"github.com/edgelesssys/constellation/v2/debugd/internal/filetransfer"
	"github.com/edgelesssys/constellation/v2/debugd/internal/filetransfer/streamer"
	pb "github.com/edgelesssys/constellation/v2/debugd/service"
	attestationconfigfetcher "github.com/edgelesssys/constellation/v2/internal/api/attestationconfig/fetcher"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func newDeployCmd() *cobra.Command {
	deployCmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploys a self-compiled bootstrapper binary on the current constellation",
		Long: `Deploys a self-compiled bootstrapper binary on the current constellation.
	Uses config provided by --config and reads constellation config from its default location.
	If required, you can override the IP addresses that are used for a deployment by specifying "--ips" and a list of IP addresses.
	Specifying --bootstrapper will upload the bootstrapper from the specified path.`,
		RunE:    runDeploy,
		Example: "cdbg deploy\ncdbg deploy --config /path/to/config\ncdbg deploy --bootstrapper /path/to/bootstrapper --ips 192.0.2.1,192.0.2.2,192.0.2.3 --config /path/to/config",
	}
	deployCmd.Flags().StringSlice("ips", nil, "override the ips that the bootstrapper will be uploaded to (defaults to ips from constellation config)")
	deployCmd.Flags().String("bootstrapper", "./bootstrapper", "override the path to the bootstrapper binary uploaded to instances")
	deployCmd.Flags().String("upgrade-agent", "./upgrade-agent", "override the path to the upgrade-agent binary uploaded to instances")
	deployCmd.Flags().StringToString("info", nil, "additional info to be passed to the debugd, in the form --info key1=value1,key2=value2")
	deployCmd.Flags().Int("verbosity", 0, logger.CmdLineVerbosityDescription)
	return deployCmd
}

func runDeploy(cmd *cobra.Command, _ []string) error {
	verbosity, err := cmd.Flags().GetInt("verbosity")
	if err != nil {
		return err
	}
	log := logger.New(logger.PlainLog, logger.VerbosityFromInt(verbosity))
	configName, err := cmd.Flags().GetString("config")
	if err != nil {
		return fmt.Errorf("parsing config path argument: %w", err)
	}
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return fmt.Errorf("getting force flag: %w", err)
	}

	fs := afero.NewOsFs()
	fileHandler := file.NewHandler(fs)
	streamer := streamer.New(fs)
	transfer := filetransfer.New(log, streamer, filetransfer.ShowProgress)
	constellationConfig, err := config.New(fileHandler, configName, attestationconfigfetcher.New(), force)
	var configValidationErr *config.ValidationError
	if errors.As(err, &configValidationErr) {
		cmd.PrintErrln(configValidationErr.LongMessage())
	}
	if err != nil {
		return err
	}

	return deploy(cmd, fileHandler, constellationConfig, transfer, log)
}

func deploy(cmd *cobra.Command, fileHandler file.Handler, constellationConfig *config.Config, transfer fileTransferer, log *logger.Logger) error {
	bootstrapperPath, err := cmd.Flags().GetString("bootstrapper")
	if err != nil {
		return err
	}
	upgradeAgentPath, err := cmd.Flags().GetString("upgrade-agent")
	if err != nil {
		return err
	}

	if constellationConfig.IsReleaseImage() {
		log.Infof("WARNING: Constellation image does not look like a debug image. Are you using a debug image?")
	}

	if !constellationConfig.IsDebugCluster() {
		log.Infof("WARNING: The Constellation config has debugCluster set to false.")
		log.Infof("cdbg will likely not work unless you manually adjust the firewall / load balancing rules.")
		log.Infof("If you create the cluster with a debug image, you should also set debugCluster to true.")
	}

	ips, err := cmd.Flags().GetStringSlice("ips")
	if err != nil {
		return err
	}
	if len(ips) == 0 {
		var idFile clusterIDsFile
		if err := fileHandler.ReadJSON(constants.ClusterIDsFileName, &idFile); err != nil {
			return fmt.Errorf("reading cluster IDs file: %w", err)
		}
		ips = []string{idFile.IP}
	}

	info, err := cmd.Flags().GetStringToString("info")
	if err != nil {
		return err
	}
	if err := checkInfoMap(info); err != nil {
		return err
	}

	files := []filetransfer.FileStat{
		{
			SourcePath:          bootstrapperPath,
			TargetPath:          debugd.BootstrapperDeployFilename,
			Mode:                debugd.BinaryAccessMode,
			OverrideServiceUnit: "constellation-bootstrapper",
		},
		{
			SourcePath:          upgradeAgentPath,
			TargetPath:          debugd.UpgradeAgentDeployFilename,
			Mode:                debugd.BinaryAccessMode,
			OverrideServiceUnit: "constellation-upgrade-agent",
		},
	}

	for _, ip := range ips {
		input := deployOnEndpointInput{
			debugdEndpoint: ip,
			infos:          info,
			files:          files,
			transfer:       transfer,
			log:            log,
		}
		if err := deployOnEndpoint(cmd.Context(), input); err != nil {
			return err
		}
	}

	return nil
}

type deployOnEndpointInput struct {
	debugdEndpoint string
	files          []filetransfer.FileStat
	infos          map[string]string
	transfer       fileTransferer
	log            *logger.Logger
}

// deployOnEndpoint deploys a custom built bootstrapper binary to a debugd endpoint.
func deployOnEndpoint(ctx context.Context, in deployOnEndpointInput) error {
	in.log.Infof("Deploying on %v", in.debugdEndpoint)

	client, closer, err := newDebugdClient(ctx, in.debugdEndpoint)
	if err != nil {
		return fmt.Errorf("creating debugd client: %w", err)
	}
	defer closer.Close()

	if err := setInfo(ctx, in.log, client, in.infos); err != nil {
		return fmt.Errorf("sending info: %w", err)
	}

	if err := uploadFiles(ctx, client, in); err != nil {
		return fmt.Errorf("uploading bootstrapper: %w", err)
	}

	return nil
}

func newDebugdClient(ctx context.Context, ip string) (pb.DebugdClient, io.Closer, error) {
	conn, err := grpc.DialContext(
		ctx,
		net.JoinHostPort(ip, strconv.Itoa(constants.DebugdPort)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("connecting to other instance via gRPC: %w", err)
	}

	return pb.NewDebugdClient(conn), conn, nil
}

func setInfo(ctx context.Context, log *logger.Logger, client pb.DebugdClient, infos map[string]string) error {
	log.Infof("Setting info with length %d", len(infos))

	var infosPb []*pb.Info
	for key, value := range infos {
		infosPb = append(infosPb, &pb.Info{Key: key, Value: value})
	}

	req := &pb.SetInfoRequest{Info: infosPb}

	status, err := client.SetInfo(ctx, req, grpc.WaitForReady(true))
	if err != nil {
		return fmt.Errorf("setting info: %w", err)
	}

	switch status.Status {
	case pb.SetInfoStatus_SET_INFO_SUCCESS:
		log.Infof("Info set")
	case pb.SetInfoStatus_SET_INFO_ALREADY_SET:
		log.Infof("Info already set")
	default:
		log.Warnf("Unknown status %v", status.Status)
	}
	return nil
}

func uploadFiles(ctx context.Context, client pb.DebugdClient, in deployOnEndpointInput) error {
	in.log.Infof("Uploading files")

	stream, err := client.UploadFiles(ctx, grpc.WaitForReady(true))
	if err != nil {
		return fmt.Errorf("starting bootstrapper upload to instance %v: %w", in.debugdEndpoint, err)
	}

	in.transfer.SetFiles(in.files)
	if err := in.transfer.SendFiles(stream); err != nil {
		return fmt.Errorf("sending files to %v: %w", in.debugdEndpoint, err)
	}

	uploadResponse, closeErr := stream.CloseAndRecv()
	if closeErr != nil {
		return fmt.Errorf("closing upload stream after uploading files to %v: %w", in.debugdEndpoint, closeErr)
	}
	switch uploadResponse.Status {
	case pb.UploadFilesStatus_UPLOAD_FILES_SUCCESS:
		in.log.Infof("Upload successful")
	case pb.UploadFilesStatus_UPLOAD_FILES_ALREADY_FINISHED:
		in.log.Infof("Files already uploaded")
	case pb.UploadFilesStatus_UPLOAD_FILES_UPLOAD_FAILED:
		return fmt.Errorf("uploading files to %v failed: %v", in.debugdEndpoint, uploadResponse)
	case pb.UploadFilesStatus_UPLOAD_FILES_ALREADY_STARTED:
		return fmt.Errorf("upload already started on %v", in.debugdEndpoint)
	default:
		return fmt.Errorf("unknown upload status %v", uploadResponse.Status)
	}

	return nil
}

func checkInfoMap(info map[string]string) error {
	logPrefix, logFields := logcollector.InfoFields()
	for k := range info {
		if !strings.HasPrefix(k, logPrefix) {
			continue
		}
		subkey := strings.TrimPrefix(k, logPrefix)

		if _, ok := logFields[subkey]; !ok {
			return fmt.Errorf("invalid subkey %q for info key %q", subkey, fmt.Sprintf("%s.%s", logPrefix, k))
		}
	}

	return nil
}

type fileTransferer interface {
	SendFiles(stream filetransfer.SendFilesStream) error
	SetFiles(files []filetransfer.FileStat)
}

type clusterIDsFile struct {
	ClusterID string
	OwnerID   string
	IP        string
}

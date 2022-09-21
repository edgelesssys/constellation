/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/edgelesssys/constellation/v2/debugd/internal/bootstrapper"
	"github.com/edgelesssys/constellation/v2/debugd/internal/debugd"
	pb "github.com/edgelesssys/constellation/v2/debugd/service"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func newDeployCmd() *cobra.Command {
	deployCmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploys a self-compiled bootstrapper binary and SSH keys on the current constellation",
		Long: `Deploys a self-compiled bootstrapper binary and SSH keys on the current constellation.
	Uses config provided by --config and reads constellation config from its default location.
	If required, you can override the IP addresses that are used for a deployment by specifying "--ips" and a list of IP addresses.
	Specifying --bootstrapper will upload the bootstrapper from the specified path.`,
		RunE:    runDeploy,
		Example: "cdbg deploy\ncdbg deploy --config /path/to/config\ncdbg deploy --bootstrapper /path/to/bootstrapper --ips 192.0.2.1,192.0.2.2,192.0.2.3 --config /path/to/config",
	}
	deployCmd.Flags().StringSlice("ips", nil, "override the ips that the bootstrapper will be uploaded to (defaults to ips from constellation config)")
	deployCmd.Flags().String("bootstrapper", "./bootstrapper", "override the path to the bootstrapper binary uploaded to instances")
	return deployCmd
}

func runDeploy(cmd *cobra.Command, args []string) error {
	configName, err := cmd.Flags().GetString("config")
	if err != nil {
		return fmt.Errorf("parsing config path argument: %w", err)
	}
	fileHandler := file.NewHandler(afero.NewOsFs())
	constellationConfig, err := config.FromFile(fileHandler, configName)
	if err != nil {
		return err
	}

	return deploy(cmd, fileHandler, constellationConfig, bootstrapper.NewFileStreamer(afero.NewOsFs()))
}

func deploy(cmd *cobra.Command, fileHandler file.Handler, constellationConfig *config.Config, reader fileToStreamReader) error {
	bootstrapperPath, err := cmd.Flags().GetString("bootstrapper")
	if err != nil {
		return err
	}

	if !constellationConfig.IsDebugImage() {
		log.Println("WARNING: Constellation image does not look like a debug image. Are you using a debug image?")
	}

	if !constellationConfig.IsDebugCluster() {
		log.Println("WARNING: The Constellation config has debugCluster set to false.")
		log.Println("cdbg will likely not work unless you manually adjust the firewall / load balancing rules.")
		log.Println("If you create the cluster with a debug image, you should also set debugCluster to true.")
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

	for _, ip := range ips {
		input := deployOnEndpointInput{
			debugdEndpoint:   net.JoinHostPort(ip, strconv.Itoa(constants.DebugdPort)),
			bootstrapperPath: bootstrapperPath,
			reader:           reader,
			authorizedKeys:   constellationConfig.SSHUsers,
		}
		if err := deployOnEndpoint(cmd.Context(), input); err != nil {
			return err
		}
	}

	return nil
}

type deployOnEndpointInput struct {
	debugdEndpoint   string
	bootstrapperPath string
	reader           fileToStreamReader
	authorizedKeys   []config.UserKey
}

// deployOnEndpoint deploys SSH public keys and a locally built bootstrapper binary to a debugd endpoint.
func deployOnEndpoint(ctx context.Context, in deployOnEndpointInput) error {
	log.Printf("Deploying on %v\n", in.debugdEndpoint)
	dialCTX, cancel := context.WithTimeout(ctx, debugd.GRPCTimeout)
	defer cancel()
	conn, err := grpc.DialContext(dialCTX, in.debugdEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("connecting to other instance via gRPC: %w", err)
	}
	defer conn.Close()
	client := pb.NewDebugdClient(conn)

	log.Println("Uploading authorized keys")
	pbKeys := []*pb.AuthorizedKey{}
	for _, key := range in.authorizedKeys {
		pbKeys = append(pbKeys, &pb.AuthorizedKey{
			Username: key.Username,
			KeyValue: key.PublicKey,
		})
	}
	authorizedKeysResponse, err := client.UploadAuthorizedKeys(ctx, &pb.UploadAuthorizedKeysRequest{Keys: pbKeys}, grpc.WaitForReady(true))
	if err != nil || authorizedKeysResponse.Status != pb.UploadAuthorizedKeysStatus_UPLOAD_AUTHORIZED_KEYS_SUCCESS {
		return fmt.Errorf("uploading authorized keys to instance %v failed: %v / %w", in.debugdEndpoint, authorizedKeysResponse, err)
	}

	stream, err := client.UploadBootstrapper(ctx)
	if err != nil {
		return fmt.Errorf("starting bootstrapper upload to instance %v: %w", in.debugdEndpoint, err)
	}
	streamErr := in.reader.ReadStream(in.bootstrapperPath, stream, debugd.Chunksize, true)

	uploadResponse, closeErr := stream.CloseAndRecv()
	if closeErr != nil {
		return fmt.Errorf("closing upload stream after uploading bootstrapper to %v: %w", in.debugdEndpoint, closeErr)
	}
	if uploadResponse.Status == pb.UploadBootstrapperStatus_UPLOAD_BOOTSTRAPPER_FILE_EXISTS {
		log.Println("Bootstrapper was already uploaded")
		return nil
	}
	if uploadResponse.Status != pb.UploadBootstrapperStatus_UPLOAD_BOOTSTRAPPER_SUCCESS || streamErr != nil {
		return fmt.Errorf("uploading bootstrapper to instance %v failed: %v / %w", in.debugdEndpoint, uploadResponse, streamErr)
	}
	log.Println("Uploaded bootstrapper")
	return nil
}

type fileToStreamReader interface {
	ReadStream(filename string, stream bootstrapper.WriteChunkStream, chunksize uint, showProgress bool) error
}

type clusterIDsFile struct {
	ClusterID string
	OwnerID   string
	IP        string
}

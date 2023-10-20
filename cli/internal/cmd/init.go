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
	"os"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	clientcodec "k8s.io/client-go/tools/clientcmd/api/latest"
	"sigs.k8s.io/yaml"

	"github.com/edgelesssys/constellation/v2/bootstrapper/initproto"
	"github.com/edgelesssys/constellation/v2/cli/internal/helm"
	"github.com/edgelesssys/constellation/v2/cli/internal/state"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/grpc/grpclog"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
)

// NewInitCmd returns a new cobra.Command for the init command.
func NewInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize the Constellation cluster",
		Long: "Initialize the Constellation cluster.\n\n" +
			"Start your confidential Kubernetes.",
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Define flags for apply backend that are not set by init
			cmd.Flags().Bool("yes", false, "")
			// Don't skip any phases
			// The apply backend should handle init calls correctly
			cmd.Flags().StringSlice("skip-phases", []string{}, "")
			cmd.Flags().Duration("timeout", time.Hour, "")
			return runApply(cmd, args)
		},
		Deprecated: "use 'constellation apply' instead.",
	}
	cmd.Flags().Bool("conformance", false, "enable conformance mode")
	cmd.Flags().Bool("skip-helm-wait", false, "install helm charts without waiting for deployments to be ready")
	cmd.Flags().Bool("merge-kubeconfig", false, "merge Constellation kubeconfig file with default kubeconfig file in $HOME/.kube/config")
	return cmd
}

type initDoer struct {
	dialer        grpcDialer
	endpoint      string
	req           *initproto.InitRequest
	resp          *initproto.InitSuccessResponse
	log           debugLog
	spinner       spinnerInterf
	connectedOnce bool
	fh            file.Handler
}

func (d *initDoer) Do(ctx context.Context) error {
	// connectedOnce is set in handleGRPCStateChanges when a connection was established in one retry attempt.
	// This should cancel any other retry attempts when the connection is lost since the bootstrapper likely won't accept any new attempts anymore.
	if d.connectedOnce {
		return &nonRetriableError{
			logCollectionErr: errors.New("init already connected to the remote server in a previous attempt - resumption is not supported"),
			err:              errors.New("init already connected to the remote server in a previous attempt - resumption is not supported"),
		}
	}

	conn, err := d.dialer.Dial(ctx, d.endpoint)
	if err != nil {
		d.log.Debugf("Dialing init server failed: %s. Retrying...", err)
		return fmt.Errorf("dialing init server: %w", err)
	}
	defer conn.Close()

	var wg sync.WaitGroup
	defer wg.Wait()

	grpcStateLogCtx, grpcStateLogCancel := context.WithCancel(ctx)
	defer grpcStateLogCancel()
	d.handleGRPCStateChanges(grpcStateLogCtx, &wg, conn)

	protoClient := initproto.NewAPIClient(conn)
	d.log.Debugf("Created protoClient")
	resp, err := protoClient.Init(ctx, d.req)
	if err != nil {
		return &nonRetriableError{
			logCollectionErr: errors.New("rpc failed before first response was received - no logs available"),
			err:              fmt.Errorf("init call: %w", err),
		}
	}

	res, err := resp.Recv() // get first response, either success or failure
	if err != nil {
		if e := d.getLogs(resp); e != nil {
			d.log.Debugf("Failed to collect logs: %s", e)
			return &nonRetriableError{
				logCollectionErr: e,
				err:              err,
			}
		}
		return &nonRetriableError{err: err}
	}

	switch res.Kind.(type) {
	case *initproto.InitResponse_InitFailure:
		if e := d.getLogs(resp); e != nil {
			d.log.Debugf("Failed to get logs from cluster: %s", e)
			return &nonRetriableError{
				logCollectionErr: e,
				err:              errors.New(res.GetInitFailure().GetError()),
			}
		}
		return &nonRetriableError{err: errors.New(res.GetInitFailure().GetError())}
	case *initproto.InitResponse_InitSuccess:
		d.resp = res.GetInitSuccess()
	case nil:
		d.log.Debugf("Cluster returned nil response type")
		err = errors.New("empty response from cluster")
		if e := d.getLogs(resp); e != nil {
			d.log.Debugf("Failed to collect logs: %s", e)
			return &nonRetriableError{
				logCollectionErr: e,
				err:              err,
			}
		}
		return &nonRetriableError{err: err}
	default:
		d.log.Debugf("Cluster returned unknown response type")
		err = errors.New("unknown response from cluster")
		if e := d.getLogs(resp); e != nil {
			d.log.Debugf("Failed to collect logs: %s", e)
			return &nonRetriableError{
				logCollectionErr: e,
				err:              err,
			}
		}
		return &nonRetriableError{err: err}
	}

	return nil
}

func (d *initDoer) getLogs(resp initproto.API_InitClient) error {
	d.log.Debugf("Attempting to collect cluster logs")
	for {
		res, err := resp.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		switch res.Kind.(type) {
		case *initproto.InitResponse_InitFailure:
			return errors.New("trying to collect logs: received init failure response, expected log response")
		case *initproto.InitResponse_InitSuccess:
			return errors.New("trying to collect logs: received init success response, expected log response")
		case nil:
			return errors.New("trying to collect logs: received nil response, expected log response")
		}

		log := res.GetLog().GetLog()
		if log == nil {
			return errors.New("received empty logs")
		}

		if err := d.fh.Write(constants.ErrorLog, log, file.OptAppend); err != nil {
			return err
		}
	}
	return nil
}

func (d *initDoer) handleGRPCStateChanges(ctx context.Context, wg *sync.WaitGroup, conn *grpc.ClientConn) {
	grpclog.LogStateChangesUntilReady(ctx, conn, d.log, wg, func() {
		d.connectedOnce = true
		d.spinner.Stop()
		d.spinner.Start("Initializing cluster ", false)
	})
}

func writeRow(wr io.Writer, col1 string, col2 string) {
	fmt.Fprint(wr, col1, "\t", col2, "\n")
}

type configMerger interface {
	mergeConfigs(configPath string, fileHandler file.Handler) error
	kubeconfigEnvVar() string
}

type kubeconfigMerger struct {
	log debugLog
}

func (c *kubeconfigMerger) mergeConfigs(configPath string, fileHandler file.Handler) error {
	constellConfig, err := clientcmd.LoadFromFile(configPath)
	if err != nil {
		return fmt.Errorf("loading admin kubeconfig: %w", err)
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.Precedence = []string{
		clientcmd.RecommendedHomeFile,
		configPath, // our config should overwrite the default config
	}
	c.log.Debugf("Kubeconfig file loading precedence: %v", loadingRules.Precedence)

	// merge the kubeconfigs
	cfg, err := loadingRules.Load()
	if err != nil {
		return fmt.Errorf("loading merged kubeconfig: %w", err)
	}

	// Set the current context to the cluster we just created
	cfg.CurrentContext = constellConfig.CurrentContext
	c.log.Debugf("Set current context to %s", cfg.CurrentContext)

	json, err := runtime.Encode(clientcodec.Codec, cfg)
	if err != nil {
		return fmt.Errorf("encoding merged kubeconfig: %w", err)
	}

	mergedKubeconfig, err := yaml.JSONToYAML(json)
	if err != nil {
		return fmt.Errorf("converting merged kubeconfig to YAML: %w", err)
	}

	if err := fileHandler.Write(clientcmd.RecommendedHomeFile, mergedKubeconfig, file.OptOverwrite); err != nil {
		return fmt.Errorf("writing merged kubeconfig to file: %w", err)
	}
	c.log.Debugf("Merged kubeconfig into default config file: %s", clientcmd.RecommendedHomeFile)
	return nil
}

func (c *kubeconfigMerger) kubeconfigEnvVar() string {
	return os.Getenv(clientcmd.RecommendedConfigPathEnvVar)
}

type grpcDialer interface {
	Dial(ctx context.Context, target string) (*grpc.ClientConn, error)
}

type nonRetriableError struct {
	logCollectionErr error
	err              error
}

// Error returns the error message.
func (e *nonRetriableError) Error() string {
	return e.err.Error()
}

// Unwrap returns the wrapped error.
func (e *nonRetriableError) Unwrap() error {
	return e.err
}

type helmApplier interface {
	PrepareApply(conf *config.Config, stateFile *state.State,
		flags helm.Options, serviceAccURI string, masterSecret uri.MasterSecret) (
		helm.Applier, bool, error)
}

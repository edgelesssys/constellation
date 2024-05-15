/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package constellation

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/edgelesssys/constellation/v2/bootstrapper/initproto"
	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/constellation/state"
	"github.com/edgelesssys/constellation/v2/internal/grpc/grpclog"
	grpcRetry "github.com/edgelesssys/constellation/v2/internal/grpc/retry"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/edgelesssys/constellation/v2/internal/retry"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"google.golang.org/grpc"
	"k8s.io/client-go/tools/clientcmd"
)

// InitPayload contains the configurable data for the init RPC.
type InitPayload struct {
	MasterSecret    uri.MasterSecret
	MeasurementSalt []byte
	K8sVersion      versions.ValidK8sVersion
	ConformanceMode bool
	ServiceCIDR     string
}

// GrpcDialer dials a gRPC server.
type GrpcDialer interface {
	Dial(ctx context.Context, target string) (*grpc.ClientConn, error)
}

// Init performs the init RPC.
func (a *Applier) Init(
	ctx context.Context,
	validator atls.Validator,
	state *state.State,
	clusterLogWriter io.Writer,
	payload InitPayload,
) (
	InitOutput,
	error,
) {
	// Prepare the Request
	req := &initproto.InitRequest{
		KmsUri:               payload.MasterSecret.EncodeToURI(),
		StorageUri:           uri.NoStoreURI,
		MeasurementSalt:      payload.MeasurementSalt,
		KubernetesVersion:    versions.VersionConfigs[payload.K8sVersion].ClusterVersion,
		KubernetesComponents: versions.VersionConfigs[payload.K8sVersion].KubernetesComponents,
		ConformanceMode:      payload.ConformanceMode,
		InitSecret:           state.Infrastructure.InitSecret,
		ClusterName:          state.Infrastructure.Name,
		ApiserverCertSans:    state.Infrastructure.APIServerCertSANs,
		ServiceCidr:          payload.ServiceCIDR,
	}

	doer := &initDoer{
		dialer: a.newDialer(validator),
		endpoint: net.JoinHostPort(
			state.Infrastructure.ClusterEndpoint,
			strconv.Itoa(constants.BootstrapperPort),
		),
		req:              req,
		log:              a.log,
		clusterLogWriter: clusterLogWriter,
		spinner:          a.spinner,
	}

	// Create a wrapper function that allows logging any returned error from the retrier before checking if it's the expected retriable one.
	serviceIsUnavailable := func(err error) bool {
		isServiceUnavailable := grpcRetry.ServiceIsUnavailable(err)
		a.log.Debug(fmt.Sprintf("Encountered error (retriable: %t): %q", isServiceUnavailable, err))
		return isServiceUnavailable
	}

	// Perform the RPC
	a.log.Debug("Initialization call", "endpoint", doer.endpoint)
	a.spinner.Start("Connecting ", false)
	retrier := retry.NewIntervalRetrier(doer, 30*time.Second, serviceIsUnavailable)
	if err := retrier.Do(ctx); err != nil {
		return InitOutput{}, fmt.Errorf("doing init call: %w", err)
	}
	a.spinner.Stop()
	a.log.Debug("Initialization request finished")

	a.log.Debug(fmt.Sprintf("Rewriting cluster server address in kubeconfig to %q", state.Infrastructure.ClusterEndpoint))
	kubeconfig, err := clientcmd.Load(doer.resp.Kubeconfig)
	if err != nil {
		return InitOutput{}, fmt.Errorf("loading kubeconfig: %w", err)
	}
	if len(kubeconfig.Clusters) != 1 {
		return InitOutput{}, fmt.Errorf("expected exactly one cluster in kubeconfig, got %d", len(kubeconfig.Clusters))
	}
	for _, cluster := range kubeconfig.Clusters {
		kubeEndpoint, err := url.Parse(cluster.Server)
		if err != nil {
			return InitOutput{}, fmt.Errorf("parsing kubeconfig server URL: %w", err)
		}
		kubeEndpoint.Host = net.JoinHostPort(state.Infrastructure.ClusterEndpoint, kubeEndpoint.Port())
		cluster.Server = kubeEndpoint.String()
	}

	kubeconfigBytes, err := clientcmd.Write(*kubeconfig)
	if err != nil {
		return InitOutput{}, fmt.Errorf("writing kubeconfig: %w", err)
	}

	return InitOutput{
		ClusterID:  hex.EncodeToString(doer.resp.ClusterId),
		OwnerID:    hex.EncodeToString(doer.resp.OwnerId),
		Kubeconfig: kubeconfigBytes,
	}, nil
}

// InitOutput contains the output of the init RPC.
type InitOutput struct {
	// ClusterID is the hex encoded ID of the cluster.
	ClusterID string
	// OwnerID is the hex encoded ID of the owner of the cluster.
	OwnerID string
	// Kubeconfig is the kubeconfig for the cluster.
	Kubeconfig []byte
}

// the initDoer performs the actual init RPC with retry logic.
type initDoer struct {
	dialer        GrpcDialer
	endpoint      string
	req           *initproto.InitRequest
	log           debugLog
	connectedOnce bool
	spinner       spinnerInterf

	// clusterLogWriter is the writer to which the cluster logs are written.
	clusterLogWriter io.Writer

	// Read-Only-fields:

	// resp is the response returned upon successful initialization.
	resp *initproto.InitSuccessResponse
}

type spinnerInterf interface {
	Start(text string, showDots bool)
	Stop()
	io.Writer
}

// Do performs the init gRPC call.
func (d *initDoer) Do(ctx context.Context) error {
	// connectedOnce is set in handleGRPCStateChanges when a connection was established in one retry attempt.
	// This should cancel any other retry attempts when the connection is lost since the bootstrapper likely won't accept any new attempts anymore.
	if d.connectedOnce {
		return &NonRetriableInitError{
			LogCollectionErr: errors.New("init already connected to the remote server in a previous attempt - resumption is not supported"),
			Err:              errors.New("init already connected to the remote server in a previous attempt - resumption is not supported"),
		}
	}

	conn, err := d.dialer.Dial(ctx, d.endpoint)
	if err != nil {
		d.log.Debug(fmt.Sprintf("Dialing init server failed: %q. Retrying...", err))
		return fmt.Errorf("dialing init server: %w", err)
	}
	defer conn.Close()

	var wg sync.WaitGroup
	defer wg.Wait()

	grpcStateLogCtx, grpcStateLogCancel := context.WithCancel(ctx)
	defer grpcStateLogCancel()
	d.handleGRPCStateChanges(grpcStateLogCtx, &wg, conn)

	protoClient := initproto.NewAPIClient(conn)
	d.log.Debug("Created protoClient")
	resp, err := protoClient.Init(ctx, d.req)
	if err != nil {
		return &NonRetriableInitError{
			LogCollectionErr: errors.New("rpc failed before first response was received - no logs available"),
			Err:              fmt.Errorf("init call: %w", err),
		}
	}

	res, err := resp.Recv() // get first response, either success or failure
	if err != nil {
		if e := d.getLogs(resp); e != nil {
			d.log.Debug(fmt.Sprintf("Failed to collect logs: %q", e))
			return &NonRetriableInitError{
				LogCollectionErr: e,
				Err:              err,
			}
		}
		return &NonRetriableInitError{Err: err}
	}

	switch res.Kind.(type) {
	case *initproto.InitResponse_InitSuccess:
		d.resp = res.GetInitSuccess()
	case *initproto.InitResponse_InitFailure:
		if e := d.getLogs(resp); e != nil {
			d.log.Debug(fmt.Sprintf("Failed to get logs from cluster: %q", e))
			return &NonRetriableInitError{
				LogCollectionErr: e,
				Err:              errors.New(res.GetInitFailure().GetError()),
			}
		}
		return &NonRetriableInitError{Err: errors.New(res.GetInitFailure().GetError())}
	case nil:
		d.log.Debug("Cluster returned nil response type")
		err = errors.New("empty response from cluster")
		if e := d.getLogs(resp); e != nil {
			d.log.Debug(fmt.Sprintf("Failed to collect logs: %q", e))
			return &NonRetriableInitError{
				LogCollectionErr: e,
				Err:              err,
			}
		}
		return &NonRetriableInitError{Err: err}
	default:
		d.log.Debug("Cluster returned unknown response type")
		err = errors.New("unknown response from cluster")
		if e := d.getLogs(resp); e != nil {
			d.log.Debug(fmt.Sprintf("Failed to collect logs: %q", e))
			return &NonRetriableInitError{
				LogCollectionErr: e,
				Err:              err,
			}
		}
		return &NonRetriableInitError{Err: err}
	}
	return nil
}

// getLogs retrieves the cluster logs from the bootstrapper and saves them in the initDoer.
func (d *initDoer) getLogs(resp initproto.API_InitClient) error {
	d.log.Debug("Attempting to collect cluster logs")
	for {
		res, err := resp.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("receiving logs: %w", err)
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
		if _, err := d.clusterLogWriter.Write(log); err != nil {
			return fmt.Errorf("writing logs: %w", err)
		}
	}

	d.log.Debug("Received cluster logs")
	return nil
}

func (d *initDoer) handleGRPCStateChanges(ctx context.Context, wg *sync.WaitGroup, conn *grpc.ClientConn) {
	grpclog.LogStateChangesUntilReady(ctx, conn, d.log, wg, func() {
		d.connectedOnce = true
		d.spinner.Stop()
		d.spinner.Start("Initializing cluster ", false)
	})
}

// NonRetriableInitError is returned when the init RPC fails and the error is not retriable.
type NonRetriableInitError struct {
	LogCollectionErr error
	Err              error
}

// Error returns the error message.
func (e *NonRetriableInitError) Error() string {
	return e.Err.Error()
}

// Unwrap returns the wrapped error.
func (e *NonRetriableInitError) Unwrap() error {
	return e.Err
}

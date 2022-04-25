package core

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"sync"
	"time"

	"github.com/edgelesssys/constellation/cli/file"
	"github.com/edgelesssys/constellation/coordinator/attestation/vtpm"
	"github.com/edgelesssys/constellation/coordinator/config"
	kmsSetup "github.com/edgelesssys/constellation/coordinator/kms"
	"github.com/edgelesssys/constellation/coordinator/nodestate"
	"github.com/edgelesssys/constellation/coordinator/role"
	"github.com/edgelesssys/constellation/coordinator/state"
	"github.com/edgelesssys/constellation/coordinator/store"
	"github.com/edgelesssys/constellation/coordinator/storewrapper"
	"github.com/edgelesssys/constellation/coordinator/util"
	"github.com/edgelesssys/constellation/kms/kms"
	"go.uber.org/zap"
)

var coordinatorVPNIP = netip.AddrFrom4([4]byte{10, 118, 0, 1})

type Core struct {
	state                  state.State
	openTPM                vtpm.TPMOpenFunc
	mut                    sync.Mutex
	store                  store.Store
	vpn                    VPN
	kube                   Cluster
	metadata               ProviderMetadata
	cloudControllerManager CloudControllerManager
	cloudNodeManager       CloudNodeManager
	clusterAutoscaler      ClusterAutoscaler
	encryptedDisk          EncryptedDisk
	kms                    kms.CloudKMS
	zaplogger              *zap.Logger
	persistentStoreFactory PersistentStoreFactory
	lastHeartbeats         map[string]time.Time
	fileHandler            file.Handler
}

// NewCore creates and initializes a new Core object.
func NewCore(vpn VPN, kube Cluster,
	metadata ProviderMetadata, cloudControllerManager CloudControllerManager, cloudNodeManager CloudNodeManager, clusterAutoscaler ClusterAutoscaler,
	encryptedDisk EncryptedDisk, zapLogger *zap.Logger, openTPM vtpm.TPMOpenFunc, persistentStoreFactory PersistentStoreFactory, fileHandler file.Handler,
) (*Core, error) {
	stor := store.NewStdStore()
	c := &Core{
		openTPM:                openTPM,
		store:                  stor,
		vpn:                    vpn,
		kube:                   kube,
		metadata:               metadata,
		cloudNodeManager:       cloudNodeManager,
		cloudControllerManager: cloudControllerManager,
		clusterAutoscaler:      clusterAutoscaler,
		encryptedDisk:          encryptedDisk,
		zaplogger:              zapLogger,
		kms:                    nil, // KMS is set up during init phase
		persistentStoreFactory: persistentStoreFactory,
		lastHeartbeats:         make(map[string]time.Time),
		fileHandler:            fileHandler,
	}
	if err := c.data().IncrementPeersResourceVersion(); err != nil {
		return nil, err
	}

	if err := vpn.Setup(nil); err != nil {
		return nil, err
	}

	return c, nil
}

// GetVPNPubKey returns the peer's VPN public key.
func (c *Core) GetVPNPubKey() ([]byte, error) {
	return c.vpn.GetPublicKey()
}

// GetVPNPubKey returns the peer's VPN public key.
func (c *Core) InitializeStoreIPs() error {
	return c.data().InitializeStoreIPs()
}

// SetVPNIP sets the peer's VPN IP.
func (c *Core) SetVPNIP(ip string) error {
	return c.vpn.SetInterfaceIP(ip)
}

// GetVPNIP returns the cores VPN IP.
func (c *Core) GetVPNIP() (string, error) {
	return c.vpn.GetInterfaceIP()
}

// GetNextNodeIP gets the next free IP-Addr.
func (c *Core) GetNextNodeIP() (string, error) {
	tx, err := c.store.BeginTransaction()
	if err != nil {
		return "", err
	}
	txwrapper := storewrapper.StoreWrapper{Store: tx}
	ip, err := txwrapper.PopNextFreeNodeIP()
	if err != nil {
		return "", err
	}
	return ip.String(), tx.Commit()
}

// GetNextCoordinatorIP gets the next free IP-Addr.
func (c *Core) GetNextCoordinatorIP() (string, error) {
	tx, err := c.store.BeginTransaction()
	if err != nil {
		return "", err
	}
	txwrapper := storewrapper.StoreWrapper{Store: tx}
	ip, err := txwrapper.PopNextFreeCoordinatorIP()
	if err != nil {
		return "", err
	}
	return ip.String(), tx.Commit()
}

// SwitchToPersistentStore creates a new store using the persistentStoreFactory and transfers the initial temporary store into it.
func (c *Core) SwitchToPersistentStore() error {
	newStore, err := c.persistentStoreFactory.New()
	if err != nil {
		c.zaplogger.Error("error creating persistent store")
		return err
	}
	if err := c.store.Transfer(newStore); err != nil {
		c.zaplogger.Error("transfer to persistent store failed")
		return err
	}
	c.store = newStore
	c.zaplogger.Info("transition to persistent store successful")
	return nil
}

// GetIDs returns the ownerID and clusterID.
// Pass a masterSecret to generate new IDs.
// Pass nil to obtain the existing IDs.
func (c *Core) GetIDs(masterSecret []byte) (ownerID, clusterID []byte, err error) {
	if masterSecret == nil {
		clusterID, err = c.data().GetClusterID()
		if err != nil {
			return nil, nil, err
		}
		masterSecret, err = c.data().GetMasterSecret()
		if err != nil {
			return nil, nil, err
		}
	} else {
		clusterID, err = util.GenerateRandomBytes(config.RNGLengthDefault)
		if err != nil {
			return nil, nil, err
		}
		if err := c.data().PutMasterSecret(masterSecret); err != nil {
			return nil, nil, err
		}
	}

	// TODO: Choose a way to salt ownerID
	ownerID, err = deriveOwnerID(masterSecret)
	if err != nil {
		return nil, nil, err
	}
	return ownerID, clusterID, nil
}

// NotifyNodeHeartbeat notifies the core of a received heartbeat from a node.
func (c *Core) NotifyNodeHeartbeat(addr net.Addr) {
	ip := addr.String()
	now := time.Now()
	c.mut.Lock()
	c.lastHeartbeats[ip] = now
	c.mut.Unlock()
}

// Initialize initializes the state machine of the core and handles re-joining the VPN.
// Blocks until the core is ready to be used.
func (c *Core) Initialize() (nodeActivated bool, err error) {
	nodeActivated, err = vtpm.IsNodeInitialized(c.openTPM)
	if err != nil {
		return false, fmt.Errorf("failed to check for previous activation using vTPM: %w", err)
	}
	if !nodeActivated {
		c.zaplogger.Info("Node was never activated. Allowing node to be activated.")
		c.state.Advance(state.AcceptingInit)
		return false, nil
	}
	c.zaplogger.Info("Node was previously activated. Attempting re-join.")
	nodeState, err := nodestate.FromFile(c.fileHandler)
	if err != nil {
		return false, fmt.Errorf("failed to read node state: %w", err)
	}
	var initialState state.State
	switch nodeState.Role {
	case role.Coordinator:
		initialState = state.ActivatingNodes
	case role.Node:
		initialState = state.IsNode
	default:
		return false, fmt.Errorf("invalid node role for initialized node: %v", nodeState.Role)
	}
	// TODO: if node was previously initialized, attempt to re-join wireguard here.
	// Steps to rejoining should include:
	// - retrieve list of coordinators from cloud provider API
	// - attempt to retrieve list of wireguard public keys from any other coordinator while checking for correct PCRs in ATLS
	// - re-establish wireguard connections
	// - call update function successfully at least once
	// - advance state to IsNode or ActivatingNodes respectively
	// - restart update loop
	// This procedure can be retried until it succeeds.
	// The node must be put into the correct state before the update loop is started.
	panic("not implemented")

	//nolint:govet // this code is unreachable as long as the above is unimplemented
	c.state.Advance(initialState)
	return nodeActivated, nil
}

// PersistNodeState persists node state to disk.
func (c *Core) PersistNodeState(role role.Role, ownerID []byte, clusterID []byte) error {
	vpnPrivKey, err := c.vpn.GetPrivateKey()
	if err != nil {
		return fmt.Errorf("failed to retrieve VPN private key: %w", err)
	}
	nodeState := nodestate.NodeState{
		Role:       role,
		VPNPrivKey: vpnPrivKey,
		OwnerID:    ownerID,
		ClusterID:  clusterID,
	}
	return nodeState.ToFile(c.fileHandler)
}

// SetUpKMS sets the Coordinators key management service and key encryption key ID.
// Creates a new key encryption key in the KMS, if requested.
// Otherwise the KEK is assumed to already exist in the KMS.
func (c *Core) SetUpKMS(ctx context.Context, storageURI, kmsURI, kekID string, useExistingKEK bool) error {
	kms, err := kmsSetup.SetUpKMS(ctx, storageURI, kmsURI)
	if err != nil {
		return err
	}
	c.kms = kms

	if useExistingKEK {
		return nil
	}
	// import Constellation master secret as key encryption key
	kek, err := c.data().GetMasterSecret()
	if err != nil {
		return err
	}
	if err := kms.CreateKEK(ctx, kekID, kek); err != nil {
		return err
	}
	if err := c.data().PutKEKID(kekID); err != nil {
		return err
	}
	bundeldedKMSInfo := kmsSetup.KMSInformation{KmsUri: kmsURI, KeyEncryptionKeyID: kekID, StorageUri: storageURI}
	if err := c.data().PutKMSData(bundeldedKMSInfo); err != nil {
		return err
	}
	return nil
}

func (c *Core) GetKMSInfo() (kmsSetup.KMSInformation, error) {
	return c.data().GetKMSData()
}

// GetDataKey derives a key of length from the Constellation's master secret.
func (c *Core) GetDataKey(ctx context.Context, keyID string, length int) ([]byte, error) {
	if c.kms == nil {
		c.zaplogger.Error("trying to request data key before KMS is set up")
		return nil, errors.New("trying to request data key before KMS is set up")
	}

	kekID, err := c.data().GetKEKID()
	if err != nil {
		return nil, err
	}

	return c.kms.GetDEK(ctx, kekID, keyID, length)
}

func (c *Core) data() storewrapper.StoreWrapper {
	return storewrapper.StoreWrapper{Store: c.store}
}

type PersistentStoreFactory interface {
	New() (store.Store, error)
}

// deriveOwnerID uses the Constellation's master secret to derive a unique value tied to that secret.
func deriveOwnerID(masterSecret []byte) ([]byte, error) {
	// TODO: Choose a way to salt the key derivation
	return util.DeriveKey(masterSecret, []byte("Constellation"), []byte("id"), config.RNGLengthDefault)
}

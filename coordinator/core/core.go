package core

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/edgelesssys/constellation/coordinator/attestation/vtpm"
	"github.com/edgelesssys/constellation/coordinator/config"
	kmsSetup "github.com/edgelesssys/constellation/coordinator/kms"
	"github.com/edgelesssys/constellation/coordinator/state"
	"github.com/edgelesssys/constellation/coordinator/store"
	"github.com/edgelesssys/constellation/coordinator/storewrapper"
	"github.com/edgelesssys/constellation/coordinator/util"
	"github.com/edgelesssys/constellation/kms/pkg/kms"
	"go.uber.org/zap"
)

var coordinatorVPNIP = net.IP{10, 118, 0, 1}

type Core struct {
	state                  state.State
	openTPM                vtpm.TPMOpenFunc
	mut                    sync.Mutex
	store                  store.Store
	vpn                    VPN
	kube                   Cluster
	metadata               ProviderMetadata
	cloudControllerManager CloudControllerManager
	clusterAutoscaler      ClusterAutoscaler
	kms                    kms.CloudKMS
	zaplogger              *zap.Logger
	persistentStoreFactory PersistentStoreFactory
	lastHeartbeats         map[string]time.Time
}

// NewCore creates and initializes a new Core object.
func NewCore(vpn VPN, kube Cluster,
	metadata ProviderMetadata, cloudControllerManager CloudControllerManager, clusterAutoscaler ClusterAutoscaler,
	zapLogger *zap.Logger, openTPM vtpm.TPMOpenFunc, persistentStoreFactory PersistentStoreFactory,
) (*Core, error) {
	stor := store.NewStdStore()
	c := &Core{
		openTPM:                openTPM,
		store:                  stor,
		vpn:                    vpn,
		kube:                   kube,
		metadata:               metadata,
		cloudControllerManager: cloudControllerManager,
		clusterAutoscaler:      clusterAutoscaler,
		zaplogger:              zapLogger,
		kms:                    nil, // KMS is set up during init phase
		persistentStoreFactory: persistentStoreFactory,
		lastHeartbeats:         make(map[string]time.Time),
	}

	if err := c.data().PutLastNodeIP(coordinatorVPNIP.To4()); err != nil {
		return nil, err
	}

	if err := c.data().IncrementPeersResourceVersion(); err != nil {
		return nil, err
	}

	privk, err := vpn.Setup(nil)
	if err != nil {
		return nil, err
	}

	pubk, err := vpn.GetPublicKey(privk)
	if err != nil {
		return nil, err
	}

	if err := c.data().PutVPNKey(pubk); err != nil {
		return nil, err
	}

	c.state.Advance(state.AcceptingInit)

	return c, nil
}

// GetVPNPubKey returns the peer's VPN public key.
func (c *Core) GetVPNPubKey() ([]byte, error) {
	return c.data().GetVPNKey()
}

// SetVPNIP sets the peer's VPN IP.
func (c *Core) SetVPNIP(ip string) error {
	return c.vpn.SetInterfaceIP(ip)
}

// GetCoordinatorVPNIP returns the VPN IP designated for the Coordinator.
func (*Core) GetCoordinatorVPNIP() string {
	return coordinatorVPNIP.String()
}

// AddAdmin adds an admin to the VPN.
func (c *Core) AddAdmin(pubKey []byte) (string, error) {
	vpnIP, err := c.GenerateNextIP()
	if err != nil {
		return "", err
	}
	if err := c.vpn.AddPeer(pubKey, "", vpnIP); err != nil {
		return "", err
	}
	return vpnIP, nil
}

// GenerateNextIP gets the next free IP-Addr.
func (c *Core) GenerateNextIP() (string, error) {
	tx, err := c.store.BeginTransaction()
	if err != nil {
		return "", err
	}
	txwrapper := storewrapper.StoreWrapper{Store: tx}
	ip, err := txwrapper.PopNextFreeNodeIP()
	if err != nil {
		return "", err
	}
	return ip, tx.Commit()
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

// SetUpKMS sets the Coordinators key management service and key encryption key ID.
// Creates a new key encryption key in the KMS, if requested.
// Otherwise the KEK is assumed to already exist in the KMS.
func (c *Core) SetUpKMS(ctx context.Context, storageURI, kmsURI, kekID string, useExisting bool) error {
	kms, err := kmsSetup.SetUpKMS(ctx, storageURI, kmsURI)
	if err != nil {
		return err
	}

	if !useExisting {
		// import Constellation master secret as key encryption key
		kek, err := c.data().GetMasterSecret()
		if err != nil {
			return err
		}
		if err := kms.CreateKEK(ctx, kekID, kek); err != nil {
			return err
		}
	}

	if err := c.data().PutKEKID(kekID); err != nil {
		return err
	}

	c.kms = kms
	return nil
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

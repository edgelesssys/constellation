package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"syscall"
	"time"

	"github.com/edgelesssys/constellation/internal/deploy/ssh"
	"github.com/edgelesssys/constellation/internal/deploy/user"
	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/spf13/afero"
	"go.uber.org/zap"

	v1 "k8s.io/api/core/v1"
	v1Options "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	// hostPath holds the path to the host's root file system we chroot into.
	hostPath = "/host"

	// normalHomePath holds the general home directory of a system.
	normalHomePath = "/var/home"

	// evictedHomePath holds the directory to which deleted user directories are moved to.
	evictedHomePath = "/var/evicted"

	// relativePathToSSHKeys holds the path inside a user's directory to the SSH keys.
	// Needs to be in sync with internal/deploy/ssh.go.
	relativePathToSSHKeys = ".ssh/authorized_keys.d/constellation-ssh-keys"

	// timeout is the maximum time to wait for communication with the Kubernetes API server.
	timeout = 60 * time.Second
)

// uidGidPair holds the user owner and group owner of a directory.
type uidGIDPair struct {
	UID uint32
	GID uint32
}

func main() {
	verbosity := flag.Int("v", 0, logger.CmdLineVerbosityDescription)

	flag.Parse()
	log := logger.New(logger.JSONLog, logger.VerbosityFromInt(*verbosity))

	hostname, err := os.Hostname()
	if err != nil {
		log.Warnf("Starting constellation-access-manager as unknown pod")
	} else {
		log.Infof("Starting constellation-access-manager as %q", hostname)
	}

	// Retrieve configMap from Kubernetes API before we chroot into the host filesystem.
	configMap, err := retrieveConfigMap(log)
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to retrieve ConfigMap from Kubernetes API")
	}

	// Chroot into main system
	if err := syscall.Chroot(hostPath); err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to chroot into host filesystem")
	}
	if err := syscall.Chdir("/"); err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to chdir into host filesystem")
	}

	fs := afero.NewOsFs()
	linuxUserManager := user.NewLinuxUserManager(fs)

	if err := run(log, fs, linuxUserManager, configMap); err != nil {
		// So far there is only one error path in this code, and this is getting the user directories... So just make the error specific here for now.
		log.With(zap.Error(err)).Fatalf("Failed to retrieve existing user directories")
	}
}

// loadClientSet loads the Kubernetes API client.
func loadClientSet() (*kubernetes.Clientset, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

// deployKeys creates or evicts users based on the ConfigMap and deploy their SSH keys.
func deployKeys(
	ctx context.Context, log *logger.Logger, configMap *v1.ConfigMap, fs afero.Fs,
	linuxUserManager user.LinuxUserManager, userMap map[string]uidGIDPair, sshAccess *ssh.Access,
) {
	// If no ConfigMap exists or has been emptied, evict all users and exit.
	if configMap == nil || len(configMap.Data) == 0 {
		for username, ownership := range userMap {
			log := log.With(zap.String("username", username))
			if username != "root" {
				evictedUserPath := path.Join(evictedHomePath, username)
				log.With(zap.Uint32("UID", ownership.UID), zap.Uint32("GID", ownership.GID)).
					Infof("Evicting user to %q", evictedUserPath)

				if err := evictUser(username, fs, linuxUserManager); err != nil {
					log.With(zap.Error(err)).Errorf("Did not evict user")
					continue
				}
			} else {
				log.Infof("Removing any old keys for 'root', if existent")
				// Remove root's SSH key specifically instead of evicting the whole directory.
				if err := evictRootKey(fs, linuxUserManager); err != nil && !os.IsNotExist(err) {
					log.With(zap.Error(err)).Errorf("Failed to remove previously existing root key")
					continue
				}
			}
		}

		return
	}

	// First, recreate users that already existed, if they are defined in the configMap.
	// For users which do not exist, we move their user directories to avoid accidental takeovers but also loss of data.
	for username, ownership := range userMap {
		log := log.With(zap.String("username", username))
		if username != "root" {
			if _, ok := configMap.Data[username]; ok {
				log.With(zap.Uint32("UID", ownership.UID), zap.Uint32("GID", ownership.GID)).
					Infof("Recreating user, if not existent")

				if err := linuxUserManager.Creator.CreateUserWithSpecificUIDAndGID(
					ctx, username, int(ownership.UID), int(ownership.GID),
				); err != nil {
					if errors.Is(err, user.ErrUserOrGroupAlreadyExists) {
						log.Infof("User already exists, skipping")
					} else {
						log.With(zap.Error(err)).Errorf("Failed to recreate user")
					}
					continue
				}
			} else {
				evictedUserPath := path.Join(evictedHomePath, username)
				log.With(zap.Uint32("UID", ownership.UID), zap.Uint32("GID", ownership.GID)).
					Infof("Evicting user to %q", evictedUserPath)
				if err := evictUser(username, fs, linuxUserManager); err != nil {
					log.With(zap.Error(err)).Errorf("Did not evict user")
					continue
				}
			}
		} else {
			log.Infof("Removing any old keys for 'root', if existent")
			// Always remove the root key first, even if it is about to be redeployed.
			if err := evictRootKey(fs, linuxUserManager); err != nil && !os.IsNotExist(err) {
				log.With(zap.Error(err)).Errorf("Failed to remove previously existing root key")
				continue
			}
		}
	}

	// Then, create the remaining users from the configMap (if remaining) and deploy SSH keys for all users.
	for username, publicKey := range configMap.Data {
		log := log.With(zap.String("username", username))
		if _, ok := userMap[username]; !ok {
			log.Infof("Creating user")
			if err := linuxUserManager.Creator.CreateUser(ctx, username); err != nil {
				if errors.Is(err, user.ErrUserOrGroupAlreadyExists) {
					log.Infof("User already exists, skipping")
				} else {
					log.With(zap.Error(err)).Errorf("Failed to create user")
				}
				continue
			}
		}

		// If we created a user, let's actually get the home directory instead of assuming it's the same as the normal home directory.
		user, err := linuxUserManager.GetLinuxUser(username)
		if err != nil {
			log.With(zap.Error(err)).Errorf("Failed to retrieve information about user")
			continue
		}

		// Delete already deployed keys
		pathToSSHKeys := filepath.Join(user.Home, relativePathToSSHKeys)
		if err := fs.Remove(pathToSSHKeys); err != nil && !os.IsNotExist(err) {
			log.With(zap.Error(err)).Errorf("Failed to delete remaining managed SSH keys for user")
			continue
		}

		// And (re)deploy the keys from the ConfigMap
		newKey := ssh.UserKey{
			Username:  username,
			PublicKey: publicKey,
		}

		log.Infof("Deploying new SSH key for user")
		if err := sshAccess.DeployAuthorizedKey(context.Background(), newKey); err != nil {
			log.With(zap.Error(err)).Errorf("Failed to deploy SSH keys for user")
			continue
		}
	}
}

// evictUser moves a user directory to evictedPath and changes their owner recursive to root.
func evictUser(username string, fs afero.Fs, linuxUserManager user.LinuxUserManager) error {
	if _, err := linuxUserManager.GetLinuxUser(username); err == nil {
		return fmt.Errorf("user '%s' still seems to exist", username)
	}

	// First, ensure evictedPath already exists.
	if err := fs.MkdirAll(evictedHomePath, 0o700); err != nil {
		return err
	}

	// Build paths to the user's home directory and evicted home directory, which includes a timestamp to avoid collisions.
	oldUserDir := path.Join(normalHomePath, username)
	evictedUserDir := path.Join(evictedHomePath, fmt.Sprintf("%s_%d", username, time.Now().Unix()))

	// Move old, not recreated user directory to evictedPath.
	if err := fs.Rename(oldUserDir, evictedUserDir); err != nil {
		return err
	}

	// Chown the user directory and all files inside to root, but do not change permissions to allow recovery without messed up permissions.
	if err := fs.Chown(evictedUserDir, 0, 0); err != nil {
		return err
	}
	if err := afero.Walk(fs, evictedUserDir, func(name string, info os.FileInfo, err error) error {
		if err == nil {
			err = fs.Chown(name, 0, 0)
		}

		return err
	}); err != nil {
		return err
	}

	return nil
}

// evictRootKey removes the root key from the filesystem, instead of evicting the whole user directory.
func evictRootKey(fs afero.Fs, linuxUserManager user.LinuxUserManager) error {
	user, err := linuxUserManager.GetLinuxUser("root")
	if err != nil {
		return err
	}

	// Delete already deployed keys
	pathToSSHKeys := filepath.Join(user.Home, relativePathToSSHKeys)
	if err := fs.Remove(pathToSSHKeys); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

// retrieveConfigMap contacts the Kubernetes API server and retrieves the ssh-users ConfigMap.
func retrieveConfigMap(log *logger.Logger) (*v1.ConfigMap, error) {
	// Authenticate with the Kubernetes API and get the information from the ssh-users ConfigMap to recreate the users we need.
	log.Infof("Authenticating with Kubernetes...")
	clientset, err := loadClientSet()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	log.Infof("Requesting 'ssh-users' ConfigMap...")
	configmap, err := clientset.CoreV1().ConfigMaps("kube-system").Get(ctx, "ssh-users", v1Options.GetOptions{})
	if err != nil {
		return nil, err
	}

	return configmap, err
}

// generateUserMap iterates the list of existing home directories to create a map of previously existing usernames to their previous respective UID and GID.
func generateUserMap(log *logger.Logger, fs afero.Fs) (map[string]uidGIDPair, error) {
	// Go through the normalHomePath directory, and create a mapping of existing user names in combination with their owner's UID & GID.
	// We use this information later to create missing users under the same UID and GID to avoid breakage.
	fileInfo, err := afero.ReadDir(fs, normalHomePath)
	if err != nil {
		return nil, err
	}

	userMap := make(map[string]uidGIDPair)
	userMap["root"] = uidGIDPair{UID: 0, GID: 0}
	// This will fail under MemMapFS, since it's not UNIX-compatible.
	for _, singleInfo := range fileInfo {
		log := log.With("username", singleInfo.Name())
		// Fail gracefully instead of hard.
		if stat, ok := singleInfo.Sys().(*syscall.Stat_t); ok {
			userMap[singleInfo.Name()] = uidGIDPair{UID: stat.Uid, GID: stat.Gid}
			log.With(zap.Uint32("UID", stat.Uid), zap.Uint32("GID", stat.Gid)).
				Infof("Found home directory for user")
		} else {
			log.Warnf("Failed to retrieve UNIX stat for user. User will not be evicted, or if this directory belongs to a user that is to be created later, it might be created under a different UID/GID than before")
			continue
		}
	}

	return userMap, nil
}

func run(log *logger.Logger, fs afero.Fs, linuxUserManager user.LinuxUserManager, configMap *v1.ConfigMap) error {
	sshAccess := ssh.NewAccess(log, linuxUserManager)

	// Generate userMap containing existing user directories and their ownership
	userMap, err := generateUserMap(log, fs)
	if err != nil {
		return err
	}

	// Try to deploy keys based on configmap.
	deployKeys(context.Background(), log, configMap, fs, linuxUserManager, userMap, sshAccess)

	return nil
}

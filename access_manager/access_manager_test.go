package main

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/edgelesssys/constellation/internal/deploy/ssh"
	"github.com/edgelesssys/constellation/internal/deploy/user"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
)

func TestEvictUser(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	fs := afero.NewMemMapFs()
	linuxUserManager := user.NewLinuxUserManagerFake(fs)

	// Create fake user directory
	homePath := path.Join(normalHomePath, "myuser")
	err := fs.MkdirAll(homePath, 0o700)
	require.NoError(err)

	// Try to evict the user
	assert.NoError(evictUser("myuser", fs, linuxUserManager))

	// Check if user has been evicted
	homeEntries, err := afero.ReadDir(fs, normalHomePath)
	require.NoError(err)
	evictedEntries, err := afero.ReadDir(fs, evictedHomePath)
	require.NoError(err)
	assert.Len(homeEntries, 0)
	assert.Len(evictedEntries, 1)
	for _, singleEntry := range evictedEntries {
		assert.Contains(singleEntry.Name(), "myuser")
	}

	/*
		Note: Unfourtunaly, due to a bug in afero, we cannot test that the files inside the directory have actually been moved.
		This works on the real filesystem, but not on the memory filesystem.
		See: https://github.com/spf13/afero/issues/141 (known since 2017, guess it will never get fixed ¯\_(ツ)_/¯)
		This limits the scope of this test, obviously... But I think as long as we can move the directory,
		the functionality on the real filesystem should be there (unless it throws an error).
	*/
}

func TestDeployKeys(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	testCases := map[string]struct {
		configMap     *v1.ConfigMap
		existingUsers map[string]uidGIDPair
	}{
		"undefined":                  {},
		"undefined map, empty users": {existingUsers: map[string]uidGIDPair{}},
		"empty map, undefined users": {configMap: &v1.ConfigMap{}},
		"both empty": {
			configMap: &v1.ConfigMap{
				Data: map[string]string{},
			},
			existingUsers: map[string]uidGIDPair{},
		},
		"create two users, no existing users": {
			configMap: &v1.ConfigMap{
				Data: map[string]string{
					"user1": "ssh-rsa abcdefgh",
					"user2": "ssh-ed25519 defghijklm",
				},
			},
			existingUsers: map[string]uidGIDPair{},
		},
		"empty configMap, user1 and user2 should be evicted": {
			configMap: &v1.ConfigMap{
				Data: map[string]string{},
			},
			existingUsers: map[string]uidGIDPair{
				"user1": {
					UID: 1000,
					GID: 1000,
				},
				"user2": {
					UID: 1001,
					GID: 1001,
				},
			},
		},
		"configMap contains user2, user1 should be evicted, user2 recreated": {
			configMap: &v1.ConfigMap{
				Data: map[string]string{
					"user2": "ssh-rsa abcdefg",
				},
			},
			existingUsers: map[string]uidGIDPair{
				"user1": {
					UID: 1000,
					GID: 1000,
				},
				"user2": {
					UID: 1001,
					GID: 1001,
				},
			},
		},
		"configMap contains user1 and user3, user1 should be recreated, user2 evicted, user3 created": {
			configMap: &v1.ConfigMap{
				Data: map[string]string{
					"user1": "ssh-rsa abcdefg",
					"user3": "ssh-ed25519 defghijklm",
				},
			},
			existingUsers: map[string]uidGIDPair{
				"user1": {
					UID: 1000,
					GID: 1000,
				},
				"user2": {
					UID: 1001,
					GID: 1001,
				},
			},
		},
		"configMap contains user1 and user3, both should be recreated": {
			configMap: &v1.ConfigMap{
				Data: map[string]string{
					"user1": "ssh-rsa abcdefg",
					"user3": "ssh-ed25519 defghijklm",
				},
			},
			existingUsers: map[string]uidGIDPair{
				"user1": {
					UID: 1000,
					GID: 1000,
				},
				"user3": {
					UID: 1002,
					GID: 1002,
				},
			},
		},
		"configMap contains user2, user1 and user3 should be evicted, user2 should be created": {
			configMap: &v1.ConfigMap{
				Data: map[string]string{
					"user2": "ssh-ed25519 defghijklm",
				},
			},
			existingUsers: map[string]uidGIDPair{
				"user1": {
					UID: 1000,
					GID: 1000,
				},
				"user3": {
					UID: 1002,
					GID: 1002,
				},
			},
		},
	}
	for _, tc := range testCases {
		fs := afero.NewMemMapFs()
		require.NoError(fs.MkdirAll(normalHomePath, 0o700))
		require.NoError(fs.Mkdir("/etc", 0o644))
		_, err := fs.Create("/etc/passwd")
		require.NoError(err)

		// Create fake user directories
		for user := range tc.existingUsers {
			userHomePath := path.Join(normalHomePath, user)
			err := fs.MkdirAll(userHomePath, 0o700)
			require.NoError(err)
			require.NoError(fs.Chown(userHomePath, int(tc.existingUsers[user].UID), int(tc.existingUsers[user].GID)))
		}

		linuxUserManager := user.NewLinuxUserManagerFake(fs)
		sshAccess := ssh.NewAccess(linuxUserManager)
		deployKeys(context.Background(), tc.configMap, fs, linuxUserManager, tc.existingUsers, sshAccess)

		// Unfourtunaly, we cannot retrieve the UID/GID from afero's MemMapFs without weird hacks,
		// as it does not have getters and it is not exported.
		if tc.configMap != nil && tc.existingUsers != nil {
			// Parse /etc/passwd and check for users
			passwdEntries, err := linuxUserManager.Passwd.Parse(fs)
			require.NoError(err)

			// Check recreation or deletion
			for user := range tc.existingUsers {
				if _, ok := tc.configMap.Data[user]; ok {
					checkHomeDirectory(user, fs, assert, true)

					// Check if user exists in /etc/passwd
					userEntry, ok := passwdEntries[user]
					assert.True(ok)

					// Check if user has been recreated with correct UID/GID
					actualUID, err := strconv.Atoi(userEntry.Uid)
					assert.NoError(err)
					assert.EqualValues(tc.existingUsers[user].UID, actualUID)
					actualGID, err := strconv.Atoi(userEntry.Gid)
					assert.NoError(err)
					assert.EqualValues(tc.existingUsers[user].GID, actualGID)

					// Check if the user has the right keys
					checkSSHKeys(user, fs, assert, tc.configMap.Data[user]+"\n")

				} else {
					// Check if home directory is not available anymore under the regular path
					checkHomeDirectory(user, fs, assert, false)

					// Check if home directory has been evicted
					homeDirs, err := afero.ReadDir(fs, evictedHomePath)
					require.NoError(err)

					var userDirectoryName string
					for _, singleDir := range homeDirs {
						if strings.Contains(singleDir.Name(), user+"_") {
							userDirectoryName = singleDir.Name()
							break
						}
					}
					assert.NotEmpty(userDirectoryName)

					// Check if user does not exist in /etc/passwd
					_, ok := passwdEntries[user]
					assert.False(ok)
				}
			}

			// Check creation of new users
			for user := range tc.configMap.Data {
				// We already checked recreated or evicted users, so skip them.
				if _, ok := tc.existingUsers[user]; ok {
					continue
				}

				checkHomeDirectory(user, fs, assert, true)
				checkSSHKeys(user, fs, assert, tc.configMap.Data[user]+"\n")
			}
		}
	}
}

func TestEvictRootKey(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	fs := afero.NewMemMapFs()

	// Create /etc/passwd with root entry
	require.NoError(fs.Mkdir("/etc", 0o644))
	file, err := fs.Create("/etc/passwd")
	require.NoError(err)
	passwdRootEntry := "root:x:0:0:root:/root:/bin/bash\n"
	n, err := file.WriteString(passwdRootEntry)
	require.NoError(err)
	require.Equal(len(passwdRootEntry), n)

	// Deploy a fake key for root
	require.NoError(fs.MkdirAll("/root/.ssh/authorized_keys.d", 0o700))
	file, err = fs.Create(filepath.Join("/root", relativePathToSSHKeys))
	require.NoError(err)
	_, err = file.WriteString("ssh-ed25519 abcdefghijklm\n")
	require.NoError(err)

	linuxUserManager := user.NewLinuxUserManagerFake(fs)

	// Parse /etc/passwd and check for users
	passwdEntries, err := linuxUserManager.Passwd.Parse(fs)
	require.NoError(err)

	// Check if user exists in /etc/passwd
	userEntry, ok := passwdEntries["root"]
	assert.True(ok)

	// Check if user has been recreated with correct UID/GID
	actualUID, err := strconv.Atoi(userEntry.Uid)
	assert.NoError(err)
	assert.EqualValues(0, actualUID)
	actualGID, err := strconv.Atoi(userEntry.Gid)
	assert.NoError(err)
	assert.EqualValues(0, actualGID)

	// Delete the key
	assert.NoError(evictRootKey(fs, linuxUserManager))

	// Check if the key has been deleted
	_, err = fs.Stat(filepath.Join("/root", relativePathToSSHKeys))
	assert.True(os.IsNotExist(err))
}

func checkSSHKeys(user string, fs afero.Fs, assert *assert.Assertions, expectedValue string) {
	// Do the same check as above
	_, err := fs.Stat(path.Join(normalHomePath, user))
	assert.NoError(err)

	// Check if the user has the right keys
	authorizedKeys, err := afero.ReadFile(fs, filepath.Join(normalHomePath, user, relativePathToSSHKeys))
	assert.NoError(err)
	assert.EqualValues(expectedValue, string(authorizedKeys))
}

func checkHomeDirectory(user string, fs afero.Fs, assert *assert.Assertions, shouldExist bool) {
	_, err := fs.Stat(path.Join(normalHomePath, user))
	if shouldExist {
		assert.NoError(err)
	} else {
		assert.Error(err)
		assert.True(os.IsNotExist(err))
	}
}

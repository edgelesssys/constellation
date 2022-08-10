/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package systemd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/afero"
)

const (
	systemdRuntimeUnitPath          = "/run/systemd/system"
	systemdUnitName                 = "systemd-cryptsetup@state.service"
	systemdDeviceRequires           = "dev-mapper-state.device.requires"
	systemdCryptsetupTargetRequires = "cryptsetup.target.requires"
)

// CryptsetupUnitGenerator generates systemd-cryptsetup@.service unit files.
type CryptsetupUnitGenerator struct {
	fs afero.Afero
}

// New returns a new CryptsetupUnitGenerator.
func New(fs afero.Afero) CryptsetupUnitGenerator {
	return CryptsetupUnitGenerator{fs: fs}
}

// Generate generates a systemd-cryptsetup@.service unit file and its dependencies.
func (g CryptsetupUnitGenerator) Generate(volumeName, encryptedDevice, keyFile, options string) error {
	unitContents, err := g.configureUnit(volumeName, encryptedDevice, keyFile, options)
	if err != nil {
		return err
	}
	return g.writeUnits(unitContents)
}

// configureUnit generates the systemd-cryptsetup@.service unit file contents.
func (g CryptsetupUnitGenerator) configureUnit(volumeName, encryptedDevice, keyFile, options string) (string, error) {
	deviceUnit := strings.ReplaceAll(encryptedDevice, "/", "-") + ".device"
	deviceUnit = strings.TrimPrefix(deviceUnit, "-")
	templ, err := template.New("").Parse(`[Unit]
Description=Cryptography Setup for %I
Documentation=man:crypttab(5) man:systemd-cryptsetup-generator(8) man:systemd-cryptsetup@.service(8)
DefaultDependencies=no
IgnoreOnIsolate=true
After=cryptsetup-pre.target systemd-udevd-kernel.socket
Before=blockdev@dev-mapper-%i.target
Wants=blockdev@dev-mapper-%i.target
Conflicts=umount.target
Before=cryptsetup.target
RequiresMountsFor={{.keyFile}}
BindsTo={{.deviceUnit}}
After={{.deviceUnit}}
Before=umount.target

[Service]
Type=oneshot
RemainAfterExit=yes
TimeoutSec=0
KeyringMode=shared
OOMScoreAdjust=500
ExecStart=/usr/lib/systemd/systemd-cryptsetup attach '{{.volumeName}}' '{{.encryptedDevice}}' '{{.keyFile}}' '{{.options}}'
ExecStop=/usr/lib/systemd/systemd-cryptsetup detach '{{.volumeName}}'
`)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = templ.Execute(&buf, map[string]string{
		"volumeName":      volumeName,
		"encryptedDevice": encryptedDevice,
		"deviceUnit":      deviceUnit,
		"keyFile":         keyFile,
		"options":         options,
	})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// writeUnits writes the unit file and its dependencies to the filesystem.
func (g CryptsetupUnitGenerator) writeUnits(unitContents string) error {
	if err := g.fs.MkdirAll(systemdRuntimeUnitPath, os.ModePerm); err != nil {
		return err
	}
	if err := g.fs.Mkdir(filepath.Join(systemdRuntimeUnitPath, systemdDeviceRequires), os.ModePerm); err != nil {
		return err
	}
	if err := g.fs.Mkdir(filepath.Join(systemdRuntimeUnitPath, systemdCryptsetupTargetRequires), os.ModePerm); err != nil {
		return err
	}
	unitPath := filepath.Join(systemdRuntimeUnitPath, systemdUnitName)
	if err := g.fs.WriteFile(unitPath, []byte(unitContents), 0o444); err != nil {
		return err
	}
	if symlinker, ok := g.fs.Fs.(afero.Symlinker); ok {
		if err := symlinker.SymlinkIfPossible(unitPath, filepath.Join(systemdRuntimeUnitPath, systemdDeviceRequires, systemdUnitName)); err != nil {
			return fmt.Errorf("creating device symlink: %w", err)
		}
		if err := symlinker.SymlinkIfPossible(unitPath, filepath.Join(systemdRuntimeUnitPath, systemdCryptsetupTargetRequires, systemdUnitName)); err != nil {
			return fmt.Errorf("creating cryptsetup target symlink: %w", err)
		}
	} else {
		return errors.New("fs does not support symlinks")
	}
	return nil
}

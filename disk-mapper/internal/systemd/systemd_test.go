/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package systemd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"))
}

func TestConfigureUnit(t *testing.T) {
	assert := assert.New(t)
	generator := CryptsetupUnitGenerator{}
	got, err := generator.configureUnit("volumeName", "/encrypted/device/path", "/key/file/path", "options")
	assert.NoError(err)
	assert.Equal(`[Unit]
Description=Cryptography Setup for %I
Documentation=man:crypttab(5) man:systemd-cryptsetup-generator(8) man:systemd-cryptsetup@.service(8)
DefaultDependencies=no
IgnoreOnIsolate=true
After=cryptsetup-pre.target systemd-udevd-kernel.socket
Before=blockdev@dev-mapper-%i.target
Wants=blockdev@dev-mapper-%i.target
Conflicts=umount.target
Before=cryptsetup.target
RequiresMountsFor=/key/file/path
BindsTo=encrypted-device-path.device
After=encrypted-device-path.device
Before=umount.target

[Service]
Type=oneshot
RemainAfterExit=yes
TimeoutSec=0
KeyringMode=shared
OOMScoreAdjust=500
ExecStart=/usr/lib/systemd/systemd-cryptsetup attach 'volumeName' '/encrypted/device/path' '/key/file/path' 'options'
ExecStop=/usr/lib/systemd/systemd-cryptsetup detach 'volumeName'
`, got)
}

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// NewUpgradeCmd returns a new cobra.Command for the upgrade command.
func NewUpgradeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Find and apply upgrades to your Constellation cluster",
		Long:  "Find and apply upgrades to your Constellation cluster.",
		Args:  cobra.ExactArgs(0),
	}

	cmd.AddCommand(newUpgradeCheckCmd())
	cmd.AddCommand(newUpgradeApplyCmd())
	return cmd
}

// upgradeCmdKind is the kind of the upgrade command (check, apply).
type upgradeCmdKind int

const (
	// upgradeCmdKindCheck corresponds to the upgrade check command.
	upgradeCmdKindCheck upgradeCmdKind = iota
	// upgradeCmdKindApply corresponds to the upgrade apply command.
	upgradeCmdKindApply
	// upgradeCmdKindIAM corresponds to the IAM upgrade command.
	upgradeCmdKindIAM
)

func generateUpgradeID(kind upgradeCmdKind) string {
	upgradeID := time.Now().Format("20060102150405") + "-" + strings.Split(uuid.New().String(), "-")[0]
	switch kind {
	case upgradeCmdKindCheck:
		// When performing an upgrade check, the upgrade directory will only be used temporarily to store the
		// Terraform state. The directory is deleted after the check is finished.
		// Therefore, add a tmp-suffix to the upgrade ID to indicate that the directory will be cleared after the check.
		upgradeID = "upgrade-" + upgradeID + "-tmp"
	case upgradeCmdKindApply:
		upgradeID = "upgrade-" + upgradeID
	case upgradeCmdKindIAM:
		upgradeID = "iam-" + upgradeID
	}
	return upgradeID
}

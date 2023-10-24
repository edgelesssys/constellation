/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/cli/internal/helm"
	"github.com/edgelesssys/constellation/v2/cli/internal/state"
	"github.com/edgelesssys/constellation/v2/internal/compatibility"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/spf13/cobra"
)

// runHelmApply handles installing or upgrading helm charts for the cluster.
func (a *applyCmd) runHelmApply(
	cmd *cobra.Command, conf *config.Config, stateFile *state.State,
	kubeUpgrader kubernetesUpgrader, upgradeDir string,
) error {
	a.log.Debugf("Installing or upgrading Helm charts")
	var masterSecret uri.MasterSecret
	if err := a.fileHandler.ReadJSON(constants.MasterSecretFilename, &masterSecret); err != nil {
		return fmt.Errorf("reading master secret: %w", err)
	}

	options := helm.Options{
		Force:            a.flags.force,
		Conformance:      a.flags.conformance,
		HelmWaitMode:     a.flags.helmWaitMode,
		AllowDestructive: helm.DenyDestructive,
	}
	helmApplier, err := a.newHelmClient(constants.AdminConfFilename, a.log)
	if err != nil {
		return fmt.Errorf("creating Helm client: %w", err)
	}

	a.log.Debugf("Getting service account URI")
	serviceAccURI, err := cloudcmd.GetMarshaledServiceAccountURI(conf, a.fileHandler)
	if err != nil {
		return err
	}

	a.log.Debugf("Preparing Helm charts")
	executor, includesUpgrades, err := helmApplier.PrepareApply(conf, stateFile, options, serviceAccURI, masterSecret)
	if errors.Is(err, helm.ErrConfirmationMissing) {
		if !a.flags.yes {
			cmd.PrintErrln("WARNING: Upgrading cert-manager will destroy all custom resources you have manually created that are based on the current version of cert-manager.")
			ok, askErr := askToConfirm(cmd, "Do you want to upgrade cert-manager anyway?")
			if askErr != nil {
				return fmt.Errorf("asking for confirmation: %w", err)
			}
			if !ok {
				cmd.Println("Skipping upgrade.")
				return nil
			}
		}
		options.AllowDestructive = helm.AllowDestructive
		executor, includesUpgrades, err = helmApplier.PrepareApply(conf, stateFile, options, serviceAccURI, masterSecret)
	}
	var upgradeErr *compatibility.InvalidUpgradeError
	if err != nil {
		if !errors.As(err, &upgradeErr) {
			return fmt.Errorf("preparing Helm charts: %w", err)
		}
		cmd.PrintErrln(err)
	}

	a.log.Debugf("Backing up Helm charts")
	if err := a.backupHelmCharts(cmd.Context(), kubeUpgrader, executor, includesUpgrades, upgradeDir); err != nil {
		return err
	}

	a.log.Debugf("Applying Helm charts")
	if !a.flags.skipPhases.contains(skipInitPhase) {
		a.spinner.Start("Installing Kubernetes components ", false)
	} else {
		a.spinner.Start("Upgrading Kubernetes components ", false)
	}

	if err := executor.Apply(cmd.Context()); err != nil {
		return fmt.Errorf("applying Helm charts: %w", err)
	}
	a.spinner.Stop()

	if a.flags.skipPhases.contains(skipInitPhase) {
		cmd.Println("Successfully upgraded Constellation services.")
	}

	return nil
}

// backupHelmCharts saves the Helm charts for the upgrade to disk and creates a backup of existing CRDs and CRs.
func (a *applyCmd) backupHelmCharts(
	ctx context.Context, kubeUpgrader kubernetesUpgrader, executor helm.Applier, includesUpgrades bool, upgradeDir string,
) error {
	// Save the Helm charts for the upgrade to disk
	chartDir := filepath.Join(upgradeDir, "helm-charts")
	if err := executor.SaveCharts(chartDir, a.fileHandler); err != nil {
		return fmt.Errorf("saving Helm charts to disk: %w", err)
	}
	a.log.Debugf("Helm charts saved to %s", a.flags.pathPrefixer.PrefixPrintablePath(chartDir))

	if includesUpgrades {
		a.log.Debugf("Creating backup of CRDs and CRs")
		crds, err := kubeUpgrader.BackupCRDs(ctx, upgradeDir)
		if err != nil {
			return fmt.Errorf("creating CRD backup: %w", err)
		}
		if err := kubeUpgrader.BackupCRs(ctx, crds, upgradeDir); err != nil {
			return fmt.Errorf("creating CR backup: %w", err)
		}
	}

	return nil
}

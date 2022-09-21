/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudtypes"
	"github.com/edgelesssys/constellation/v2/internal/state"
)

type terraformOutput struct {
	ControlPlaneIPs struct {
		Value []string `json:"value"`
	} `json:"control_plane_ips"`
	WorkerIPs struct {
		Value []string `json:"value"`
	} `json:"worker_ips"`
}

func terraformOut(workspaceDir string) (terraformOutput, error) {
	cmd := exec.Command("terraform", "output", "--json")
	cmd.Dir = workspaceDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return terraformOutput{}, fmt.Errorf("command terraform output failed: %q: %w", stderr.String(), err)
	}
	var tfOut terraformOutput
	if err := json.Unmarshal(stdout.Bytes(), &tfOut); err != nil {
		return terraformOutput{}, fmt.Errorf("unmarshaling terraform output: %w", err)
	}
	return tfOut, nil
}

func transformState(tfOut terraformOutput) state.ConstellationState {
	conState := state.ConstellationState{
		Name:                      "qemu",
		UID:                       "debug",
		CloudProvider:             cloudprovider.QEMU.String(),
		LoadBalancerIP:            tfOut.ControlPlaneIPs.Value[0],
		QEMUWorkerInstances:       cloudtypes.Instances{},
		QEMUControlPlaneInstances: cloudtypes.Instances{},
	}
	for i, ip := range tfOut.ControlPlaneIPs.Value {
		conState.QEMUControlPlaneInstances[fmt.Sprintf("control-plane-%d", i)] = cloudtypes.Instance{
			PublicIP:  ip,
			PrivateIP: ip,
		}
	}
	for i, ip := range tfOut.WorkerIPs.Value {
		conState.QEMUWorkerInstances[fmt.Sprintf("worker-%d", i)] = cloudtypes.Instance{
			PublicIP:  ip,
			PrivateIP: ip,
		}
	}
	return conState
}

func writeState(workspaceDir string, conState state.ConstellationState) error {
	rawState, err := json.Marshal(conState)
	if err != nil {
		return fmt.Errorf("marshaling state: %w", err)
	}
	stateFile := fmt.Sprintf("%s/constellation-state.json", workspaceDir)
	if err := os.WriteFile(stateFile, rawState, 0o644); err != nil {
		return fmt.Errorf("writing state: %w", err)
	}
	return nil
}

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Usage: %v <terraform-workspace-dir> <constellation-workspace-dir>\n", os.Args[0])
		os.Exit(1)
	}

	tfOut, err := terraformOut(os.Args[1])
	if err != nil {
		panic(err)
	}
	conState := transformState(tfOut)
	if err := writeState(os.Args[2], conState); err != nil {
		panic(err)
	}
}

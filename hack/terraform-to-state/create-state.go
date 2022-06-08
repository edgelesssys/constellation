package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/edgelesssys/constellation/internal/cloud/cloudtypes"
	"github.com/edgelesssys/constellation/internal/state"
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
		return terraformOutput{}, fmt.Errorf("failed to run terraform output: %q: %w", stderr.String(), err)
	}
	var tfOut terraformOutput
	if err := json.Unmarshal(stdout.Bytes(), &tfOut); err != nil {
		return terraformOutput{}, fmt.Errorf("failed to unmarshal terraform output: %w", err)
	}
	return tfOut, nil
}

func transformState(tfOut terraformOutput) state.ConstellationState {
	conState := state.ConstellationState{
		Name:             "qemu",
		UID:              "debug",
		CloudProvider:    "qemu",
		QEMUNodes:        cloudtypes.Instances{},
		QEMUCoordinators: cloudtypes.Instances{},
	}
	for i, ip := range tfOut.ControlPlaneIPs.Value {
		conState.QEMUCoordinators[fmt.Sprintf("control-plane-%d", i)] = cloudtypes.Instance{
			PublicIP:  ip,
			PrivateIP: ip,
		}
	}
	for i, ip := range tfOut.WorkerIPs.Value {
		conState.QEMUNodes[fmt.Sprintf("worker-%d", i)] = cloudtypes.Instance{
			PublicIP:  ip,
			PrivateIP: ip,
		}
	}
	return conState
}

func writeState(workspaceDir string, conState state.ConstellationState) error {
	rawState, err := json.Marshal(conState)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}
	stateFile := fmt.Sprintf("%s/constellation-state.json", workspaceDir)
	if err := os.WriteFile(stateFile, rawState, 0o644); err != nil {
		return fmt.Errorf("failed to write state: %w", err)
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

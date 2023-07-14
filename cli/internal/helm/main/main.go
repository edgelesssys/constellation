package main

import (
	"os"

	"github.com/edgelesssys/constellation/v2/cli/internal/helm"
)

func main() {
	helm.Install(os.Getenv("KUBECONFIG")) // constants.ControlPlaneAdminConfFilename)
}

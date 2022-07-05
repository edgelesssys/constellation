package resources

const (
	// Constellation images.
	joinImage          = "ghcr.io/edgelesssys/constellation/join-service:feat-coordinator-selfactivation-node"
	accessManagerImage = "ghcr.io/edgelesssys/constellation/access-manager:feat-coordinator-selfactivation-node"
	kmsImage           = "ghcr.io/edgelesssys/constellation/kmsserver:feat-coordinator-selfactivation-node"
	verificationImage  = "ghcr.io/edgelesssys/constellation/verification-service:feat-coordinator-selfactivation-node"
	gcpGuestImage      = "ghcr.io/edgelesssys/gcp-guest-agent:latest"

	// external images.
	clusterAutoscalerImage = "k8s.gcr.io/autoscaling/cluster-autoscaler:v1.23.0"
)

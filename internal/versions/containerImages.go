package versions

const (
	// Constellation images.
	JoinImage          = "ghcr.io/edgelesssys/constellation/join-service:v1.3.2-0.20220718102802-8c25a227"
	AccessManagerImage = "ghcr.io/edgelesssys/constellation/access-manager:v1.3.2-0.20220714151638-d295be31"
	KmsImage           = "ghcr.io/edgelesssys/constellation/kmsserver:v1.3.2-0.20220714151638-d295be31"
	VerificationImage  = "ghcr.io/edgelesssys/constellation/verification-service:v1.3.2-0.20220714151638-d295be31"
	GcpGuestImage      = "ghcr.io/edgelesssys/gcp-guest-agent:latest"

	// external images.
	ClusterAutoscalerImage = "k8s.gcr.io/autoscaling/cluster-autoscaler:v1.23.0"
)

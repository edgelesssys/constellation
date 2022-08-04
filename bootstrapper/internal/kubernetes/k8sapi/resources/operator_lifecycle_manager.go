package resources

import "github.com/edgelesssys/constellation/internal/crds"

// OLMCRDNames are the names of the custom resource definitions that are used by the olm operator.
var OLMCRDNames = []string{
	"catalogsources.operators.coreos.com",
	"clusterserviceversions.operators.coreos.com",
	"installplans.operators.coreos.com",
	"olmconfigs.operators.coreos.com",
	"operatorconditions.operators.coreos.com",
	"operatorgroups.operators.coreos.com",
	"operators.operators.coreos.com",
	"subscriptions.operators.coreos.com",
}

// OperatorLifecycleManagerCRDs contains custom resource definitions used by the olm operator.
type OperatorLifecycleManagerCRDs struct{}

// Marshal returns the already marshalled CRDs.
func (m *OperatorLifecycleManagerCRDs) Marshal() ([]byte, error) {
	return crds.OLMCRDs, nil
}

// OperatorLifecycleManager is the deployment of the olm operator.
type OperatorLifecycleManager struct{}

// Marshal returns the already marshalled deployment yaml.
func (m *OperatorLifecycleManager) Marshal() ([]byte, error) {
	return crds.OLM, nil
}

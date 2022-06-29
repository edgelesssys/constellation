package resources

import (
	k8s "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// ConfigMaps represent a list of k8s ConfigMap.
type ConfigMaps []*k8s.ConfigMap

// Marshal marshals config maps into multiple YAML documents.
func (s ConfigMaps) Marshal() ([]byte, error) {
	objects := make([]runtime.Object, len(s))
	for i := range s {
		objects[i] = s[i]
	}
	return MarshalK8SResourcesList(objects)
}

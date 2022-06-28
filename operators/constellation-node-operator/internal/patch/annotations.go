package patch

import "sigs.k8s.io/controller-runtime/pkg/client"

// SetAnnotations creates a patch for a client.Object by merging annotations with existing annotations.
func SetAnnotations(original, patched client.Object, annotations map[string]string) client.Patch {
	mergedAnnotations := patched.GetAnnotations()
	if mergedAnnotations == nil {
		mergedAnnotations = make(map[string]string, len(annotations))
	}
	for annotationKey, annotationValue := range annotations {
		mergedAnnotations[annotationKey] = annotationValue
	}
	patched.SetAnnotations(mergedAnnotations)
	return client.StrategicMergeFrom(original, client.MergeFromWithOptimisticLock{})
}

// UnsetAnnotations creates a patch for a client.Object by deleting annotations from the object.
func UnsetAnnotations(original, patched client.Object, annotationKeys []string) client.Patch {
	mergedAnnotations := patched.GetAnnotations()
	for _, annotationKey := range annotationKeys {
		delete(mergedAnnotations, annotationKey)
	}
	patched.SetAnnotations(mergedAnnotations)
	return client.StrategicMergeFrom(original, client.MergeFromWithOptimisticLock{})
}

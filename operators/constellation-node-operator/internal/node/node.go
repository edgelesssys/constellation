package node

import corev1 "k8s.io/api/core/v1"

func Ready(node *corev1.Node) bool {
	for _, cond := range node.Status.Conditions {
		if cond.Type == corev1.NodeReady {
			if cond.Status == corev1.ConditionTrue {
				return true
			}
			return false
		}
	}
	return false
}

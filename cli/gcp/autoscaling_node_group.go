package gcp

import "fmt"

func AutoscalingNodeGroup(project string, zone string, nodeInstanceGroup string, min int, max int) string {
	return fmt.Sprintf("%d:%d:https://www.googleapis.com/compute/v1/projects/%s/zones/%s/instanceGroups/%s", min, max, project, zone, nodeInstanceGroup)
}

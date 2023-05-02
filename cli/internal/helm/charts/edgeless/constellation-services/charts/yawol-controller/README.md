# yawol-controller

![Version: 0.12.0](https://img.shields.io/badge/Version-0.12.0-informational?style=flat-square) ![AppVersion: v0.12.0](https://img.shields.io/badge/AppVersion-v0.12.0-informational?style=flat-square)

Helm chart for yawol-controller

## Source Code

* <https://github.com/stackitcloud/yawol>

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| featureGates | object | `{}` |  |
| namespace | string | `"kube-system"` |  |
| podAnnotations | object | `{}` |  |
| podLabels | object | `{}` |  |
| proxy | object | `{}` |  |
| replicas | int | `1` |  |
| resources.yawolCloudController.limits.cpu | string | `"500m"` |  |
| resources.yawolCloudController.limits.memory | string | `"512Mi"` |  |
| resources.yawolCloudController.requests.cpu | string | `"100m"` |  |
| resources.yawolCloudController.requests.memory | string | `"64Mi"` |  |
| resources.yawolControllerLoadbalancer.limits.cpu | string | `"500m"` |  |
| resources.yawolControllerLoadbalancer.limits.memory | string | `"512Mi"` |  |
| resources.yawolControllerLoadbalancer.requests.cpu | string | `"100m"` |  |
| resources.yawolControllerLoadbalancer.requests.memory | string | `"64Mi"` |  |
| resources.yawolControllerLoadbalancermachine.limits.cpu | string | `"500m"` |  |
| resources.yawolControllerLoadbalancermachine.limits.memory | string | `"512Mi"` |  |
| resources.yawolControllerLoadbalancermachine.requests.cpu | string | `"100m"` |  |
| resources.yawolControllerLoadbalancermachine.requests.memory | string | `"64Mi"` |  |
| resources.yawolControllerLoadbalancerset.limits.cpu | string | `"500m"` |  |
| resources.yawolControllerLoadbalancerset.limits.memory | string | `"512Mi"` |  |
| resources.yawolControllerLoadbalancerset.requests.cpu | string | `"100m"` |  |
| resources.yawolControllerLoadbalancerset.requests.memory | string | `"64Mi"` |  |
| vpa.enabled | bool | `false` |  |
| vpa.yawolCloudController.mode | string | `"Auto"` |  |
| vpa.yawolController.mode | string | `"Auto"` |  |
| yawolAPIHost | string | `nil` |  |
| yawolAvailabilityZone | string | `""` |  |
| yawolCloudController.clusterRoleEnabled | bool | `true` |  |
| yawolCloudController.enabled | bool | `true` |  |
| yawolCloudController.gardenerMonitoringEnabled | bool | `false` |  |
| yawolCloudController.image.repository | string | `"ghcr.io/stackitcloud/yawol/yawol-cloud-controller"` |  |
| yawolCloudController.image.tag | string | `""` | Allows you to override the yawol version in this chart. Use at your own risk. |
| yawolController.gardenerMonitoringEnabled | bool | `false` |  |
| yawolController.image.repository | string | `"ghcr.io/stackitcloud/yawol/yawol-controller"` |  |
| yawolController.image.tag | string | `""` | Allows you to override the yawol version in this chart. Use at your own risk. |
| yawolFlavorID | string | `nil` |  |
| yawolFloatingID | string | `nil` |  |
| yawolImageID | string | `nil` |  |
| yawolNetworkID | string | `nil` |  |
| yawolOSSecretName | string | `nil` |  |


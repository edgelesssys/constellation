# Observability

In Kubernetes, observability is the ability to gain insight into the behavior and performance of applications.
It helps identify and resolve issues more effectively, ensuring stability and performance of Kubernetes workloads, reducing downtime and outages, and improving efficiency.
The "three pillars of observability" are logs, metrics, and traces.

In the context of Confidential Computing, observability is a delicate subject and needs to be applied such that it doesn't leak any sensitive information.
The following gives an overview of where and how you can apply standard observability tools in Constellation.

## Cloud resource monitoring

While inaccessible, Constellation's nodes are still visible as black box VMs to the hypervisor.
Resource consumption, such as memory and CPU utilization, can be monitored from the outside and observed via the cloud platforms directly.
Similarly, other resources, such as storage and network and their respective metrics, are visible via the cloud platform.

## Metrics

Metrics are numeric representations of data measured over intervals of time. They're essential for understanding system health and gaining insights using telemetry signals.

By default, Constellation exposes the [metrics for Kubernetes system components](https://kubernetes.io/docs/concepts/cluster-administration/system-metrics/) inside the cluster.
Similarly, the [etcd metrics](https://etcd.io/docs/v3.5/metrics/) endpoints are exposed inside the cluster.
These [metrics endpoints can be disabled](https://kubernetes.io/docs/concepts/cluster-administration/system-metrics/#disabling-metrics).

You can collect these cluster-internal metrics via tools such as [Prometheus](https://prometheus.io/) or the [Elastic Stack](https://www.elastic.co/de/elastic-stack/).

Constellation's CNI Cilium also supports [metrics via Prometheus endpoints](https://docs.cilium.io/en/latest/observability/metrics/).
However, in Constellation, they're disabled by default and must be enabled first.

## Logs

Logs represent discrete events that usually describe what's happening with your service.
The payload is an actual message emitted from your system along with a metadata section containing a timestamp, labels, and tracking identifiers.

### System logs

Detailed system-level logs are accessible via `/var/log` and [journald](https://www.freedesktop.org/software/systemd/man/systemd-journald.service.html) on the nodes directly.
They can be collected from there, for example, via [Filebeat and Logstash](https://www.elastic.co/guide/en/beats/filebeat/current/logstash-output.html), which are tools of the [Elastic Stack](https://www.elastic.co/de/elastic-stack/).

In case of an error during the initialization, the CLI automatically collects the [Bootstrapper](./microservices.md#bootstrapper) logs and returns these as a file for [troubleshooting](../workflows/troubleshooting.md). Here is an example of such an event:

```shell-session
Cluster initialization failed. This error is not recoverable.
Terminate your cluster and try again.
Fetched bootstrapper logs are stored in "constellation-cluster.log"
```

### Kubernetes logs

Constellation supports the [Kubernetes logging architecture](https://kubernetes.io/docs/concepts/cluster-administration/logging/).
By default, logs are written to the nodes' encrypted state disks.
These include the Pod and container logs and the [system component logs](https://kubernetes.io/docs/concepts/cluster-administration/logging/#system-component-logs).

[Constellation services](microservices.md) run as Pods inside the `kube-system` namespace and use the standard container logging mechanism.
The same applies for the [Cilium Pods](https://docs.cilium.io/en/latest/operations/troubleshooting/#logs).

You can collect logs from within the cluster via tools such as [Fluentd](https://github.com/fluent/fluentd), [Loki](https://github.com/grafana/loki), or the [Elastic Stack](https://www.elastic.co/de/elastic-stack/).

## Traces

Modern systems are implemented as interconnected complex and distributed microservices. Understanding request flows and system communications is challenging, mainly because all systems in a chain need to be modified to propagate tracing information. Distributed tracing is a new approach to increasing observability and understanding performance bottlenecks. A trace represents consecutive events that reflect an end-to-end request path in a distributed system.

Constellation supports [traces for Kubernetes system components](https://kubernetes.io/docs/concepts/cluster-administration/system-traces/).
By default, they're disabled and need to be enabled first.

Similarly, Cilium can be enabled to [export traces](https://cilium.io/use-cases/metrics-export/).

You can collect these traces via tools such as [Jaeger](https://www.jaegertracing.io/) or [Zipkin](https://zipkin.io/).

## Integrations

Platforms and SaaS solutions such as Datadog, logz.io, Dynatrace, or New Relic facilitate the observability challenge for Kubernetes and provide all-in-one SaaS solutions.
They install agents into the cluster that collect metrics, logs, and tracing information and upload them into the data lake of the platform.
Technically, the agent-based approach is compatible with Constellation, and attaching these platforms is straightforward.
However, you need to evaluate if the exported data might violate Constellation's compliance and privacy guarantees by uploading them to a third-party platform.

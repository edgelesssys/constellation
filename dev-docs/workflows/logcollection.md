# Logcollection

One can deploy [Filebeat](https://www.elastic.co/guide/en/beats/filebeat/current/index.html) and [Logstash](https://www.elastic.co/guide/en/logstash/current/index.html) to enable collection of logs to [OpenSearch](https://search-e2e-logs-y46renozy42lcojbvrt3qq7csm.eu-central-1.es.amazonaws.com/_dashboards/app/home#/), which allows for agreggation and easy inspection of said logs.
The logcollection functionality can be deployed to both [debug](./debug-cluster.md) and non-debug clusters.

## Deployment in Debug Clusters

In debug clusters, logcollection functionality should be deployed automatically through the debug daemon `debugd`, which runs *before* the bootstrapper
and can therefore, contrary to non-debug clusters, also collect logs of the bootstrapper.

> [!WARNING]
> If logs from a E2E test run for a debug-cluster with a bootstrapping-failure are missing in OpenSearch, this might be caused by a race condition
> between the termination of the cluster and the start-up of the logcollection containers in the debugd.
> If the failure can be reproduced manually, it is best to do so and observe the serial console of the bootstrapping node with the following command until the logcollection containers have started.
> ```bash
> journalctl _SYSTEMD_UNIT=debugd.service | grep > logcollect
> ```

## Deployment in Non-Debug Clusters

In non-debug clusters, logcollection functionality needs to be explicitly deployed as a Kubernetes Deployment through Helm. To do that, a few steps need to be followed:

1. Template the deployment configuration through the `loco` CLI.

    ```bash
    bazel run //hack/logcollector template -- \
        --dir $(realpath .) \
        --username <OPENSEARCH_USERNAME> \
        --password <OPENSEARCH_PW> \
        --info deployment-type={k8s, debugd}
        ...
    ```

    This will place the templated configuration in the current directory. OpenSearch user credentials can be created by any admin in OpenSearch.
    Logging in with your company CSP accounts should grant you sufficient permissions to [create a user](https://opensearch.org/docs/latest/security/access-control/users-roles/#create-users)
    and [grant him the required `all_access` role](https://opensearch.org/docs/latest/security/access-control/users-roles/#map-users-to-roles).
    One can add additional key-value pairs to the configuration by appending `--info key=value` to the command.
    These key-value pairs will be attached to the log entries and can be used to filter them in OpenSearch.
    For example, it might be helpful to add a `test=<xyz>` tag to be able to filter out logs from a specific test run.
2. Add the Elastic Helm repository
    ```bash
    helm repo add elastic https://helm.elastic.co
    helm repo update
    ```
2. Deploy Logstash

    ```bash
    cd logstash
    helm install logstash elastic/logstash \
        --wait --timeout=1200s --values values.yml
    cd ..
    ```

    This will add the required Logstash Helm charts and deploy them to your cluster.
2. Deploy Beats

    ```bash
    cd metricbeat
    helm install metricbeat-k8s elastic/metricbeat \
        --wait --timeout=1200s --values values-control-plane.yml
    helm install metricbeat-system elastic/metricbeat \
        --wait --timeout=1200s --values values-all-nodes.yml
    cd ..
    cd filebeat
    helm install filebeat elastic/filebeat \
        --wait --timeout=1200s --values values.yml
    cd ..
    ```

    This will add the required Filebeat and Metricbeat Helm charts and deploy them to your cluster.

To remove Logstash or one of the beats, `cd` into the corresponding directory and run `helm uninstall {logstash,filebeat,metricbeat}`.

## Inspecting Logs in OpenSearch

To search through logs in OpenSearch, head to the [discover page](https://search-e2e-logs-y46renozy42lcojbvrt3qq7csm.eu-central-1.es.amazonaws.com/_dashboards/app/discover) in the
OpenSearch dashboard and configure the timeframe selector in the top right accordingly.
Click `Refresh`. You can now see all logs recorded in the specified timeframe. To get a less cluttered view, select the fields you want to inspect in the left sidebar.

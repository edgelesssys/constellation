# First steps

The following steps will guide you through the process of creating a cluster and deploying a sample app. This example assumes that you have successfully [installed and set up Constellation](install.md).

## Create a cluster

1. Create the configuration file for your selected cloud provider.

    <tabs>
    <tabItem value="azure" label="Azure" default>

    On Azure you also need a *user-assigned managed identity* with the [correct permissions](install.md?id=authorization).

    Then execute:

    ```bash
    constellation config generate azure
    ```

    </tabItem>
    <tabItem value="gcp" label="GCP" default>

    ```bash
    constellation config generate gcp
    ```

    </tabItem>
    </tabs>

    This creates the file `constellation-conf.yaml` in your current working directory. Edit this file to set your cloud subscription IDs and optionally customize further options of your Constellation cluster. All configuration options are documented in this file.

    For more details, see the [reference section](../reference/config.md#required-customizations).

2. Download the measurements for your configured image.

    ```bash
    constellation config fetch-measurements
    ```

    This command is necessary to download the latest trusted measurements for your configured image.

    For more details, see the [verification section](../workflows/verify.md).

3. Create the cluster with one control-plane node and two worker nodes. `constellation create` uses options set in `constellation-conf.yaml` automatically.

    <tabs>
    <tabItem value="azure" label="Azure" default>

    ```bash
    constellation create azure --control-plane-nodes 1 --worker-nodes 2 --instance-type Standard_D4a_v4 -y
    ```

    </tabItem>
    <tabItem value="gcp" label="GCP" default>

    ```bash
    constellation create gcp --control-plane-nodes 1 --worker-nodes 2 --instance-type n2d-standard-2 -y
    ```

    </tabItem>
    </tabs>

    This should give the following output:

    ```shell-session
    $ constellation create ...
    Your Constellation cluster was created successfully.
    ```

4. Initialize the cluster

    ```bash
    constellation init
    ```

    This should give the following output:

    ```shell-session
    $ constellation init
    Creating service account ...
    Your Constellation cluster was successfully initialized.
    Constellation cluster's identifier  g6iMP5wRU1b7mpOz2WEISlIYSfdAhB0oNaOg6XEwKFY=
    Kubernetes configuration            constellation-admin.conf
    You can now connect to your cluster by executing:
            export KUBECONFIG="$PWD/constellation-admin.conf"
    ```

    The cluster's identifier will be different in your output.
    Keep `constellation-mastersecret.json` somewhere safe.
    This will allow you to [recover your cluster](../workflows/recovery.md) in case of a disaster.

5. Configure kubectl

    ```bash
    export KUBECONFIG="$PWD/constellation-admin.conf"
    ```

## Deploy a sample application

1. Deploy the [emojivoto app](https://github.com/BuoyantIO/emojivoto)

    ```bash
    kubectl apply -k github.com/BuoyantIO/emojivoto/kustomize/deployment
    ```

2. Expose the frontend service locally

    ```bash
    kubectl wait --for=condition=available --timeout=60s -n emojivoto --all deployments
    kubectl -n emojivoto port-forward svc/web-svc 8080:80 &
    curl http://localhost:8080
    kill %1
    ```

## Terminate your cluster

```bash
constellation terminate
```

This should give the following output:

```shell-session
$ constellation terminate
Terminating ...
Your Constellation cluster was terminated successfully.
```

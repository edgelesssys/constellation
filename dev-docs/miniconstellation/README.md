# Quick setup for a mini cluster running the cloud

### Prerequisites

* [Install Terraform](https://developer.hashicorp.com/terraform/downloads)
* [Install the Azure CLI](https://learn.microsoft.com/de-de/cli/azure/install-azure-cli)
  to authenticate with Azure:

  ```sh
  az login
  ```

### Instructions

Through the Terraform template for Azure it's easy to set up a MiniConstellation cluster on a remote VM.

1. Clone the Constellation repository:

    ```sh
    git clone https://github.com/edgelesssys/constellation.git
    ```

2. Set up the remote Azure VM through Terraform:

    By default, the [`Standard_D8s_v5`](https://learn.microsoft.com/de-de/azure/virtual-machines/dv5-dsv5-series) machine type is selected which supports nested virtualization, but you can also set another supported machine type in Terraform (`machine_type`) by referring to the [Azure docs](https://azure.microsoft.com/en-us/blog/nested-virtualization-in-azure/).
    Then run:

    ```sh
    cd constellation/dev-docs/miniconstellation/azure-terraform
    ./create-vm.sh
    ```

    After execution, you should be connected with the remote machine through SSH.
    If you accidentally lose connection, you can reconnect via

    ```sh
    ssh -i id_rsa adminuser@$INSERT_VM_IP_ADDRESS
    ```

3. Prepare the VM for `constellation mini up`

    Once logged into the machine, install the Constellation CLI:

    ```sh
    echo "Installing Constellation CLI"
    curl -LO <https://github.com/edgelesssys/constellation/releases/latest/download/constellation-linux-amd64>
    sudo install constellation-linux-amd64 /usr/local/bin/constellation
    ```

    and start the Docker service and make sure that it's running:

    ```sh
    sudo systemctl start docker.service && sudo systemctl enable docker.service
    # verify that it is active
    systemctl is-active docker
    ```

    At last, create the Constellation cluster in a workspace directory:

    ```sh
    mkdir constellation_workspace && cd constellation_workspace
    constellation mini up
    ```

    The cluster creation takes about 15 minutes.

   For convenience, there is a script that does these steps automatically:

   ```sh
   ./setup-miniconstellation.sh
   ```

4. Verify the Kubernetes cluster

    Running:

      ```sh
      export KUBECONFIG="$PWD/constellation-admin.conf"
      kubectl get nodes
      ```

      should show both one control-plane and one worker node.

5. Clean up cloud resources

    Exit the SSH connection (Ctrl+D) and run:

    ```sh
    terraform destroy
    ```

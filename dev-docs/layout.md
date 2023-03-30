# Repository Layout

Core components:

* [cli](/cli): The CLI is used to manage a Constellation cluster
* [bootstrapper](/bootstrapper): The bootstrapper is a node agent whose most important task is to bootstrap a node
* [image](/image): Build files for the Constellation disk image
* [keyservice](/keyservice): Constellation's key management client and server
* [csi](/csi): Package used by CSI plugins to create and mount encrypted block devices
* [disk-mapper](/disk-mapper): Contains the disk-mapper that maps the encrypted node data disk during boot

Development components:

* [3rdparty](/3rdparty): Contains the third party dependencies used by Constellation
* [debugd](/debugd): Debug daemon and client
* [hack](/hack): Development tools

Additional repositories:

* [constellation-azuredisk-csi-driver](https://github.com/edgelesssys/constellation-azuredisk-csi-driver): Azure CSI driver with encryption on node
* [constellation-gcp-compute-persistent-disk-csi-driver](https://github.com/edgelesssys/constellation-gcp-compute-persistent-disk-csi-driver): GCP CSI driver with encryption on node

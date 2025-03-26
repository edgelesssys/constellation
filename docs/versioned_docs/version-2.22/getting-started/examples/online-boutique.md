# Online Boutique
[Online Boutique](https://github.com/GoogleCloudPlatform/microservices-demo) is an e-commerce demo application by Google consisting of 11 separate microservices. In this demo, Constellation automatically sets up a load balancer on your CSP, making it easy to expose services from your confidential cluster.

<!-- vale off -->

<img src={require("../../_media/example-online-boutique.jpg").default} alt="Online Boutique - Web UI" width="662"/>

<!-- vale on -->

1. Create a namespace:
    ```bash
    kubectl create ns boutique
    ```
2. Deploy the application:
    ```bash
    kubectl apply -n boutique -f https://github.com/GoogleCloudPlatform/microservices-demo/raw/main/release/kubernetes-manifests.yaml
    ```
3. Wait for all services to become available:
    ```bash
    kubectl wait --for=condition=available --timeout=300s -n boutique --all deployments
    ```
4. Get the frontend's external IP address:
    ```shell-session
    $ kubectl get service frontend-external -n boutique | awk '{print $4}'
    EXTERNAL-IP
    <your-ip>
    ```
    (`<your-ip>` is a placeholder for the IP assigned by your CSP.)
5. Enter the IP from the result in your browser to browse the online shop.

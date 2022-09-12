# Emojivoto
[Emojivoto](https://github.com/BuoyantIO/emojivoto) is a simple and fun application that's well suited to test the basic functionality of your cluster.

<!-- vale off -->

<img src={require("../../_media/example-emojivoto.jpg").default} alt="emojivoto - Web UI" width="552"/>

<!-- vale on -->

1. Deploy the application:
    ```bash
    kubectl apply -k github.com/BuoyantIO/emojivoto/kustomize/deployment
    ```
2. Wait until it becomes available:
    ```bash
    kubectl wait --for=condition=available --timeout=60s -n emojivoto --all deployments
    ```
3. Forward the web service to your machine:
    ```bash
    kubectl -n emojivoto port-forward svc/web-svc 8080:80
    ```
4. Visit [http://localhost:8080](http://localhost:8080)

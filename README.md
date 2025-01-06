
# ErrorKube Operator
![sggw](https://raw.githubusercontent.com/Ksanthosh20ka1a0554/ErrorKube/main/errorkube-frontend/public/images/Error_Kube-logo.png)
## Overview

The **ErrorKube Operator** is a Kubernetes Operator designed to manage real-time error detection and streaming within Kubernetes clusters. The operator automates the deployment of a backend that collects error events from Kubernetes resources like pods, deployments, and namespaces, and streams them to a frontend UI. The UI, built with React, displays these errors in real time for monitoring purposes. This project is designed to make error monitoring and management more efficient by providing an easy way to track and manage error events.

### Key Features:

-   **Real-time Error Streaming**: The operator streams error events from Kubernetes to a frontend UI.
-   **Unified Backend & Frontend**: The React frontend is served directly from the Go backend, simplifying the deployment.
-   **MongoDB Integration**: Errors are stored in a MongoDB database for easier tracking.
-   **Kubernetes RBAC Management**: Proper permissions are managed using Kubernetes RBAC for accessing event data.

## Description

**ErrorKube** is designed to detect errors in various Kubernetes resources and stream these errors to a frontend UI. The operator automates the deployment and management of both the backend (Go application) and frontend (React application). The backend fetches Kubernetes events and logs, while the frontend displays them in real-time using WebSockets.

MongoDB is used to store error events for persistent logging and historical analysis.

## Getting Started

### Prerequisites

To deploy the ErrorKube Operator, ensure you have the following tools installed:

-   **kubectl version v1.11.3+**
-   **Access to a Kubernetes v1.11.3+ cluster**

## Installation Steps

To install the ErrorKube operator without needing to download the repository, follow these steps:

### 1. Apply the Operator:

Simply apply the `install.yaml` file from the GitHub repository using the following command:
```sh
kubectl apply -f https://raw.githubusercontent.com/ksanthosh20ka1a0554/errorkube-operator/main/install.yaml 
```
This will install the operator and all necessary resources.

### 2. Create an ErrorKube instance:

Once the operator is installed, you can create an instance of the ErrorKube resource by applying the example configuration:

```sh
kubectl apply -f https://raw.githubusercontent.com/ksanthosh20ka1a0554/errorkube-operator/main/example.yml
```
This will create a new `ErrorKube` resource in your cluster. The operator will automatically:

-   Configure RBAC permissions.
-   Create a MongoDB deployment and service.
-   Create the backend deployment and expose it via a service with the specified NodePort.

If you're using **Minikube**, use port forwarding to access the application:

```sh
kubectl port-forward service/backend-service <local-port>:<nodeport> 
```
If you're using a **cloud cluster**, use the public IP of the node and the specified NodePort to access the frontend:
```sh
<node-public-ip>:<nodeport>
```
## Contributing

Contributions are invited to this project! Feel free to fork the repository and submit issues or pull requests. Here's how you can contribute:

1.  Fork the repository.
2.  Create a new branch for your changes.
3.  Submit a pull request with a detailed explanation of your changes.

More information can be found via the[ Operator SDK Documentation](https://sdk.operatorframework.io/docs/).

If you have any questions or would like to discuss an issue, feel free to open an issue on GitHub.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
# Name

IBM&reg; WebSphere&reg; Liberty

# Introduction

IBM WebSphere Liberty is a modern Java EE, Jakarta EE, MicroProfile runtime, ideal for building new cloud-native applications and modernizing existing applications. Liberty is highly efficient and optimized for modern cloud technologies and practices, making it an ideal choice for container and Kubernetes-based deployments.

## Details 

The WebSphere Liberty Operator allows you to deploy and manage applications running on WebSphere Liberty into Kubernetes-based platforms, such as Red Hat OpenShift. You can also perform Day-2 operations such as gathering traces and dumps using the operator.

## Supported platforms

Red Hat OpenShift Container Platform 4.8 or newer installed on one of the following platforms:
- Linux x86_64

## Prerequisites

Review the [system requirements](https://ibm.biz/wlo-sys-req) for details. 

## Resources Required

Review the [resource requirements](https://ibm.biz/wlo-reqs) before you plan to install IBM WebSphere Liberty Operator.

## Storage

Please see the [storage requirements](https://ibm.biz/wlo-reqs) for details.

## Limitations 

IBM WebSphere Liberty Operator is not available on Power or Z architectures. Please see the [limitations](https://ibm.biz/wlo-limits) for additional information.

## Documentation

See [IBM WebSphere Liberty documentation](https://ibm.biz/wlo-docs).

## SecurityContextConstraints Requirements

The IBM WebSphere Liberty Operator runs in the `restricted` security context constraints.

## Installing

Install the IBM WebSphere Liberty Operator to the desired namespace and create an instance of the [WebSphereLibertyApplication custom resource](https://ibm.biz/wlo-crs).

## Configuration

The WebSphere Liberty Operator provides various customization options to configure the application deployments. Please see [custom resource](https://ibm.biz/wlo-crs) for details.

## Key Features

### Application Lifecyle
You can deploy your Liberty application container by either pointing to a container image, or an OpenShift ImageStream. When using an ImageStream the Operator will watch for any updates and will re-deploy the modified image.

### Custom RBAC
This Operator is capable of using a custom ServiceAccount from the caller, allowing it to follow RBAC restrictions. By default it creates a ServiceAccount if one is not specified, which can also be bound with specific roles.

### Environment Configuration
You can configure a variety of artifacts with your deployment, such as: labels, annotations, and environment variables from a ConfigMap, a Secret or a value.

### Routing
Expose your application to external users via a single toggle to create a Route on OpenShift or an Ingress on other Kubernetes environments. Advanced configuration, such as TLS settings, are also easily enabled. Renewed certificates are automatically made available to the Liberty server.

### High Availability via Horizontal Pod Autoscaling
Run multiple instances of your application for high availability. Either specify a static number of replicas or easily configure horizontal auto scaling to create (and delete) instances based on resource consumption.

### Persistence and advanced storage
Enable persistence for your application by specifying simple requirements: just tell us the size of the storage and where you would like it to be mounted and we will create and manage that storage for you. This toggles a StatefulSet resource instead of a Deployment resource, so your container can recover transactions and state upon a pod restart. We offer an advanced mode where you can specify a built-in PersistentVolumeClaim, allowing to configure many details of the persistent volume, such as its storage class and access mode. You can also easily configure and use a single storage for serviceability related day-2 operations, such as gatherig server traces and dumps.

### Service Binding
Your runtime components can expose services by a simple toggle. We take care of the heavy lifting such as creating kubernetes Secrets with information other services can use to bind. We also keep the bindable information synchronized, so your applications can dynamically reconnect to its required services without any intervention or interruption.

### Single Sign-On (SSO)
Liberty runtime provides capabilities to delegate authentication to external providers. Your application users can log in using their existing social media credentials from providers such as Google, Facebook, LinkedIn, Twitter, GitHub, and any OpenID Connect (OIDC) or OAuth 2.0 clients. WebSphere Liberty Operator allows to easily configure and manage the single sign-on information for your applications.

### Exposing metrics to Prometheus
Expose the Liberty application's metrics via the Prometheus Operator.
You can pick between a basic mode, where you simply specify the label that Prometheus is watching to scrape the metrics from the container, or you can specify the full `ServiceMonitor` spec embedded into the WebSphereLibertyApplication's `.spec.monitoring` field to control configurations such as poll interval and security credentials.

### Easily mount logs and transaction directories
Do you need to mount the logs and transaction data from your application to an external volume such as NFS (or any storage supported in your cluster)? Simply add the following configuration (to specify the volume size and the location to persist) to your WebSphereLibertyApplication CR:
``` storage: size: 2Gi mountPath: "/logs" ```

### Integration with Certificate Managers
The perator will automatically provision TLS certificates for pods as well as routes and it is automatically refreshed when the certificates are updated. The [cert-manager APIs](https://cert-manager.io/) when available on the cluster will be used to generate certificates. Otherwise, on Red Hat OpenShift, the operator will generate certificates using OpenShift's Certificate Manager. The operator will automatically provision TLS certificates for applications' pods as well as routes and they are automatically refreshed when the certificates are updated.

### Control network communication
Network policies are created for each application by default to limit incoming traffic to pods in the same namespace that are part of the same application. Only the ports configured by the service are allowed. The network policy can be configured to allow either namespaces and/or pods with certain labels. On OpenShift, operator automatically configures policy to allow traffic from ingress, when application is exposed, and from moniotoring stack.

### Integration with OpenShift Serverless
Deploy your serverless runtime component using a single toggle. The Operator will convert all of its generated resources into [Knative](https://knative.dev) resources, allowing your pod to automatically scale to 0 when it is idle.

### Integration with OpenShift's Topology UI
We set the corresponding labels to support OpenShift's Developer Topology UI, which allows you to visualize your entire set of deployments and how they are connected.


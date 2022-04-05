# Name

IBM&reg; WebSphere&reg; Liberty

# Introduction

IBM WebSphere Liberty is a modern Java EE, Jakarta EE, MicroProfile runtime, ideal for building new cloud-native applications and modernizing existing applications. Liberty is highly efficient and optimized for modern cloud technologies and practices, making it an ideal choice for container and Kubernetes-based deployments.

## Details 

WebSphere Liberty Operator enables enterprise architects to govern the way their applications get deployed & managed in the cluster, while dramatically reducing the learning curve for developers to deploy into Kubernetes - allowing them to focus on writing the code!

## Prerequisites

Review the [system requirements](https://ibm.biz/wlo-sys-req) for details. 

## Resources Required

Review the [resource requirements](https://ibm.biz/wlo-reqs) before you plan to install IBM WebSphere Liberty operator.

## Storage

Please see the [storage requirements](https://ibm.biz/wlo-reqs) for details.

## Supported platforms

Red Hat OpenShift Container Platform 4.8 or newer installed on one of the following platforms:
- Linux x86_64

## Documentation

See [IBM WebSphere Liberty documentation](https://ibm.biz/wlo-docs).

## SecurityContextConstraints Requirements

The IBM WebSphere Liberty operator runs in the `restricted` security context constraints.

## Installing

Install the IBM WebSphere Liberty operator to the desired namespace and create an instance of the [WebSphereLibertyApplication custom resource](https://ibm.biz/wlo-crs).

## Configuration

The WebSphere Liberty operator provides various customization options to configure the application deployments. Please see [custom resource](https://ibm.biz/wlo-crs) for details.

### Key Features

#### Application Lifecyle
You can deploy your WebSphere Liberty application container by either pointing to a container image, or an OpenShift ImageStream. When using an ImageStream the Operator will watch for any updates and will re-deploy the modified image.

#### Custom RBAC
This Operator is capable of using a custom ServiceAccount from the caller, allowing it to follow RBAC restrictions. By default it creates a ServiceAccount if one is not specified, which can also be bound with specific roles.

#### Environment Configuration
You can configured a variety of artifacts with your deployment, such as: labels, annotations, and environment variables from a ConfigMap, a Secret or a value.

#### Routing
Expose your application to external users via a single toggle to create a Route on OpenShift or an Ingress on other Kubernetes environments. Advanced configuration, such as TLS settings, are also easily enabled. Expiring Route certificates are re-issued.

#### High Availability via Horizontal Pod Autoscaling
Run multiple instances of your application for high availability. Either specify a static number of replicas or easily configure horizontal auto scaling to create (and delete) instances based on resource consumption.

#### Persistence and advanced storage
Enable persistence for your application by specifying simple requirements: just tell us the size of the storage and where you would like it to be mounted and We will create and manage that storage for you. This toggles a StatefulSet resource instead of a Deployment resource, so your container can recover transactions and state upon a pod restart. We offer an advanced mode where the user specifies a built-in PersistentVolumeClaim, allowing them to configure many details of the persistent volume, such as its storage class and access mode. You can also easily configure and use a single storage for serviceability related Day-2 operations, such as gatherig server traces and dumps.

#### Service Binding
Your runtime components can expose services by a simple toggle. We take care of the heavy lifting such as creating kubernetes Secrets with information other services can use to bind. We also keep the bindable information synchronized, so your applications can dynamically reconnect to its required services without any intervention or interruption.

#### Single Sign-On (SSO)
WebSphere Liberty provides capabilities to delegate authentication to external providers. Your application users can log in using their existing social media credentials from providers such as Google, Facebook, LinkedIn, Twitter, GitHub, and any OpenID Connect (OIDC) or OAuth 2.0 clients. WebSphere Liberty Operator allows to easily configure and manage the single sign-on information for your applications.

#### Exposing metrics to Prometheus
The WebSphere Liberty Operator exposes the runtime container's metrics via the [Prometheus Operator](https://operatorhub.io/operator/prometheus). Users can pick between a basic mode, where they simply specify the label that Prometheus is watching to scrape the metrics from the container, or they can specify the full `ServiceMonitor` spec embedded into the WebSphereLibertyApplication's `spec.monitoring` key controlling things like the poll internal and security credentials.

#### Easily mount logs and transaction directories
If you need to mount the logs and transaction data from your application to an external volume such as NFS (or any storage supported in your cluster), simply add the following (customizing the folder location and size) to your WebSphereLibertyApplication CR: ``` storage: size: 2Gi mountPath: \"/logs\" ```

#### Integration with OpenShift's Certificate Manager
The WebSphere Liberty Operator takes advantage of the [cert-manager tool](https://cert-manager.io/), if it is installed on the cluster. This allows the operator to automatically provision TLS certificates for pods as well as routes. When creating certificates via the WebSphereLibertyApplication CR the user can specify a particular issuer name and toggle the scopes between ClusterIssuer (cluster scoped) and Issuer (namespace scoped). If not specified, these values are retrieved from a ConfigMap, with a default value of `self-signed` and `ClusterIssuer`. The certificate is mounted into the container via a Secret so that it is automatically refreshed once the certificate is updated.

###### Integration with OpenShift Serverless
Deploy your serverless runtime component using a single toggle.  The Operator will convert all of its generated resources into [Knative](https://knative.dev) resources, allowing your pod to automatically scale to 0 when it is idle.

#### Integration with OpenShift's Topology UI
We set the corresponding labels to support OpenShift's Developer Topology UI, which allows you to visualize your entire set of deployments and how they are connected.
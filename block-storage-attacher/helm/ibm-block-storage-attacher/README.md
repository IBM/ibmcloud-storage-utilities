
# Helm chart to install IBM Cloud Block Storage Attacher

## Introduction
IBM Cloud Block Storage is persistent, high performance iSCSI storage that you can mount to apps that run in your Kubernetes cluster. To attach raw, unformatted, and unmounted block storage to a worker node in your cluster, install the IBM Cloud Block Storage Attacher by using a Helm chart.

For more information about IBM Cloud Block Storage, see [Getting started with Block Storage](https://cloud.ibm.com/docs/infrastructure/BlockStorage?topic=BlockStorage-GettingStarted#getting-started-with-block-storage).

## Chart Details
The IBM Cloud Block Storage Attacher Helm chart creates the following resources in your cluster:
- 1 Pod on every worker node in your cluster as part of a daemon set that includes the IBM Cloud Block Storage Attacher driver. The pod invokes the Linux service on the worker node to attach or detach a volume.
- RBAC roles for the driver of the IBM Cloud Block Storage Attacher to assign the required permissions within the cluster.
- 1 storage class that you later use to attach the block storage device to your worker node.

## Prerequisites
*   If you do not have one yet, [create a standard cluster](https://cloud.ibm.com/docs/containers?topic=containers-clusters#clusters_cli).
*   Kubernetes version 1.10 or later.
*   Helm version 3 or later.
*   The following CLIs and plugins.
    *  IBM Cloud CLI (`ibmcloud`)
    *  IBM Cloud Kubernetes Service plug-in (`ibmcloud ks`)
    *  IBM Cloud Container Registry plug-in (`ibmcloud cr`)
    *  Kubernetes (`kubectl`)
    *  Helm (`helm`)

## PodSecurityPolicy Requirements
The IBM Cloud Block Storage Attacher daemon set pod that includes the IBM Cloud Block Storage Attacher driver requires multiple `hostpath` volumes to install binaries. By default, IBM Cloud Kubernetes Service comes with the `ibm-privileged-psp` pod security policy that allows the IBM Cloud Block Storage Attacher daemon set to execute the required actions. If you use custom pod security policies, make sure that your pod security policy allows the IBM Cloud Block Storage Attacher pod to use `hostpath` volumes.

## Limitations
* The IBM Cloud Block Storage Attacher can run on an amd64 architecture only.

## Install the Helm version 3 client

1.  [Follow the instructions](https://cloud.ibm.com/docs/containers?topic=containers-helm#install_v3) to install the Helm version 3 client on your local machine.

2. Add the IBM Cloud Helm chart repository to the cluster where you want to use the IBM Cloud Block Storage Attacher plug-in.

3. Repeat the steps for all clusters where you want to install the IBM Cloud Block Storage Attacher.

## Installing the IBM Cloud Block Storage Attacher in your cluster
Now that the Helm server is up and running in your cluster, install the IBM Cloud Block Storage Attacher to attach the block storage on the nodes.

1. Add the IBM Cloud Helm repository `iks-charts` to your Helm instance and get the latest version of all Helm charts.
   <pre>helm repo add iks-charts https://icr.io/helm/iks-charts</pre>
   <pre>helm repo update</pre>   
2. Install the IBM Cloud Block Storage Attacher. When you install the attacher, a daemon set, RBAC roles, and pre-defined storage classes are created in your cluster.
   <pre>helm install ibm-block-storage-attacher iks-charts/ibm-block-storage-attacher</pre>

   Example output:
   ```
   NAME:   <helm_chart-name>
   LAST DEPLOYED: Wed Aug  1 14:55:15 2018
   NAMESPACE: default
   STATUS: DEPLOYED

   RESOURCES:
   ==> v1/StorageClass
   NAME                 TYPE
   ibmc-block-attacher  ibm.io/ibmc-blockattacher  

   ==> v1/ServiceAccount
   NAME                             SECRETS  AGE
   ibm-block-storage-attacher       1        2s

   ==> v1beta1/ClusterRole
   NAME                             AGE
   ibm-block-storage-attacher       2s

   ==> v1beta1/ClusterRoleBinding
   NAME                             AGE
   ibm-block-storage-attacher       2s

   ==> v1beta1/DaemonSet
   NAME                             DESIRED  CURRENT  READY  UP-TO-DATE  AVAILABLE  NODE-SELECTOR  AGE
   ibm-block-storage-attacher       1        1        0      1           0          <none>         2s
   ```

3. Verify that the IBM Cloud Block Storage Attacher is installed correctly.
   <pre>kubectl get pod -n kube-system -o wide | grep attacher</pre>

   Example output:
   ```
   ibm-block-storage-attacher-z7cv6           1/1       Running            0          19m
   ```
   The installation is successful when you see one or more `ibm-block-storage-attacher` pods. The number of `ibm-block-storage-attacher` pods equals the number of worker nodes in your cluster. All pods must be in a **Running** state.

4. Verify that the storage class for the IBM Cloud Block Storage Attacher is added to your cluster.
   <pre>kubectl get storageclasses | grep attacher</pre>

   Example output:
   ```
   ibmc-block-attacher        ibm.io/ibmc-blockattacher
   ```
5. Repeat the steps for all clusters where you want to attach block storage to your worker nodes.

## Custom configuration settings
The Helm chart has the following Values that can be overridden using the `helm install --set` parameter.

Example:
```
helm install <helm_chart_name> --set image.repository=icr.io/ibm/ibmcloud-block-storage-attacher iks-charts/ibm-block-storage-attacher (For Helm 3)
```
| Value                  | Description                      | Default                                    |
|------------------------|----------------------------------|--------------------------------------------|
| image.repository       | The image repository of attacher | icr.io/ibm/ibmcloud-block-storage-attacher |
| image.build            | The attacher driver build tag    |                                            |
| image.pullPolicy       | Image Pull Policy                | Always                                     |

## Updating the IBM Cloud Block Storage Attacher on your cluster
If you want to upgrade your existing IBM Cloud Block Storage Attacher chart to the latest version, you can update the Helm chart.

1. Find the installation name of your helm chart.

   <pre>helm ls | grep ibm-block-storage-attacher</pre>

   Example output:
   ```
   <helm_chart_name>	1       	Wed Aug  1 14:55:15 2018	DEPLOYED	ibm-block-storage-attacher-1.0.0	default
   ```

2. Upgrade the IBM Cloud Block Storage Attacher to latest.
   <pre>helm upgrade --force  &lt;helm_chart_name&gt; iks-charts/ibm-block-storage-attacher </pre>

## Removing the IBM Cloud Block Storage Attacher from your cluster
If you do not want to use IBM Cloud Block Storage for your cluster, you can uninstall the Helm chart.

1. Find the installation name of your Helm chart.

   <pre>helm ls | grep ibm-block-storage-attacher</pre>

   Example output:
   ```
   <helm_chart_name>	1       	Wed Aug  1 14:55:15 2018	DEPLOYED	ibm-block-storage-attacher-1.0.0	default
   ```

2. Delete the IBM Cloud Block Storage Attacher by removing the Helm chart.
   <pre>helm delete &lt;helm_chart_name&gt;</pre>

## What's next?
Now that you installed the IBM Cloud Block Storage Attacher, you can start to [automatically add block storage](https://cloud.ibm.com/docs/containers?topic=containers-utilities#attach_block) and [attach the block storage](https://cloud.ibm.com/docs/containers?topic=containers-utilities#automatic_block) to all your worker nodes.

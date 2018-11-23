
# Helm chart to install IBM Cloud Block Storage Attacher

## Introduction
IBM Cloud Block Storage is persistent, high performance iSCSI storage that you can mount to the apps that run in your cluster. To attach the block storage onto any desired node of your cluster, you must install the IBM Cloud Block Storage Attacher by using a helm chart. This helm chart also includes the creation of pre-defined storage class to define that the block storage device has to be attached on the node. 

For more information about IBM Cloud Block Storage, see [Getting started with Block Storage](https://console.bluemix.net/docs/infrastructure/BlockStorage/index.html#getting-started-with-block-storage).

## Chart Details
The deployment of this chart contains Storage class, Attacher driver DaemonSet and the RBAC roles for the attacher driver.

The Storage class deployment is required to enable the attacher driver to perform attach or detach of volumes.

The DaemonSet contains attacher driver which invokes the linux service on the worker node to attach or detach the volumes. Ensure it run on each node.

## Prerequisites
*   If you do not have one yet, [create a standard cluster](https://console.bluemix.net/docs/containers/cs_clusters.html#clusters_cli). 
*   Kubernetes version 1.10 or later.
*   Helm version 2.10 or later.
*   The following CLIs and plugins.
    *  IBM Cloud CLI (`ibmcloud`)
    *  IBM Cloud Kubernetes Service plug-in (`ibmcloud ks`)
    *  IBM Cloud Container Registry plug-in (`ibmcloud cr`)
    *  Kubernetes (`kubectl`)
    *  Helm (`helm`)

## Resources Required
ibm-block-storage-attacher uses Kubernetes v1.10 or higher.

## Installing the Chart
The helm installation consists of two parts, the helm client (helm) and the helm server (tiller). When helm is correctly set up on your cluster, you can install the IBM Cloud Block Storage Attacher and start attaching block storage on the node of your cluster.

### Installing the helm client on your local machine
Follow the instructions in the helm documentation to [install the helm client](https://docs.helm.sh/using_helm/#installing-helm) on your local machine. The helm client is needed to execute helm commands against your helm server in your cluster. 

### Installing the helm server on your cluster
When you set up the helm server in your cluster, you can use the helm chart provided in this repository to install the IBM Cloud Block Storage Attacher. 

1. Note the cluster name or ID where you want to install the IBM Cloud Block Storage Attacher.
   <pre>ibmcloud ks clusters</pre>
2. Set the cluster as the context for this session.
   - Get the command to set the environment variable and download the Kubernetes configuration files.
     <pre>ibmcloud ks cluster-config --cluster &lt;cluster_name_or_id&gt;</pre>
   - Copy and paste the command that is displayed in your terminal to set the KUBECONFIG environment variable.
     <pre>export KUBECONFIG=/Users/&lt;user_name&gt;/.bluemix/plugins/container-service/clusters/&lt;cluster_name&gt;/kube-config-prod-dal10-<cluster_name>.yml</pre>
3. Initialize helm to set up a helm server in your cluster.
   <pre>helm init</pre>
   
   Example output: 
   ```
   Creating /Users/ibm/.helm 
   Creating /Users/ibm/.helm/repository 
   Creating /Users/ibm/.helm/repository/cache 
   Creating /Users/ibm/.helm/repository/local 
   Creating /Users/ibm/.helm/plugins 
   Creating /Users/ibm/.helm/starters 
   Creating /Users/ibm/.helm/cache/archive 
   Creating /Users/ibm/.helm/repository/repositories.yaml 
   Adding stable repo with URL: https://kubernetes-charts.storage.googleapis.com 
   Adding local repo with URL: http://127.0.0.1:8879/charts 
   $HELM_HOME has been configured at /Users/ibm/.helm.

   Tiller (the Helm server-side component) has been installed into your Kubernetes Cluster.
   Happy Helming!
   ```

4. Repeat the steps for all clusters where you want to install the IBM Cloud Block Storage Attacher.
   
## Installing the IBM Cloud Block Storage Attacher on your cluster
Now that the helm server is up and running in your cluster, install the IBM Cloud Block Storage Attacher to attach the block storage on the nodes.

1. Clone the GitHub repository where the helm chart is stored.
   <pre>git clone https://github.com/IBM/ibmcloud-storage-utilities.git</pre>
                          OR
   <pre>git clone git@github.com:IBM/ibmcloud-storage-utilities.git</pre>
2. Navigate to the installation directory.
   <pre>cd ibmcloud-storage-utilities/block-storage-attacher/helm</pre>
3. Install the IBM Cloud Block Storage Attacher. Replace `<helm_chart_name>` with a name for your helm chart. When you install the attacher, pre-defined storage classes are added to your cluster.
   <pre>helm install ibm-block-storage-attacher/ --name &lt;helm_chart_name&gt;</pre>
   
   Example output: 
   ```
   NAME:   elevated-rodent
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

4. Verify that the attacher was installed correctly. 
   <pre>kubectl get pod -n kube-system -o wide | grep attacher</pre>
   
   Example output: 
   ```
   ibm-block-storage-attacher-z7cv6           1/1       Running            0          19m
   ```
   The installation is successful when you see one or more `ibm-block-storage-attacher` pods. The number of `ibm-block-storage-attacher` pods equals the number of worker nodes in your cluster. All pods must be in a **Running** state.
   
5. Verify that the storage classes for block storage attacher were added to your cluster.
   <pre>kubectl get storageclasses | grep attacher</pre>
   
   Example output: 
   ```
   ibmc-block-attacher        ibm.io/ibmc-blockattacher
   ```
6. Repeat the steps for all clusters where you want to attach block storage for your cluster nodes.

## Configuration
The helm chart has the following Values that can be overriden using the install --set parameter. For example:
```
helm install --set image.repository=registry.stage1.ng.bluemix.net/ibm/ibmcloud-block-storage-attacher ibm/ibm-block-storage-attacher
```
| Value                  | Description                             | Default                                                  |
|------------------------|-----------------------------------------|----------------------------------------------------------|
| image.repository       | The image repository of attacher        | registry.bluemix.net/ibm/ibmcloud-block-storage-attacher |
| image.build            | The attacher driver build tag           | latest                                                   |
| image.pullPolicy       | Image Pull Policy                       | Always                                                   |

## Updating the IBM Cloud Block Storage Attacher on your cluster
If you want to upgrade your existing IBM Cloud Block Storage Attacher chart to latest version, you can do it as below.

1. Find the installation name of your helm chart.

   <pre>helm ls | grep ibm-block-storage-attacher</pre>

   Example output:
   ```
   <helm_chart_name>	1       	Wed Aug  1 14:55:15 2018	DEPLOYED	ibm-block-storage-attacher-1.0	default
   ```

2. Upgrade the IBM Cloud Block Storage Attacher to latest.
   <pre>helm upgrade --force --recreate-pods &lt;helm_chart_name&gt; ibm-block-storage-attacher</pre>

## Removing the IBM Cloud Block Storage Attacher from your cluster
If you do not want to use IBM Cloud Block Storage for your cluster, you can uninstall the helm chart. 

1. Find the installation name of your helm chart. 

   <pre>helm ls | grep ibm-block-storage-attacher</pre>
   
   Example output: 
   ```
   <helm_chart_name>	1       	Wed Aug  1 14:55:15 2018	DEPLOYED	ibm-block-storage-attacher-1.0	default
   ```
   
2. Delete the IBM Cloud Block Storage Attacher by removing the helm chart. 
   <pre>helm delete &lt;helm_chart_name&gt; --purge</pre>

## Limitations
* This Chart can run only on amd64 architecture type.

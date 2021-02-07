# px-cleanup

The PX cleanup script does the following

- Removes Portworx from the cluster
- Removes the portworx helm release
- Removes the portworx IBM Cloud service from catalog


When uninstalling Portworx from cluster, There are  2  options:

Stop Portworx and remove the Kubernetes specs and completely wipe the data.
   you will be prompted for confirmation if yes then removes the volumes data also

2. Stop Portworx and remove the Kubernetes specs without wiping the data.

  Uninstalling or deleting the Portworx daemonset only removes the Portworx containers from the nodes. As the configurations files which Portworx use are persisted on the nodes the storage devices and the data volumes are still intact. These Portworx volumes can be used again if the Portworx containers are started with the same configuration files. (edited)


Prerequisites

*   Helm version 3.0 or later.


### Installing the Helm client on your local machine
Follow the instructions in the Helm documentation to [install the Helm client](https://docs.helm.sh/using_helm/#installing-helm) on your local machine. The Helm client is needed to execute Helm commands against your Helm server in your cluster.



Once the cluster config is set run the scirpt  


sudo ./px_cleanup.sh







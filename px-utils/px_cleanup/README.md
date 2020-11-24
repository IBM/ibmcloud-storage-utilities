# px-cleanup

The PX Cleanup script  works in the following way.

When uninstalling Portworx from cluster, There are  2  options:

Stop Portworx and remove the Kubernetes specs and completely wipe the data.
   you will be prompted for confirmation if yes then removes the volumes data also

2. Stop Portworx and remove the Kubernetes specs without wiping the data.

  Uninstalling or deleting the Portworx daemonset only removes the Portworx containers from the nodes. As the configurations files which Portworx use are persisted on the nodes the storage devices and the data volumes are still intact. These Portworx volumes can be used again if the Portworx containers are started with the same configuration files. (edited)


Prerequisites

1) Set the cluster config using following command

  export KUBECONFIG="Path to the config yaml file"


Once the cluster config is set run the scirpt  


sudo ./px_cleanup.sh







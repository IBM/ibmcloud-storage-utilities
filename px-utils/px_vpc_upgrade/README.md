# px-cleanup

The PX VPC Upgrade script  upgrade or replace the worker node sequentailly. It will accept the workerids or worker-pool as input

If the worker ids are provided then workers will be upgrades/replaced

if the worker pool is given as input then all the workers in the worker pool will be replaced/upgraded in sequence


Prerequisites

1) Set the cluster config using following command

  export KUBECONFIG="Path to the config yaml file"


Once the cluster config is set run the scirpt  


sudo ./vpc_upgrade_util.sh  mycluster  replace/upgrade  worker/worker-pool (workerid1 workerid2) / (worker-pool-id1 worker-pool-id2) ....

If the worker ids not provided then all the workers in the cluster will be replaced/upgraded 







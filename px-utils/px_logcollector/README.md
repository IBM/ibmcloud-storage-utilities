# px-debug
This script collects the follwing portworx logs from the cluster and place on the "/tmp/pxlogs/node-ips/" on the local machine

Prerequisites
1) Set the cluster config using following command

export KUBECONFIG="Path to the config yaml file"


Once the cluster config is set run the scirpt  with or with out arguments. 
If you run with out arguments it collects  logs from all the availble workere nodes (example below)


sudo ./PX_Log_Collect.sh 

In this case it creates the directories at "/tmp/pxlogs" with the ipaddress of all workers nodes as folder name  

if you want to run with arguments then example below

sudo ./PX_Log_Collect.sh   --workers worker1 worker2 worker3

In this case it creates the directories at "/tmp/pxlogs" with the ipaddresspassed as arguments as the folder name

The following logs will be availble
Journal log
Pxstatus
diag tar file
pxcluster status






# block-storage-attacher-e2e

## block-storage-attacher-e2e  on development environment (ex: devmex)

	***Get a 16.04 instance for the setup. Also, make sure you open a VPN connection to the environment from another session.***

1. Install go

	You need [go](https://golang.org/doc/install) in your path (see [here](development.md#go-versions) for supported versions), please make sure it is installed and in your ``$PATH``.
	
	```sh
	GO_VERSION=1.7.4
	curl -o go${GO_VERSION}.linux-amd64.tar.gz https://storage.googleapis.com/golang/go${GO_VERSION}.linux-amd64.tar.gz
	tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz
	export GOPATH=<Go Path location>
	export PATH=/usr/local/go/bin:$PATH
	```
	
2. Installing CF CLI

	```sh
	wget -q -O - https://packages.cloudfoundry.org/debian/cli.cloudfoundry.org.key | sudo apt-key add -
	echo "deb http://packages.cloudfoundry.org/debian stable main" | sudo tee /etc/apt/sources.list.d/cloudfoundry-cli.list
	
	apt-get update
	apt-get install cf-cli
	```
	
3. Installing Bluemix CLI

	```sh
	BLUEMIX_CLI_URL="http://ftp.icap.cdl.ibm.com/OERuntime/BluemixCLIs/CliProvider/bluemix-cli"
	export BM_CLI_LATEST=$(curl -sSL ${BLUEMIX_CLI_URL}/ | grep 'amd64' | grep 'tar.gz' | grep 'Bluemix_CLI_[0-9]' | tail -n 1 | sed -e 's/^.*href="//' -e 's/">.*//') \
	&& curl -s ${BLUEMIX_CLI_URL}/${BM_CLI_LATEST} | tar -xvz \
	&& cd Bluemix_CLI \
	&& ./install_bluemix_cli \
	&& bx config --check-version false \
	&& bx config --usage-stats-collect false \
	&& bx --version
	```
	
4. Install Bluemix containers and registry plugin

	```sh
	bx plugin repo-add stage https://plugins.stage1.ng.bluemix.net \
	&& bx plugin install cs -r stage \
	&& bx plugin install container-registry -r stage \
	&& bx plugin list
	```
	
5. Login to Bluemix CLI supplying your IBM ID and password. Select the appropriate organization (typically the same as your IBM ID). Note, selection of a space is not required.

	```sh
	Export the needed variables:
	
	export PVG_BX_USER=<>
	export PVG_BX_PWD=<>
	bx login -a https://api.stage1.ng.bluemix.net -u $PVG_BX_USER -p $PVG_BX_PWD -sso
	```

6. Login to the Container Service plugin in the Bluemix CLI. Note you will be asked to supply your IBM ID (email) and password once again. This is known issue (see below).

	```sh
	#export ARMADA_API_ENDPOINT=<API Endpoint>
	bx cs init --host $ARMADA_API_ENDPOINT
	```
  
7. To provision a paid (cruiser) cluster run the following. The paid cluster provisions worker nodes (VMs) into your account. The number of worker nodes is based on the integer supplied in the --workers parameter. The paid size currently is an hourly, public, 2 core, 4GB VSI. 

	```sh
	# Before running this script, export the needed variables
	#export PVG_SL_USERNAME=<SL account user name>
	#export PVG_SL_API_KEY=<SL API Key>
	#export PVG_CRUISER_PRIVATE_VLAN=<Private VLAN>
	#export PVG_CRUISER_PUBLIC_VLAN=<Public VLAN>
	#export FREE_DATACENTER=<Datacenter name: mex01>
	#export PVG_CLUSTER_CRUISER=testcluster
        export SL_USERNAME=$PVG_SL_USERNAME
        SL_API_KEY=$PVG_SL_API_KEY
	bx cs credentials-set --softlayer-username $PVG_SL_USERNAME --softlayer-api-key $PVG_SL_API_KEY
        bx sl init -u $PVG_SL_USERNAME -p $PVG_SL_API_KEY 
	# Create a cruiser
	```
	
8. To list your clusters run the following

	```
	bx cs clusters
	```
	Note: Wait for the cluster for getting the ready state
	
9. Set the kubeconfig

	```
	configfile=$(bx cs cluster-config $PVG_CLUSTER_CRUISER | grep ^export | cut -d '=' -f 2)
	export KUBECONFIG=$configfile
	```

10. Now, run the following commands to perform the e2e test execution.
	
	```
	cd $GOPATH/src/github.ibm.com/alchemy-containers/ibmcloud-storage-utilities/block-storage-attacher/tests/e2e
	export PVG_PHASE=armada-prestage
	sed -i "s/PVG_PHASE/"$PVG_PHASE"/g" common/constants.go
	export API_SERVER=$(kubectl config view | grep server | cut -f 2- -d ":" | tr -d " ")
	make KUBECONFIGPATH=$KUBECONFIG PVG_PHASE=$PVG_PHASE armada-portworx-e2e-test
	
	Note: Remove the "exit 0" to avoid session logouts from the above two scripts.
	```
	
7. Get the pod details of cluster

	```sh
	kubectl get pods -n kube-system
	```

8. Get the logs of stroage plugin pod

	```sh
	kubectl  logs <Storage Plugin Pod ID> -n kube-system
	```


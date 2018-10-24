# Create and Attach Remote Block Storage
`mkpvyaml` is intended to automate the provisioning of remote block storage
for Portworx running on IBM IKS.

Creation of remote block storage devices is driven by a configuration file
names `yamlgen.yaml` of the following form:

```
#
# Can only specify 'performance' OR 'endurance' and associated clause
#
cluster:  jeffpx-lon02   #  name of IKS cluster
region:   uk-south       #  cluster region
type:  endurance          #  performance | endurance
offering: storage_as_a_service   #  storage_as_a_service | enterprise | performance
# performance:
#    - iops:  40000          #   INTEGER between 100 and 1000 in multiples of 100
endurance:
  - tier:  2              #   [0.25|2|4|10]
size:  [ 20 ]             #  Number and size of disks (in GB).  
```

The `cluster` and `region` parameters correspond to the IKS cluster and region.
The `type` must be either `performance` or `endurance`,
with the following clause corresponding to one or the other.
Lastly the `size` parameter determines the number and sizes of the remote volumes
to be created per worker node.
If `size` were given as `size: [ 100, 200 ]`,
then 2 block devices would be created *per worker node*, sized 100GB and 200GB respectively.

The name of this input configuration file must be `yamlgen.yaml`.

The successful output result will be a Kubernetes spec file to generate the corresponding
physical volumes (e.g. `pv`) objects through the `block-attacher` script that
is detailed at [https://github.com/IBM/ibmcloud-storage-utilities/tree/master/block-storage-attacher](https://github.com/IBM/ibmcloud-storage-utilities/tree/master/block-storage-attacher)

## Build
To build this from scratch, clone this repo and run:
```
docker build -t mkpvyaml .
```

## Run
To run, Here's your docker command
Note the volume mapping
* Assumes that `yamlgen.yaml` is in the current working directory.
* Assumes that you have logged in to bluemix CLI
* Assumes that you have SL_API_KEY and SL_USERNAME

<br>SL_USERNAME is a SoftLayer user name. Go to [https://control.bluemix.net/account/user/profile](https://control.bluemix.net/account/user/profile), 
scroll down, and check API Username.
<br>SL_API_KEY is a SoftLayer API Key. Go to [https://control.bluemix.net/account/user/profile](https://control.bluemix.net/account/user/profile), 
scroll down, and check Authentication Key.
<br>BM_API_KEY – An API key for IBM Cloud services. If you don’t have one already, go to 
[https://console.bluemix.net/iam/#/apikeys](https://console.bluemix.net/iam/#/apikeys) and create a new key.

Caveats:  May also need to log in to SoftLayer via `bx sl init`

```
docker run --rm -v `pwd`:/data -v ~/.bluemix:/config -e SL_API_KEY=$SL_API_KEY -e SL_USERNAME=$SL_USERNAME mkpvyaml
```

The `/data` directory should map to whereever the configuration file `yamlgen.yaml` is located.
The output file will be generated in the same directory.

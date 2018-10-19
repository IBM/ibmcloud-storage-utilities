# Create and Attach Remote Block Storage

## Build
To build this from scratch, clone this repo and run:
```
docker build -t mkpvyaml .
```

## Run
To run, Here's your docker command
Note the volume mapping
* Assumes that you have a local `yamlgen.yaml` file
* Assumes that you have logged in to bluemix CLI
* Assumes that you have SL_API_KEY and SL_USERNAME

Caveats:  May also need to log in to SoftLayer via `bx sl init`

```
docker run --rm -v `pwd`:/data -v /root/.bluemix:/config -e KUBECONFIG=$KUBECONFIG -e SL_API_KEY=$SL_API_KEY -e SL_USERNAME=$SL_USERNAME mkpvyaml
```

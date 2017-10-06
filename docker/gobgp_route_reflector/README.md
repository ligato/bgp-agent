## GoBGP Route reflector docker

The purpose of this image is to be [BGP route reflector node](https://en.wikipedia.org/wiki/Route_reflector).
The GoBGP Route reflector docker image is [basic golang docker image](https://hub.docker.com/_/golang/) that contains [GoBGP](https://github.com/osrg/gobgp) binaries. The running GoBGP server running inside container simulates the route reflector behaviour. 
   

![arch](../../docs/imgs/dockerGoBGP.png "Docker GoBGP")

The Route reflector image is used in [tests](https://github.com/ligato/bgp-agent/tree/master/examples/gobgp_watch_plugin) that are running automatically or manual.
Predefined GoBGP configuration for docker container used in test can be found [here](https://github.com/ligato/bgp-agent/blob/master/docker/gobgp_route_reflector/usage_scripts/gobgp-client-in-host/gobgp.conf).

This folder also contains some helper scripts:
##### Build Docker
Script builds the docker image based on DockerFile.

```
./build-image-routereflector.sh
```

##### Pull Docker
Script pulls docker image from [Dockerhub](https://hub.docker.com/r/ligato/gobgp-for-rr/)

```
./pull-docker.sh
```

##### Push Benchmark Docker 
Script push created docker image to [Dockerhub](https://hub.docker.com/r/ligato/gobgp-for-rr/)

```
./push_docker.sh
```

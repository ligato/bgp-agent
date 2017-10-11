## Usage scripts for GoBGP Route reflector docker

The purpose of the scripts is to simplify work with GoBGP Route reflector docker image/container. These scripts are meant to run without `sudo` command, therefore the environment for docker must be [altered accordingly](https://docs.docker.com/engine/installation/linux/linux-postinstall/#manage-docker-as-a-non-root-user).

##### Create new docker network
Script creates new network that should be used by docker containers. The purpose of the network is to have more control over networking of docker containers. Default docker network i.e. doesn't allow to assing statis IP addresses to starting containers. 

```
./create-ligato-network-for-docker.sh
```

##### Remove previously created docker network
Script removes previously created docker network. It can be used when network is not needed anymore, i.e. in clean up.

```
./remove-ligato-network-for-docker.sh
```

##### Start GoBGP Route reflector docker container
Script start route reflector docker container. It is about convenience not to remember exact label and version of docker image.

```
./start-routereflector.sh
```

##### Stop GoBGP Route reflector docker container
Script stops route reflector docker container. It is about convenience not to remember exact label of docker container.

```
./stop-routereflector.sh
```

##### Connection to GoBGP Route reflector docker container
Script connects to the linux terminal inside the route reflector docker container.
```
./connect-to-routereflector.sh
```

##### Add data/path to route reflector 
Script adds route to GoBGP inside the Route reflector docker container. This route is then automatically learned by others using BGP protocol. Script can serve as input trigger for examples. 

```
./addPath.sh
```
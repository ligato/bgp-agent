#!/usr/bin/env bash
RR_IP="172.18.0.2"
ROOT="docker/gobgp_route_reflector/usage_scripts/"
NETWORK="ligato-bgp-network"
DOCKER_NAME="gobgp-for-rr"
DOCKER_IMAGE="ligato/gobgp-for-rr:v1.24"

docker run -d -v `pwd`/$ROOT$1:/etc/gobgp:rw --net $NETWORK --ip $RR_IP --name $DOCKER_NAME $DOCKER_IMAGE
#!/usr/bin/env bash
RR_IP="172.18.0.2"

docker run --name rr --net ligato-bgp-network --ip $RR_IP -e GOBGP_CONFIG=$1 -it routereflector
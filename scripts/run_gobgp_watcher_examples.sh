#!/usr/bin/env bash
source ./scripts/testOuput.sh
exitCode=0

#####Setup Docker Network ##################################################
./docker/gobgp_route_reflector/create-ligato-network-for-docker.sh
#####Setup Docker ##################################################
./docker/gobgp_for_rr/pull-docker.sh

#####Run Docker with GoBgp##################################################

## Create Docker with GoBGP Config
./docker/gobgp_route_reflector/start-routereflector.sh gobgp-client-in-host
sleep 2

## Advertize Path
./docker/gobgp_route_reflector/addPath.sh &
sleep 2

#Run example app
expected=("Agent received path &{65001 101.0.0.0/24 101.0.10.1}
")

testOutput ./examples/gobgp_watch_plugin/gobgp_watch_plugin "${expected}"

## Stop and remove docker
./docker/gobgp_route_reflector/stop-routereflector.sh

#####Remove Docker Network ##################################################
./docker/gobgp_route_reflector/remove-ligato-network-for-docker.sh
##########################################################################
exit ${exitCode}
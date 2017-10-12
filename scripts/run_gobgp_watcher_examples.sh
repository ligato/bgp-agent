#!/usr/bin/env bash
source ./scripts/testOuput.sh
exitCode=0

echo "## setup infrastructure"
#####Setup Docker Network ##################################################
echo "### setup docker network"
./docker/gobgp_route_reflector/usage_scripts/create-ligato-network-for-docker.sh
echo "done"
#####Setup Docker ##################################################
echo "### pull docker image (route reflector)"
./docker/gobgp_route_reflector/pull-docker.sh
echo "done"

echo ""
echo "## running examples"
#####Run Docker with GoBgp##################################################
echo "### running gobgp_watch_plugin example"
## Create Docker with GoBGP Config
echo "#### starting route reflector docker container"
./docker/gobgp_route_reflector/usage_scripts/start-routereflector.sh gobgp-client-in-host
sleep 2
echo "done"

## Advertize Path
echo "#### advertizing path to route reflector docker container"
./docker/gobgp_route_reflector/usage_scripts/addPath.sh &
sleep 2
echo "done"

#Run example app
echo "#### running go example (gobgp plugin,example plugin)"
expected=("Agent received path &{65001 101.0.0.0/24 101.0.10.1}
")

./examples/gobgp_watch_plugin/gobgp_watch_plugin &> log &
sleep 20
echo "$(less log)"
echo "#### validating Go example output"
testOutput "$(less log)" "${expected}"
echo "done"

echo ""
echo "## cleanup"
## Stop and remove docker
echo "### stop and remove docker container"
./docker/gobgp_route_reflector/usage_scripts/stop-routereflector.sh
echo "done"

#####Remove Docker Network ##################################################
echo "### remove docker network"
./docker/gobgp_route_reflector/usage_scripts/remove-ligato-network-for-docker.sh
##########################################################################
echo "done"
exit ${exitCode}
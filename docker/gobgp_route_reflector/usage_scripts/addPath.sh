#!/usr/bin/env bash

PREFIX="101.0.0.0/24"
NEXTHOP="101.0.10.1"
## Advertize path
echo "##### prefix "$PREFIX
echo "##### nexthop "$NEXTHOP

(
echo "gobgp global rib add -a ipv4 "$PREFIX" nexthop "$NEXTHOP
) | docker exec -i gobgp-for-rr bash
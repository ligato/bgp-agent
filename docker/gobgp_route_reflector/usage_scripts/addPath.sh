#!/usr/bin/env bash

## Advertize path
sudo docker exec -i gobgp-for-rr bash <<< "gobgp global rib add -a ipv4 101.0.0.0/24 nexthop 101.0.10.1"
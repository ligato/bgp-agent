#!/usr/bin/env bash
docker network create --subnet 172.18.0.0/24 --ip-range 172.18.0.255/25 --gateway 172.18.0.254 ligato-bgp-network

#static ip address for docker vm instances:
# 172.18.0.1 client for Route Reflector (e.g. BGP-VPP-Agent)
# 172.18.0.2 RouteReflector
# 172.18.0.3 reserved for IBGP Client on Benchmark test
# 172.18.0.254 host machine(=gateway)

#note: static ip addresses are deliberately outside of ip-range so that they can't be assigned automatically to any
# other docker vm instances using this network
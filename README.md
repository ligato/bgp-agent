# Ligato BGP Agent
[![Build Status](https://travis-ci.org/ligato/bgp-agent.svg?branch=master)](https://travis-ci.org/ligato/bgp-agent)
[![Coverage Status](https://coveralls.io/repos/github/ligato/bgp-agent/badge.svg?branch=master)](https://coveralls.io/github/ligato/bgp-agent?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/ligato/bgp-agent)](https://goreportcard.com/report/github.com/ligato/bgp-agent)
[![GoDoc](https://godoc.org/github.com/ligato/bgp-agent?status.svg)](https://godoc.org/github.com/ligato/bgp-agent)
[![GitHub license](https://img.shields.io/badge/license-Apache%20license%202.0-blue.svg)](https://github.com/ligato/bgp-agent/blob/master/LICENSE)

The `Ligato BGP Agent` is a BGP information provider. It provides BGP information in an unified format to allow to retrieve BGP information from different sources(different BGP frameworks, eg: GoBGP, ExaBGP) and supporting multiple extensions (different AFI/SAFIS ex: IPV4 / Unicast, IPV6 / Unicast).
## Architecture

The architecture of the `Ligato BGP Agent` is shown in the following figure.

![arch](docs/imgs/bgpagent.png "High Level Architecture of BGP Agent")

`Ligato BGP Agent` is set of `Ligato CN-Infra Plugin` implementations. Purpose of each plugin is to forward retrieved BGP information to clients(registered watchers). 

Every plugin has its source of BGP information (GoBGP,ExaBGP, Quagga). Communication with the source is vendor specific and therefore also the retrieved BGP information is usually vendor specific. The BGP information is translated into unified format and forwarded to clients. If the source of the BGP information supports listening for updates, plugin forwards to its clients also informations from updates.

Clients can register directly to plugins, choosing what information they want to consume. Each plugin can expose different set of BGP informations depending at capabilities of their source(GoBGP,ExaBGP,...) or plugin's client target(different plugins for different AFI/SAFI). But each type of BGP information, no matter from which plugin it came, has the same unified format. 

Plugins can be clients of other plugins too. This means that the architecture is quite flexible for the future usage. Different plugins can provide different types of BGP information/from different sources and can be kept separate as building stones for (hierarchy of) aggregator plugins. The aggregator plugin uses other plugins to retrieve needed information for its own registered clients. The aggregator plugins can for example provide information for one AFI/SAFI across multiple sources. This can be usefull if one source can provide all needed information or such source exists but can't be used for whatever reason. There are many possibilities how to combine plugins together to satisfy specific use cases.   

Currently, only [GoBGP plugin](bgp/gobgp/README.md) that exposes IPv4 reachable routes is available. ExaBGP,Quagga plugins are not implemented.

## Quickstart
For a quick start with the BGP Agent, you can use makefile and start examples
```
make run-examples
```
The command pulls needed docker images from [Dockerhub](https://hub.docker.com/r/ligato/gobgp-for-rr/), 
setups networking, builds the examples, runs them, validates their output and cleans after them.

You can check in the command output the most basic test, the [gobgp_watch_plugin](https://github.com/fgschwan/bgp-agent/tree/master/examples/gobgp_watch_plugin).

## Documentation
GoDoc can be browsed [online](https://godoc.org/github.com/ligato/bgp-agent).

## Contribution:
If you are interested in contributing, please see the [Ligato CN-Infra contribution guidelines](https://github.com/ligato/cn-infra/blob/master/CONTRIBUTING.md). The `Ligato BGP Agent` follows the same contribution guidlines as the [Ligato CN-Infra](https://github.com/ligato/cn-infra).

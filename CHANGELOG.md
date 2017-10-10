# Release v1.0.0 (2017-10-10)

## Initial release

Implemented first and basic plugin that serves as BGP information provider. It uses [GoBGP](https://github.com/osrg/gobgp) library as 
source of the BGP information and it provides only the reachable IPv4 routes. Further information about this implementation can be found [here](https://github.com/ligato/bgp-agent/tree/master/bgp/gobgp).

Release also contains [basic example of usage](https://github.com/ligato/bgp-agent/tree/master/examples/gobgp_watch_plugin) of this plugin implementation.

## Known Issues
Use cases regarding dropping connection and reconnection of the Route reflector(BGP node used that plugin uses for getting BGP information) are not taken into account. 

## Known Limitiations
The BGP information flow is one directional. We can not advertise the BGP information to the BGP network nodes.
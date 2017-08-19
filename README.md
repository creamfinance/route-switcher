# Route Switcher

Route switcher is a small go utility that allows to monitor multiple interfaces for their internet availability.

It automatically adds and removes a default route.

At the moment the a ping is sent every second to the specified targets.
If within an interval of 30 seconds more than 90% are not received back, the interface is marked as bad.
If the interface is marked bad 3 consecutive times, it's removed from the "good interfaces" list and removed from the route.

There's no limit on how many targets or interfaces can be managed.

Multiple active and up interfaces means that a load balancing route is established.

For correct multi interface functionality 2 things need to be done:

add ip rule for each source ip:

```
ip rule add from {first_ip} lookup 100
ip rule add from {second_ip} lookup 200
ip rule add from {third_ip} lookup 300
```

add a default route to each of the tables:

```
ip r add default via {first_ip_gateway} table 100
ip r add default via {second_ip_gateway} table 200
ip r add default via {third_ip_gateway} table 300
```

These tables and rules are not modified by the program, only the table set via the --table parameter is modified.
By default this is the main table.

## Starting the application

./route-switcher --external-interfaces eth0-192.168.0.254,eth0.4-10.0.0.254 --ping-targets 8.8.8.8,8.8.4.4

which will ping the corresponding google ips for each interface, and use the passed ips as gateways.
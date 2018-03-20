sysctl net.ipv4.conf.all.rp_filter=0
sysctl net.ipv4.conf.default.rp_filter=0
ip l add link eth0 name eth0.100 type macvlan mode bridge
ip l add link eth0 name eth0.101 type macvlan mode bridge
ip a add 172.17.0.20/16 dev eth0.100
ip a add 172.17.0.21/16 dev eth0.101
ip l set eth0.100 up
ip l set eth0.101 up

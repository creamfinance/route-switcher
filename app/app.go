package app

import (
	"errors"
	"fmt"
	"strings"
	"net"
	"sync"
	"os"
	"os/signal"
	"syscall"
	"time"
	"reflect"

	"github.com/golang/glog"
	"github.com/creamfinance/route-switcher/app/options"

	"github.com/tatsushid/go-fastping"
	"github.com/vishvananda/netlink"
)

type RouteSwitcher struct {
	config 		*options.RouteSwitcherConfig
	statistics  []*InterfaceStatistic
	lastRoute   *netlink.Route
}

type InterfaceStatistic struct {
	link 	 	netlink.Link
	gateway     *net.IP
	sent 		int
	received	int
	badCounter  int
}

func NewRouteSwitcher (config *options.RouteSwitcherConfig) (*RouteSwitcher, error) {

	if config.ExternalInterfaces == "" {
		return nil, errors.New("Require at least two external interfaces, --external-interfaces need to be defined")
	}

	if config.PingTargets == "" {
		return nil, errors.New("Require at least one ping target, --ping-targets need to be filled")
	}

	rs := &RouteSwitcher{}

	rs.config = config

	return rs, nil
}

func (rs *RouteSwitcher) Run() error {
	var wg sync.WaitGroup

	fmt.Printf("Running Route Switcher\n")

	interfaces := strings.Split(rs.config.ExternalInterfaces, ",")
	ping_targets := strings.Split(rs.config.PingTargets, ",")

	links := make([]*InterfaceStatistic, 0)
	targets := make([]net.IP, 0)

	// Check if the passed interfaces are really there
	for _, ife := range interfaces {
		parts := strings.Split(ife, "-")

		if len(parts) != 2 {
			return errors.New("External interface doesn't have 2 pairs e.g. eth0-10.21.0.254")
		}

		rife, err := netlink.LinkByName(parts[0])

		if err != nil && err.Error() == "Link not found" {
			return errors.New("Link " + ife + " not found!")
		} else if err != nil {
			return err
		}

		gip := net.ParseIP(parts[1])

		if gip == nil {
			return errors.New("Invalid gateway ip: " + parts[1])
		}

		stat := InterfaceStatistic{
			link: rife,
			gateway: &gip,
			received: 0,
			sent: 0,
		}

		links = append(links, &stat)
	}

	rs.statistics = links

	// Parse all the ips
	for _, ip := range ping_targets {
		ipp := net.ParseIP(ip)

		if ipp == nil {
			return errors.New("IP " + ip + " is invalid")
		}

		targets = append(targets, ipp)
	}

	stopChan := make(chan struct{})


	// Run go subroutine for every interface
	for _, link := range links {
		wg.Add(1)
		go rs.MonitorInterface(link, targets, stopChan, &wg)
	}

	wg.Add(1)
	go rs.SwitchInterfaces(stopChan, &wg)


	// Handle SIGINT / SIGTERM
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	glog.Infof("Shutting down interface monitoring")

	// TODO stop controllers

	close(stopChan)

	wg.Wait()

	return nil
}

func (rs *RouteSwitcher) SwitchInterfaces(stopCh <-chan struct{}, wg *sync.WaitGroup) error {
	t := time.NewTicker(time.Second * 10)
	defer t.Stop()
	defer wg.Done()

	for {
        select {
        case <-stopCh:
            glog.Infof("Shutting down router switcher")
            return nil
        default:
        }

        goodInterfaces := make([]*InterfaceStatistic, 0)

        for _, stat := range rs.statistics {
        	glog.Infof("Statistics %s %d / %d ", stat.link.Attrs().Name, stat.received, stat.sent)

        	if float64(stat.received) / float64(stat.sent) > 0.9 {
        		glog.Infof("    Interface is good!")
        		stat.badCounter = 0
        	} else {
        		glog.Infof("    Interface is bad!")
        		stat.badCounter += 1
        	}

        	stat.sent = 0
        	stat.received = 0

        	if stat.badCounter <= 3 {
        		goodInterfaces = append(goodInterfaces, stat)
        	}
        }


        _, allNet, _ := net.ParseCIDR("0.0.0.0/0")
        multipath := make([]*netlink.NexthopInfo, 0)

        for _, stat := range goodInterfaces {
        	multipath = append(multipath, &netlink.NexthopInfo{
        		LinkIndex: stat.link.Attrs().Index,
        		Gw: *stat.gateway,
        		Hops: 0,
        	})

        	if rs.config.RoutePreference == "single" {
        		break
        	}
        }

        route := netlink.Route{
        	Dst: allNet,
        	MultiPath: multipath,
        	Table: rs.config.Table,
        }

        if rs.lastRoute != nil && !reflect.DeepEqual(route, rs.lastRoute) {
        	// Routes are not equal, remove the old one
        	err := netlink.RouteDel(rs.lastRoute);

        	if err != nil {
        		glog.Errorf("[Route Delete] %v", err)
        	}
        }

        err := netlink.RouteAdd(&route);

		if err != nil {
    		glog.Errorf("[Route Add] %v", err)
    	}

        rs.lastRoute = &route


	    /*
        if len(goodInterfaces) == 0 {
        	// No good interface, what do we do?
        	glog.Infof("No healty interface available - what do we do?")

        	// remove added routes
    	} else if len(goodInterfaces) == 1 {
    		glog.Infof("Only one healty interface available - put all traffic there")

    		// ip route add default via {gateway_ip} dev {link_name}
    	} else {
    		glog.Infof("Multiple healty interfaces available - load balancing")

    		// ip route add default nexthop via {gateway_ip} dev {link_name} weight 1 nexthop {gateway_ip2} dev {link_name2} weight 1
    	}
    	*/

    	select {
    	case <-stopCh:
	        glog.Infof("Shutting down router switcher")
	        return nil
        case <-t.C:
        }
    }
}

func (rs *RouteSwitcher) MonitorInterface(statistic *InterfaceStatistic, targets []net.IP, stopCh <-chan struct{}, wg *sync.WaitGroup) error {
    /*t := time.NewTicker(time.Second * 5)
    defer t.Stop()*/
    defer wg.Done()

   	// Maybe make interface monitor dynamically select the ip, all the time

    link := statistic.link

    ips, _ := netlink.AddrList(link, netlink.FAMILY_V4)

    glog.Infof("Monitoring interface %v with IP %v", link, ips)

    for _, ip := range ips {
    	glog.Infof("IP: %v Mask: %v", ip.IP, ip.Mask)
    }

    if len(ips) == 0 {
    	glog.Infof("No available IPs to check.")
    	return errors.New("No available IPs to check.")
    }

    p := fastping.NewPinger()
    p.MaxRTT = time.Second
    p.Source(ips[0].IP.String())

    items := 0
    rcvd := 0

    for _, ip := range targets {
    	p.AddIP(ip.String())
    	items += 1
    }

	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		// fmt.Printf("IP Addr: %s receive, RTT: %v\n", addr.String(), rtt)
		rcvd += 1
	}

	p.OnIdle = func() {
		statistic.received += rcvd
		statistic.sent += items

		rcvd = 0

	}

	p.RunLoop()

    <-stopCh
    glog.Infof("Shutting down interface monitor")
    p.Stop()

    return nil
}






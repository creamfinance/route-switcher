package options

import (
	"github.com/spf13/pflag"
)

type RouteSwitcherConfig struct {
	HelpRequested 		bool
	ExternalInterfaces	string
	PingTargets			string
	Table               int
	RoutePreference     string // multi or single, if multi hop routes are allowed or not
}

func NewRouteSwitcherConfig() *RouteSwitcherConfig {
	return &RouteSwitcherConfig{
		Table: 		254,
		RoutePreference: "single",
	}
}

func (s *RouteSwitcherConfig) AddFlags (fs *pflag.FlagSet) {
    fs.BoolVarP(&s.HelpRequested, "help", "h", false, "Print usage information.")
 	fs.StringVar(&s.ExternalInterfaces, "external-interfaces", s.ExternalInterfaces, "External interfaces which should be switched based on availability - for example eth0-91.208.133.2")
 	fs.StringVar(&s.PingTargets, "ping-targets", s.PingTargets, "Targets which will be used to determine if the interface is available.")
 	fs.IntVar(&s.Table, "table", s.Table, "Target table for the route manipulation.")
 	fs.StringVar(&s.RoutePreference, "route-preference", s.RoutePreference, "Defines wether next multi hop routing is used - (multi|single) - in single mode the order in external-interfaces gives the preference")
}
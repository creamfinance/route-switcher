package options

import (
	"github.com/spf13/pflag"
)

type RouteSwitcherConfig struct {
	HelpRequested 		bool
	ExternalInterfaces	string
	PingTargets			string
	Table               int
}

func NewRouteSwitcherConfig() *RouteSwitcherConfig {
	return &RouteSwitcherConfig{
		Table: 		254,
	}
}

func (s *RouteSwitcherConfig) AddFlags (fs *pflag.FlagSet) {
    fs.BoolVarP(&s.HelpRequested, "help", "h", false, "Print usage information.")
 	fs.StringVar(&s.ExternalInterfaces, "external-interfaces", s.ExternalInterfaces, "External interfaces which should be switched based on availability - for example eth0-91.208.133.2")
 	fs.StringVar(&s.PingTargets, "ping-targets", s.PingTargets, "Targets which will be used to determine if the interface is available.")
 	fs.IntVar(&s.Table, "table", s.Table, "Target table for the route manipulation.")
}
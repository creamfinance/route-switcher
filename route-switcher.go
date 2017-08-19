package main

import (
	"fmt"
	"os"

	"flag"

	"github.com/creamfinance/route-switcher/app"
	"github.com/creamfinance/route-switcher/app/options"
	"github.com/spf13/pflag"
)

func main() {
	// Parse Command Line Args
	config := options.NewRouteSwitcherConfig()
	config.AddFlags(pflag.CommandLine)
	pflag.Parse()


	// Workaround for
	flag.CommandLine.Parse([]string{})
	flag.Set("logtostderr", "true")

	if config.HelpRequested {
		pflag.Usage()
		os.Exit(0)
	}

	rs, err := app.NewRouteSwitcher(config)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse route-switcher config: %v\n", err)
		os.Exit(1)
	}

	err = rs.Run()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to run route-switcher: %v\n", err);
		os.Exit(1);
	}
}
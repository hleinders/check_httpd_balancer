/*
	Nagios-Plugin to check the Apache Balancer Status. To use this plugin it is
	mandatory that the last digit of the ip address corresponds to the last digit
	of the jvmRoute.
	The Apache configuration should read like this:

	Worker URL              Route      RouteRedir  Factor    Set    Status   Elected  To    From
	ajp://192.168.0.1:8009  content01              1         0      Init Ok  7959     13K   424M
	ajp://192.168.0.2:8009  content02              1         0      Init Ok  7958     8.0M  426M
	                ^               ^
    The markers explain the mapping.
*/

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	appVersion = "0.4"
	author     = "Harald Leinders (2015-05-29) / harald@leinders.de"
)

// Nagios return codes
const (
	OK         = 0
	ErrWarn    = 1
	ErrCrit    = 2
	ErrUnknown = 3
)

// FlagType is a compound type for command line flags
type FlagType struct {
	Verbose, Debug, DryRun    bool
	Version, UseSSL           bool
	Hostname, IPAddress, Port string
	TimeOut, URL              string
	Warning, Critical         int
	User, Password            string
	WorkerMap                 string
}

// PoolWorker represents a balancer worker
type PoolWorker struct {
	Type, Address, Route, Status string
}

// BalancerPool represents an apache mod_proxy balancer pool
type BalancerPool struct {
	Name          string
	StickySession string
	StatusOK      bool
	WorkersOK     int
	WorkersCount  int
	Workers       []PoolWorker
}

func (p BalancerPool) String() string {
	var mbs []string
	for _, m := range p.Workers {
		mbs = append(mbs, m.Address)
	}
	return fmt.Sprintf("Name: %s (Status OK: %t, Workers: %d/%d [%s])",
		p.Name, p.StatusOK, p.WorkersOK, p.WorkersCount, strings.Join(mbs, ", "))
}

// WorkerMapping is a helper type for the mapping of worker address to jvmRoute value
type WorkerMapping map[string]string

// Helper functions
func check(e error) {
	if e != nil {
		fmt.Println(e)
		os.Exit(ErrUnknown)
	}
}

func version() {
	fmt.Fprintf(os.Stderr, "Plugin:   %s\n", filepath.Base(os.Args[0]))
	fmt.Fprintf(os.Stderr, "Version:  %s\n", appVersion)
	fmt.Fprintf(os.Stderr, "Author:   %s\n", author)
	os.Exit(0)
}

func usage() {
	fmt.Fprintf(os.Stderr, "Version: %s\n", appVersion)
	fmt.Fprintf(os.Stderr, "Usage:   %s [-h] [options] -H Hostname -M Mapping -u URL\n", filepath.Base(os.Args[0]))
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	flag.PrintDefaults()
}

func main() {
	var flags FlagType
	var poolList []BalancerPool
	var status string
	var rcode int

	// Command line parsing
	// Bools
	flag.BoolVar(&flags.Debug, "d", false, "Debug mode")
	flag.BoolVar(&flags.Verbose, "v", false, "Verbose mode")
	flag.BoolVar(&flags.DryRun, "n", false, "Dry run")
	flag.BoolVar(&flags.Version, "V", false, "Show version")
	flag.BoolVar(&flags.UseSSL, "S", false, "Connect via SSL. Port defaults to 443")

	// ArgOpts
	flag.IntVar(&flags.Warning, "w", 50, "Warning threshold for offline workers (in %)")
	flag.IntVar(&flags.Critical, "c", 75, "Critical threshold for offline workers (in %)")

	flag.StringVar(&flags.Hostname, "H", "localhost", "Host name")
	flag.StringVar(&flags.IPAddress, "I", "127.0.0.1", "Host ip address (not implemented yet)")
	flag.StringVar(&flags.URL, "u", "/balancer-manager", "URL to check")
	flag.StringVar(&flags.Port, "p", "", "TCP port")
	flag.StringVar(&flags.User, "l", "", "Basic Auth: user")
	flag.StringVar(&flags.Password, "a", "", "Basic Auth: password")
	flag.StringVar(&flags.WorkerMap, "M", "192.168.0.1:01 192.168.0.2:02", "List of worker mappings (IP):(jvmRoute-suffix)")

	flag.Usage = usage

	flag.Parse()

	if len(os.Args) < 2 {
		// no args
		usage()
		os.Exit(OK)
	}

	if flags.Port != "" {
		flags.Hostname = fmt.Sprintf("%s:%s", flags.Hostname, flags.Port)
	}

	if flags.Version {
		version()
	}

	// get status page content
	content, err := GetContent(flags)
	check(err)

	// parse content for balancer pools
	poolList, err = ParseContent(flags, content)
	check(err)

	// check pools
	status, rcode = CheckPools(flags, poolList)

	if flags.Debug {
		fmt.Fprintf(os.Stderr, "\nPools found: %d\n\n", len(poolList))
	}

	fmt.Println(status)
	os.Exit(rcode)
}

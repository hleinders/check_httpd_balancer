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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	flag "github.com/spf13/pflag"
)

const (
	appVersion = "1.1"
	author     = "Harald Leinders (2021-08-06) / harald@leinders.de"
)

// Nagios return codes
const (
	OK = iota
	ErrWarn
	ErrCrit
	ErrUnknown
)

var agentString = "Golang: Apache Balancer Check (" + appVersion + ")"

// flagType is a compound type for command line flags
type flagType struct {
	Help, Verbose, Debug, DryRun bool
	Version, UseSSL              bool
	FullStatus                   bool
	Hostname, IPAddress, Port    string
	TimeOut, URL, Agent          string
	Warning, Critical            int
	User, Password               string
	WorkerMap                    string
	ConfigFile                   string
}

// Update is a function to populate the flags from a config file
func (f *flagType) Update(c configType) {
	f.Port = choice(c.Port, f.Port, "")
	f.Hostname = choice(c.Host, f.Hostname, "127.0.0.1")
	f.URL = choice(c.URL, f.URL, "/balancer-manager")
	f.UseSSL = f.UseSSL || c.UseSSL

	if len(c.WorkerMap) > 0 && f.WorkerMap == "" {
		f.WorkerMap = strings.Join(c.WorkerMap, " ")
	}
}

// PoolWorker represents a balancer worker
type PoolWorker struct {
	Type, Address, Route, Status string
}

// BalancerPool represents an apache mod_proxy balancer pool
type BalancerPool struct {
	Name          string
	Nonce         string
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
type WorkerMapping map[string][]string

// Helper functions
func choice(a, b, c string) string {
	if a != "" {
		return a
	}
	if b != "" {
		return b
	}
	return c
}

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
}

func usage() {
	fmt.Fprintf(os.Stderr, "Version: %s\n", appVersion)
	fmt.Fprintf(os.Stderr, "Usage:   %s [-h] [options] -H Hostname -M Mapping -u URL\n", filepath.Base(os.Args[0]))
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	flag.PrintDefaults()
}

func main() {
	var flags flagType
	var config configType
	var poolList []BalancerPool
	var status string
	var rcode int
	var err error

	// Command line parsing
	// Iniitial flags settings
	flag.Usage = usage
	flag.CommandLine.SortFlags = false

	// Bools
	flag.BoolVarP(&flags.Help, "help", "h", false, "Show help")
	flag.BoolVarP(&flags.Debug, "debug", "d", false, "Debug mode")
	flag.BoolVarP(&flags.Verbose, "verbose", "v", false, "Verbose mode")
	flag.BoolVarP(&flags.DryRun, "dry-run", "n", false, "Dry run")
	flag.BoolVarP(&flags.Version, "version", "V", false, "Show version")
	flag.BoolVarP(&flags.FullStatus, "full", "F", false, "Show full balancer status")
	flag.BoolVarP(&flags.UseSSL, "ssl", "S", false, "Connect via SSL. Port defaults to 443")

	// ArgOpts
	flag.StringVarP(&flags.Agent, "agent", "A", agentString, "user agent")
	flag.IntVarP(&flags.Warning, "warning", "w", 50, "Warning `threshold` for offline workers (in %)")
	flag.IntVarP(&flags.Critical, "critical", "c", 75, "Critical `threshold` for offline workers (in %)")

	flag.StringVarP(&flags.ConfigFile, "config", "C", "", "Read settings from config file")
	flag.StringVarP(&flags.Hostname, "host", "H", "localhost", "`Host` name")
	flag.StringVarP(&flags.IPAddress, "ip", "I", "", "Host ip `address` (set 'host' flag to select a vhost)")
	flag.StringVarP(&flags.URL, "url", "U", "", "`URL` to check (default: /balancer-manager)")
	flag.StringVarP(&flags.Port, "port", "P", "", "TCP `port`")
	flag.StringVarP(&flags.User, "user", "u", "", "Basic Auth: `user`")
	flag.StringVarP(&flags.Password, "pass", "p", "", "Basic Auth: `password`")
	flag.StringVarP(&flags.WorkerMap, "map", "M", "", "List of worker `mappings` (IP):(jvmRoute-suffix)")

	flag.CommandLine.MarkHidden("debug")
	flag.Parse()

	if flags.Help {
		flag.Usage()
		os.Exit(OK)
	}

	if flags.Version {
		version()
		os.Exit(OK)
	}

	// no args
	if len(os.Args) < 2 {
		usage()
		os.Exit(OK)
	}

	// Read config file, if any
	if flags.ConfigFile != "" {
		config, err = readConfig(flags)
		check(err)
		if flags.Debug {
			fmt.Println("Config read:")
			fmt.Println(prettyPrintJSON(config))
		}
	}

	// update flags from config file
	flags.Update(config)

	if flags.Debug {
		fmt.Printf("Flags found:\n%+v\n", flags)
	}

	// get status page content
	content, err := GetContent(flags)
	check(err)

	// parse content for balancer pools
	poolList, err = parseContent(flags, content)
	check(err)

	// check pools
	status, rcode = checkPools(flags, poolList)

	if flags.Debug || flags.FullStatus {
		fmt.Fprintf(os.Stderr, "\nPools found: %d\n\n", len(poolList))
	}

	fmt.Println(status)
	os.Exit(rcode)
}

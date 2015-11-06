package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// newWorkerMapping is used to create a new mapping from the command line parameter (-M)
func newWorkerMapping(mapList string) WorkerMapping {
	wmap := make(WorkerMapping)

	reSplitString := regexp.MustCompile(`(\S+):(\S+)`)

	for _, l := range reSplitString.FindAllStringSubmatch(mapList, -1) {
		if len(l) > 2 {
			wmap[l[1]] = l[2]
		}
	}

	return wmap
}

// checkPools is used to check the status of all pools from a list
func checkPools(flags flagType, pools []BalancerPool) (string, int) {
	var wAddr, msg string
	var fails, suspect, good, problemBears []string
	var p *BalancerPool
	var workerStat bool

	wmap := newWorkerMapping(flags.WorkerMap)

	globalStatus := OK

	// look at all pools from list:
	for i := range pools {
		// don't work on the copy :-)
		p = &pools[i]

		if flags.Debug || flags.FullStatus {
			fmt.Fprintf(os.Stderr, "Check pool: %s\n", p.Name)
		}
		workerStat = false

		// check every worker from member list
		for j := range p.Workers {
			w := p.Workers[j]
			if flags.Debug || flags.FullStatus {
				fmt.Fprintf(os.Stderr, "  > Worker: %s on %s (type: %s) -> ", w.Address, w.Route, w.Type)
			}

			// Is worker online?
			if !strings.Contains(w.Status, "Ok") {
				// ignore it
				if flags.Debug || flags.FullStatus {
					fmt.Fprintln(os.Stderr, "Fail!")
				}
				w.Status = "ERR"
				problemBears = append(problemBears, w.Route)
				continue
			}

			// strip off port from address, if any
			wAddr = strings.Split(w.Address, ":")[0]

			// Is worker unknown?
			if wmap[wAddr] == "" {
				if flags.Debug || flags.FullStatus {
					fmt.Fprintln(os.Stderr, "Mapping error!")
				}
				w.Status = "ERR"
				continue
			}

			// Pool worker disordered?
			if !strings.HasSuffix(w.Route, wmap[wAddr]) {
				p.StatusOK = false
				w.Status = "ERR"
				problemBears = append(problemBears, w.Route)
				if flags.Debug || flags.FullStatus {
					fmt.Fprintln(os.Stderr, "Fail!")
				}
			} else {
				if flags.Debug || flags.FullStatus {
					fmt.Fprintln(os.Stderr, "Ok.")
				}

				// if we got here, at last one is ok
				workerStat = true
				w.Status = "Ok"
				p.WorkersOK++
			}
		}

		// if at least one worker is online, enable pool for now
		p.StatusOK = workerStat

		// now evaluate the balancer conditions:
		disfunctRatio := int(float64(100) * (float64(1) - float64(p.WorkersOK)/float64(p.WorkersCount)))

		if p.StatusOK == false || disfunctRatio >= flags.Critical {
			// Even one disordered pool or one pool with too few workers is critical
			globalStatus = ErrCrit
			fails = append(fails, p.Name)
			if flags.Debug || flags.FullStatus {
				fmt.Fprintln(os.Stderr, "  > Status: FAIL!")
			}
		} else if disfunctRatio >= flags.Warning {
			globalStatus = ErrWarn
			suspect = append(suspect, p.Name)
			if flags.Debug || flags.FullStatus {
				fmt.Fprintln(os.Stderr, "  > Status: Warn!")
			}
		} else {
			good = append(good, p.Name)
			if flags.Debug || flags.FullStatus {
				fmt.Fprintln(os.Stderr, "  > Status: Ok.")
			}
		}
	}

	// now, set up the overall results.
	switch globalStatus {
	case OK:
		msg = fmt.Sprintf("BALANCER OK - %d pools safe and sound", len(pools))
		if flags.Verbose {
			msg += fmt.Sprintf(" (%s)", strings.Join(good, ", "))
		}
	case ErrWarn:
		msg = fmt.Sprintf("BALANCER WARNING - %d of %d pools with broken workers (%s)", len(suspect), len(pools), strings.Join(suspect, ", "))
		if flags.Verbose {
			msg += fmt.Sprintf(". Bad workers: (%s)", strings.Join(problemBears, ", "))
		}
	case ErrCrit:
		msg = fmt.Sprintf("BALANCER CRITICAL - %d of %d pools broken (%s)", len(fails), len(pools), strings.Join(fails, ", "))
		if flags.Verbose {
			msg += fmt.Sprintf(". Bad workers: (%s)", strings.Join(problemBears, ", "))
		}
	}

	return msg, globalStatus
}

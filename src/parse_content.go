// GetContent: Parse pools from page content

package main

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

func stripTags(s string) string {
	reStripString := "<[^>]*>"

	reStrip := regexp.MustCompile(reStripString)

	return reStrip.ReplaceAllString(s, " ")
}

// parseContent is used to analyse the content of a balancer status page
func parseContent(flags flagType, content string) ([]BalancerPool, error) {
	var err error
	var poolList []BalancerPool

	poolFound := false

	// Remove html Tags
	content = stripTags(content)

	// prepare some RegExps
	reBalancer := regexp.MustCompile(`LoadBalancer Status for (<a.+/a>)? balancer:`)
	reBalancerName := regexp.MustCompile(`//\b(.+)\b`)
	reWorkerData := regexp.MustCompile(`(ajp.?|http.?)://(\S+)\s+(\S+)\s+[\d.]+\s+\d+\s+(\w+\s+\w+)`)

	list := reBalancer.Split(content, -1)

	if len(list) < 2 {
		return nil, errors.New("BALANCER WARNING - Not a balancer manager page")
	}

	// create new reader from buffer
	for _, block := range list[1:] {
		pTmp := reBalancerName.FindStringSubmatch(block)
		if pTmp == nil || len(pTmp) < 2 {
			continue
		}
		// Yes, something found
		pool := BalancerPool{}
		poolFound = true
		pFields := strings.Fields(pTmp[1])
		pool.Name = pFields[0]
		if len(pFields) > 1 {
			pool.Nonce = pFields[1]
		}

		// Search for worker data:
		for _, rawLine := range strings.Split(block, "\n") {
			line := stripTags(rawLine)

			/* 			if flags.Debug {
			   				fmt.Println("*** Deb:" + line)
			   			}
			*/
			if reWorkerData.MatchString(line) {
				wTmp := reWorkerData.FindStringSubmatch(line)
				if wTmp == nil || len(wTmp) < 5 {
					continue
				}

				worker := PoolWorker{}

				worker.Type = wTmp[1]
				worker.Address = wTmp[2]
				worker.Route = wTmp[3]
				worker.Status = wTmp[4]

				pool.Workers = append(pool.Workers, worker)
			}
		}
		pool.WorkersCount = len(pool.Workers)
		poolList = append(poolList, pool)

	}

	if flags.Debug {
		fmt.Printf("*** Deb: Pools found: %+v\n", poolList)
	}

	if !poolFound {
		return nil, errors.New("BALANCER WARNING - No pools found")
	}

	return poolList, err
}

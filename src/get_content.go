// GetContent: Read content of the given web page

package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

// GetContent is used to retrieve the content of a servers balancer status page
func GetContent(flags flagType) (string, error) {
	var err error
	var body []byte
	var host, proto, dsn string

	var useIP bool = false

	if flags.UseSSL {
		proto = "https"
	} else {
		proto = "http"
	}

	if flags.IPAddress != "" {
		host = flags.IPAddress
		useIP = true
	} else {
		host = flags.Hostname
	}

	if flags.Port != "" {
		host = fmt.Sprintf("%s:%s", host, flags.Port)
	}

	if !strings.HasPrefix(flags.URL, "/") {
		flags.URL = "/" + flags.URL
	}

	dsn = fmt.Sprintf("%s://%s%s", proto, host, flags.URL)

	if flags.Debug {
		fmt.Fprintln(os.Stderr, "\nChecking: "+dsn+"\n")
	}

	client := &http.Client{}

	req, err := http.NewRequest("GET", dsn, nil)
	check(err)

	req.Header.Set("User-Agent", flags.Agent)

	// set VHost:
	if useIP {
		req.Host = flags.Hostname
		if flags.Debug {
			fmt.Fprintln(os.Stderr, "\n  VHost: "+flags.Hostname+"\n")
		}
	}

	if flags.User != "" {
		req.SetBasicAuth(flags.User, flags.Password)
	}

	resp, err := client.Do(req)
	check(err)

	if resp.StatusCode != 200 {
		return "", errors.New("Connection refused: " + resp.Status)
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	check(err)

	/* 	if flags.Debug {
	   		fmt.Printf("%+v", string(body))
	   	}
	*/
	return string(body), nil
}

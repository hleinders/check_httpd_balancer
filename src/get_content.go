// GetContent: Read content of the given web page

package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

// GetContent is used to retrieve the content of a servers balancer status page
func GetContent(flags FlagType) (string, error) {
	var err error
	var body []byte
	var proto string

	if flags.UseSSL {
		proto = "https"
	} else {
		proto = "http"
	}

	dsn := fmt.Sprintf("%s://%s/%s", proto, flags.Hostname, flags.URL)

	if flags.Debug {
		fmt.Fprintln(os.Stderr, "\nChecking: "+dsn+"\n")
	}

	client := &http.Client{}

	req, err := http.NewRequest("GET", dsn, nil)
	check(err)

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

	return string(body), nil
}

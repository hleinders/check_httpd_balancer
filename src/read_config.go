package main

import (
	"encoding/json"
	"io/ioutil"
)

// configType is a struct type for the main configuration
type configType struct {
	IPAddress string   `json:"IPAddress"`
	Host      string   `json:"Host"`
	Port      string   `json:"Port"`
	URL       string   `json:"URL"`
	UseSSL    bool     `json:"UseSSL"`
	NoProxy   bool     `json:"NoProxy"`
	WorkerMap []string `json:"WorkerMap"`
}

func readConfig(flags flagType) (configType, error) {
	var config configType
	var err error

	buffer, err := ioutil.ReadFile(flags.ConfigFile)

	if err == nil {
		err = json.Unmarshal(buffer, &config)
	}
	return config, err
}

// prettyPrintJSON (Print JSON Objects with indents)
func prettyPrintJSON(jsonObject interface{}) (string, error) {
	out, err := json.MarshalIndent(jsonObject, "", "   ")

	if err != nil {
		return "", err
	}

	return string(out), nil
}

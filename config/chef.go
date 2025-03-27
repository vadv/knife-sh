package config

import (
	"fmt"
	"github.com/vadv/knife-sh/src/github.com/vadv/chef"
	"io/ioutil"
	"os"
)

// load hosts from chef

func (config *Config) fetchHostsFromChef(q string) error {

	if config.chefKey == `` {
		data, err := ioutil.ReadFile(config.defaultchefKeyPath)
		if err == nil {
			config.chefKey = string(data)
		} else {
			fmt.Fprintf(os.Stderr, "Chef key file access error: %s\n", err.Error())
			os.Exit(1)
		}
	}
	client, err := chef.NewClient(&chef.Config{
		SkipSSL: true,
		Name:    config.chefClient,
		Key:     config.chefKey,
		BaseURL: config.chefUrl,
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Start chef `%s` search query: `%s`\n", config.chefUrl, q)

	part := make(map[string]interface{})
	part["chefAttr"] = []string{config.chefAttr}
	part["fqdn"] = []string{"fqdn"}

	res, err := client.Search.PartialExec("node", q, part)
	if err != nil {
		return err
	}

	for count, row := range res.Rows {
		// row = {"url": "http://chef/node", "data": {"chefAttr": "<response>"}}
		line, ok := row.(map[string]interface{})
		if !ok {
			fmt.Fprintf(os.Stderr, "Bad chef response #1: %#v\n", line)
			os.Exit(1)
		}
		dataRaw, found := line["data"]
		if !found {
			fmt.Fprintf(os.Stderr, "Bad chef response #2: %#v\n", line)
			os.Exit(1)
		}
		data, transform := dataRaw.(map[string]interface{})
		if !transform {
			fmt.Fprintf(os.Stderr, "Bad chef response #3: %#v\n", dataRaw)
			os.Exit(1)
		}
		chefAttr, completed := data["chefAttr"]
		if !completed {
			fmt.Fprintf(os.Stderr, "Empty chefAttr from chef: %#v\n", data)
			os.Exit(1)
		}
		fqdn, completed := data["fqdn"]
		if !completed {
			fmt.Fprintf(os.Stderr, "Empty fqdn from chef: %#v\n", data)
			os.Exit(1)
		}
		config.connectionAddrHosts = append(config.connectionAddrHosts, fmt.Sprintf("%v", chefAttr))
		config.humanReadableHosts = append(config.humanReadableHosts, fmt.Sprintf("%v <%d>", fqdn, count))
	}

	fmt.Fprintf(os.Stderr, "Chef search return %d items\n", len(config.connectionAddrHosts))
	if len(config.connectionAddrHosts) == 0 {
		os.Exit(1)
	}

	return nil
}

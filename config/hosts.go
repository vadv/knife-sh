package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// config.hostSource => config.Hosts

func (config *Config) buildHosts() {

	fromString := func(source, sep string) {
		for _, h := range strings.Split(source, sep) {
			if h == `` {
				continue
			}
			config.hosts = append(config.hosts, strings.Trim(h, ` `))
		}
	}

	fromStdin := func() error {
		if config.hostSource != `-` {
			return fmt.Errorf("")
		}
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			// from pipe
			data, err := ioutil.ReadAll(os.Stdin)
			if err == nil {
				fromString(string(data), "\n")
			} else {
				fmt.Printf("Can't read STDIN from pipe: %v\n", err)
				os.Exit(1)
			}
		} else {
			// from terminal
			fmt.Printf("Send ^D for end input>\n")
			data, err := ioutil.ReadAll(os.Stdin)
			if err == nil {
				fromString(string(data), "\n")
				fmt.Printf("<\n")
			} else {
				fmt.Printf("Can't read STDIN from input: %v\n", err)
				os.Exit(1)
			}
		}
		return nil
	}

	fromFile := func() error {
		data, err := ioutil.ReadFile(config.hostSource)
		if err == nil {
			fmt.Printf("Read file: %v\n", config.hostSource)
			fromString(string(data), "\n")
			return nil
		} else {
			return errors.New("File not found")
		}
	}

	err := fromStdin()
	if err != nil {
		err = fromFile()
	}
	if err != nil {
		if strings.Index(config.hostSource, `:`) != -1 {
			err = config.fetchHostsFromChef(config.hostSource)
		}
	}
	if err != nil {
		fromString(config.hostSource, " ")
	}
}

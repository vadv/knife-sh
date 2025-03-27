package config

import (
	"fmt"
	"io/ioutil"
	"net/netip"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	sshUser, sshKey     string
	command             string
	concurrency         int64
	timeoutSshConnect   int64
	timeoutExec         int64
	stopOnFirstError    bool
	connectionAddrHosts []string
	humanReadableHosts  []string
	// chef settings
	chefClient, chefKey string
	chefUrl, chefAttr   string
	// copy file
	scpSource, scpDest string
	// host input
	hostSource string
	// default variables
	defaultConfigPath  string
	defaultSshKeyPath  string
	defaultchefKeyPath string

	// Proxy ssh Host
	jumpSshHost string
	jumpSshPort uint16
	// Proxy ssh AuthKey
	jumpSshKey string
	// Proxy ssh User
	jumpSshUser string
	// Proxy ssh UserPassword
	jumpSshPass    string
	JumpSshEnabled bool
}

func New() *Config {
	config := getDefaultConfig()
	if err := config.parseFile(config.defaultConfigPath); err != nil {
		fmt.Fprintf(os.Stderr, "Can't parse config `%s`: %s\n", config.defaultConfigPath, err.Error())
		os.Exit(1)
	}
	config.parseCli()
	config.buildHosts()
	return config
}

func getDefaultConfig() *Config {
	home := os.Getenv("HOME")
	user := os.Getenv("USER")
	config := &Config{
		sshUser:             user,
		chefClient:          user,
		chefAttr:            "fqdn",
		chefUrl:             "https://chef.itv.restr.im/organizations/restream/",
		timeoutExec:         0,
		timeoutSshConnect:   10,
		concurrency:         100,
		stopOnFirstError:    false,
		connectionAddrHosts: make([]string, 0),
		humanReadableHosts:  make([]string, 0),
		defaultSshKeyPath:   filepath.Join(home, ".ssh", "id_rsa"),
		defaultchefKeyPath:  filepath.Join(home, ".chef", fmt.Sprintf("%s.pem", user)),
		defaultConfigPath:   filepath.Join(home, ".knife-sh.rc"),
		jumpSshHost:         "",
		jumpSshPort:         22,
		jumpSshKey:          "",
		jumpSshUser:         user,
		jumpSshPass:         "",
	}
	return config
}

func (c *Config) set(key, val string) error {

	switch key {

	case "command":
		c.command = val

	case "timeout-connect", "ssh-timeout", "timeout-ssh":
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		c.timeoutSshConnect = i

	case "timeout-exec", "timeout-execution", "execution-timeout":
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		c.timeoutExec = i

	case "concurrency":
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		c.concurrency = i

	case "ssh-user":
		c.sshUser = val

	case "ssh-key", "identity-file", "identity":
		data, err := ioutil.ReadFile(val)
		if err != nil {
			return err
		}
		c.sshKey = string(data)

	case "chef-client":
		c.chefClient = val

	case "chef-key", "chef-certificate":
		data, err := ioutil.ReadFile(val)
		if err != nil {
			return err
		}
		c.chefKey = string(data)

	case "chef-url":
		c.chefUrl = val

	case "chef-attribute", "chef-attr":
		c.chefAttr = val

	case "stop-on-first-error", "stop-on-error":
		c.stopOnFirstError = (strings.ToLower(val) == "true")

	case "copy-file":
		data := strings.Split(val, ":")
		if len(data) != 2 {
			return fmt.Errorf("bad format for scp. excepted: `source:dest` get: `%s`", val)
		}
		c.scpSource, c.scpDest = data[0], data[1]
		if stat, err := os.Stat(c.scpSource); err != nil {
			return err
		} else {
			fmt.Fprintf(os.Stderr, "File for coping `%s` has size `%d bytes`.\n", c.SCPSource(), stat.Size())
		}

	// enabled use jump host
	case "jump-ssh-enabled":
		c.JumpSshEnabled = val == "true"
	// settings jump connection
	case "jump-ssh-host":
		c.jumpSshHost = val
		if c.jumpSshHost != "" {
			addr, err := netip.ParseAddrPort(val)
			if err != nil {
				return err
			}
			if addr.Port() == 0 {
				c.jumpSshPort = addr.Port()
			}
			c.jumpSshHost = addr.Addr().String()
		}
	case "jump-ssh-user":
		c.jumpSshUser = val
	case "jump-ssh-pass":
		return fmt.Errorf("Don't save jump-ssh-pass in config, it's not secure")
	case "jump-ssh-key", "jump-identity-file", "jump-identity":
		data, err := ioutil.ReadFile(val)
		if err != nil {
			return err
		}
		c.jumpSshKey = string(data)

	default:
		return fmt.Errorf("Unknown key: `%s`", key)
	}

	return nil

}

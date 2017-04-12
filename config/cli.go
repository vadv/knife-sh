package config

import (
	"fmt"
	"os"
	"strings"
)

func printHelpAndExit(err error) {
	config := getDefaultConfig()
	fmt.Println(err)
	fmt.Println("knife-sh HOSTS COMMAND (options)")
	fmt.Printf("\tHOST is 'host1 host2' or /path/to/ip.txt or CHEF:QUERY or - for STDIN\n")
	fmt.Printf("\t-C, --concurrency NUM\tThe number of concurrent connections, default: %d\n", config.concurrency)
	fmt.Printf("\t-x, --ssh-user USERNAME\tThe ssh username, default: %s\n", config.sshUser)
	fmt.Printf("\t-i, --identity-file IDENTITY_FILE,\tdefault: %v\n", config.defaultSshKeyPath)
	fmt.Printf("\t-t, --ssh-timeout SSH TIMEOUT(s)\tThe ssh connection timeout, default: %d\n", config.timeoutSshConnect)
	fmt.Printf("\t-e, --execution-timeout EXECUTION TIMEOUT(s)\tThe command execution timeout, default: %d\n", config.timeoutExec)
	fmt.Printf("\t-c, --copy-file\tCopy file before execution, format: 'local-source:remote-destination'\n")
	fmt.Printf("\t    --chef-client CHEF CLIENT\tChef client name, default: %v\n", config.chefClient)
	fmt.Printf("\t    --chef-certificate CERT FILE\t Path to client certificate, default: %v\n", config.defaultchefKeyPath)
	fmt.Printf("\t-a, --chef-attribute ATTRIBUTE\tChef attribute for connect, default: %v\n", config.chefAttr)
	fmt.Printf("\t-u, --chef-url URL\t\tChef server url, default: %v\n", config.chefUrl)
	fmt.Printf("\nYou can also specify the long-attributes in the config file: ~/.knife-sh.rc in format like ~/.ssh/config ('key = value' or 'key value')\n")

	os.Exit(1)
}

func (config *Config) parseCli() {

	getNextArg := func(sl []string, i int) string {
		if i >= len(sl)-1 {
			printHelpAndExit(fmt.Errorf("Missing args:"))
		}
		return sl[i+1]
	}

	setConfig := func(key, val string) {
		if err := config.set(key, val); err != nil {
			fmt.Fprintf(os.Stderr, "Can't set %s to %s: %s\n", key, val, err.Error())
			os.Exit(1)
		}
	}

	args := os.Args
	args = args[1:len(args)]
	hosts := make([]string, 0)

	skipNextArg := false
	for i, arg := range args {

		if skipNextArg {
			skipNextArg = false
			continue
		}
		skipNextArg = true

		switch arg {

		case "-C", "--concurrency":
			setConfig("concurrency", getNextArg(args, i))

		case "-x", "--ssh-user":
			setConfig("ssh-user", getNextArg(args, i))

		case "-i", "--identity-file":
			setConfig("ssh-key", getNextArg(args, i))

		case "-t", "--ssh-timeout":
			setConfig("ssh-timeout", getNextArg(args, i))

		case "--chef-client":
			setConfig("chef-client", getNextArg(args, i))

		case "--certificate", "--chef-certificate", "--chef-key":
			setConfig("chef-key", getNextArg(args, i))

		case "-u", "--chef-url":
			setConfig("chef-url", getNextArg(args, i))

		case "-a", "--chef-attribute":
			setConfig("chef-attribute", getNextArg(args, i))

		case "-c", "--copy", "--copy-file":
			setConfig("copy-file", getNextArg(args, i))

		case "-e", "--timeout-exec", "--timeout-execution", "--execution-timeout":
			setConfig("timeout-exec", getNextArg(args, i))

		case "-h", "--help":
			printHelpAndExit(fmt.Errorf("Help:"))

		default:
			skipNextArg = false
			hosts = append(hosts, args[i])
		}
	}

	if len(hosts) == 0 {
		printHelpAndExit(fmt.Errorf(""))
	}

	config.command, hosts = hosts[len(hosts)-1], hosts[:len(hosts)-1]
	config.hostSource = strings.Trim(strings.Join(hosts, " "), " ")

	if config.command == "" || config.hostSource == "" {
		printHelpAndExit(fmt.Errorf("Missing commands or hosts"))
	}

}

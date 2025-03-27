package ssh

import (
	"fmt"
	"golang.org/x/term"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh"
)

type hostState struct {
	connectionAddress string
	visibleHostName   string
	command           string
	timeoutConnect    int64
	timeoutExec       int64
	scpData           []byte
	scpDest           string
	startedAt         *time.Time
	connectedAt       *time.Time
	endedAt           *time.Time
	err               error
}

type config interface {
	Hosts() ([]string, []string)
	ConnectTimeout() int64
	ExecTimeout() int64
	SshKeyContent() string
	SshUser() string
	Command() string
	Concurrency() int64
	SCPSource() string
	SCPDest() string
	StopOnFirstError() bool

	JumpEnabled() bool
	JumpSshConnectString() string
	JumpSshUser() string
	JumpSshKeyContent() string
}

func Run(c config) {

	sshConfig := &ssh.ClientConfig{User: c.SshUser(), HostKeyCallback: ssh.InsecureIgnoreHostKey()}
	if c.SshKeyContent() == `` {
		fmt.Fprintf(os.Stderr, "Connect via ssh-agent...\n")
		sshConfig.Auth = []ssh.AuthMethod{ssh.PublicKeys(makeSignersFromAgent()...)}
	} else {
		fmt.Fprintf(os.Stderr, "Connect via ssh-key...\n")
		sshConfig.Auth = []ssh.AuthMethod{ssh.PublicKeys(makeSignersFromKey(c)...)}
	}

	// читаем scp source
	scpData := []byte{}
	if c.SCPSource() != "" {
		if data, err := ioutil.ReadFile(c.SCPSource()); err != nil {
			fmt.Fprintf(os.Stderr, "Read source file error: %s\n", err.Error())
			os.Exit(2)
		} else {
			scpData = data
		}
	}

	hosts := make([]*hostState, 0)
	addrSlice, visibleSlice := c.Hosts()
	for i, addr := range addrSlice {
		connectionAddress, visibleHostName := addr, visibleSlice[i]
		formatedVisibleHostName := connectionAddress
		formatedVisibleHostName = fmt.Sprintf("%s <%s>", visibleHostName, connectionAddress)
		hosts = append(hosts, &hostState{
			scpData:           scpData,
			scpDest:           c.SCPDest(),
			connectionAddress: connectionAddress,
			visibleHostName:   formatedVisibleHostName,
			command:           c.Command(),
			timeoutExec:       c.ExecTimeout(),
			timeoutConnect:    c.ConnectTimeout()})
	}

	// настраиваем вывод
	stdout := make(chan string, 10)
	stderr := make(chan string, 10)
	go func() {
		for {
			select {
			case line := <-stdout:
				fmt.Fprintf(os.Stdout, "%v\n", line)
			case line := <-stderr:
				fmt.Fprintf(os.Stderr, "%v\n", line)
			}
		}
	}()

	// перехватываем Ctr+C
	halt := make(chan os.Signal, 1)
	signal.Notify(halt, os.Interrupt)
	signal.Notify(halt, syscall.SIGTERM)
	go func() {
		<-halt
		report(hosts)
		os.Exit(1)
	}()

	stopOnFirstError := c.StopOnFirstError()
	var wg sync.WaitGroup
	limit := make(chan struct{}, c.Concurrency())
	skipExecNew := false

	bClient, err := BastionConnection(c)
	if err != nil {
		log.Fatal(err)
	}

	for _, h := range hosts {
		if skipExecNew {
			continue
		}
		limit <- struct{}{}
		wg.Add(1)
		go func(h *hostState) {
			h.exec(sshConfig, stdout, stderr, bClient)
			if stopOnFirstError && h.err != nil {
				skipExecNew = true // позволит нам дождаться уже запущенных и пропустить создание новых
			}
			defer wg.Done()
			defer func() { <-limit }()
		}(h)
	}

	wg.Wait()
	report(hosts)

}

func BastionConnection(c config) (*ssh.Client, error) {
	if !c.JumpEnabled() {
		return nil, nil
	}
	sshConfig := &ssh.ClientConfig{
		User:            c.JumpSshUser(),
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	if len(c.JumpSshKeyContent()) > 0 {
		fmt.Printf("Connect to jumpHoset(%s) via ssh-key...\n", c.JumpSshConnectString())
		sshConfig.Auth = []ssh.AuthMethod{ssh.PublicKeys(makeSignersFromKey(c)...)}
	} else {
		fmt.Printf("Connect to jumpHoset(%s) via ssh-user_pass...\n", c.JumpSshConnectString())
		fmt.Print("Please enter your jump password: ")
		password, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return nil, err
		}
		fmt.Printf("\n")
		sshConfig.Auth = []ssh.AuthMethod{ssh.Password(string(password))}
	}
	bClient, err := ssh.Dial("tcp", c.JumpSshConnectString(), sshConfig)
	if err != nil {
		return nil, err
	}
	return bClient, nil
}

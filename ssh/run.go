package ssh

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh"
)

type hostState struct {
	hostname       string
	command        string
	timeoutConnect int64
	timeoutExec    int64
	startedAt      *time.Time
	connectedAt    *time.Time
	endedAt        *time.Time
	err            error
}

type config interface {
	Hosts() []string
	ConnectTimeout() int64
	ExecTimeout() int64
	SshKeyContent() string
	SshUser() string
	Command() string
	Concurrency() int64
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

	hosts := make([]*hostState, 0)
	for _, h := range c.Hosts() {
		hosts = append(hosts, &hostState{
			hostname:       h,
			command:        c.Command(),
			timeoutExec:    c.ExecTimeout(),
			timeoutConnect: c.ConnectTimeout()})
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

	var wg sync.WaitGroup
	limit := make(chan struct{}, c.Concurrency())
	for _, h := range hosts {
		limit <- struct{}{}
		wg.Add(1)
		go func(h *hostState) {
			h.exec(sshConfig, stdout, stderr)
			defer wg.Done()
			defer func() { <-limit }()
		}(h)
	}

	wg.Wait()
	report(hosts)

}

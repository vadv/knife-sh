package ssh

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

func (h *hostState) exec(sshConfig *ssh.ClientConfig, stdout, stderr chan<- string) {

	getSshClient := func(host string) (*ssh.Client, error) {
		address := host
		if strings.Index(host, `:`) == -1 {
			address = fmt.Sprintf("%s:22", host)
		}
		client, err := ssh.Dial("tcp", address, sshConfig)
		if err != nil {
			return nil, err
		}
		return client, nil
	}

	// подготавливаем сессию
	var err error
	var client *ssh.Client
	var session *ssh.Session
	doneSession := make(chan error, 1)
	started := time.Now()
	h.startedAt = &started
	go func() {
		client, err = getSshClient(h.hostname)
		if err != nil {
			doneSession <- err
			return
		}
		session, err = client.NewSession()
		if err != nil {
			doneSession <- err
			return
		}
		doneSession <- nil
	}()

	timeoutConnect := time.Second * time.Duration(h.timeoutConnect)
	if h.timeoutConnect == 0 {
		timeoutConnect = time.Duration(60 * 60 * 24 * time.Second)
	}
	select {
	case h.err = <-doneSession:
		if h.err != nil {
			return
		}
	case <-time.After(timeoutConnect):
		h.err = fmt.Errorf("ssh timeout after: %v", timeoutConnect)
		return
	}
	connectedAt := time.Now()
	h.connectedAt = &connectedAt

	defer client.Close()
	defer session.Close()

	// здесь получили сессию, подготовимся для запуска
	var stdoutPipe, stderrPipe io.Reader
	if h.err = session.RequestPty("xterm", 80, 40, ssh.TerminalModes{ssh.ECHO: 0}); h.err != nil {
		return
	}
	if stdoutPipe, h.err = session.StdoutPipe(); h.err != nil {
		return
	}
	if stderrPipe, err = session.StderrPipe(); err != nil {
		return
	}
	go pipeFeeder(h.hostname+"\t\t", stdoutPipe, stdout)
	go pipeFeeder(h.hostname+"\t\t", stderrPipe, stderr)

	// стартуем комманду
	if h.err = session.Start(h.command); h.err != nil {
		return
	}
	doneExec := make(chan error, 1)
	go func() { doneExec <- session.Wait() }()
	timeoutExecution := time.Second * time.Duration(h.timeoutExec)
	if h.timeoutExec == 0 {
		timeoutExecution = time.Duration(60 * 60 * 24 * time.Second) // мне честно лень
	}
	select {
	case h.err = <-doneExec:
		if h.err != nil {
			return
		}
	case <-time.After(timeoutExecution):
		h.err = fmt.Errorf("exec timeout after: %v", timeoutConnect)
		return
	}
	endedAt := time.Now()
	h.endedAt = &endedAt

	return
}

func pipeFeeder(prefix string, pipe io.Reader, sink chan<- string) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		sink <- prefix + scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		return
	}
}

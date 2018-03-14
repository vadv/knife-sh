package ssh

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
		client, err = getSshClient(h.connectionAddress)
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

	var stdoutPipe, stderrPipe io.Reader
	var stdinPipe io.WriteCloser

	// загружаем файл аля scp
	if h.scpDest != "" {
		session, h.err = client.NewSession()
		if err != nil {
			return
		}
		if stdinPipe, h.err = session.StdinPipe(); err != nil {
			return
		}
		// SendEnv
		if lang := os.Getenv(`LANG`); lang != `` {
			session.Setenv(`LANG`, lang)
		}
		for _, env := range os.Environ() {
			if strings.HasPrefix(env, `LC_`) {
				session.Setenv(env, os.Getenv(env))
			}
		}
		remoteDir, remoteFile := filepath.Dir(h.scpDest), filepath.Base(h.scpDest)
		go func() { session.Run("scp -t " + remoteDir) }()
		time.Sleep(50 * time.Millisecond)
		_, err := fmt.Fprintf(stdinPipe, "C0644 %d %s\n", len(h.scpData), remoteFile)
		if err != nil && err != io.EOF {
			h.err = fmt.Errorf("SCP: create file error: %s", err.Error())
			return
		}
		_, err = fmt.Fprintf(stdinPipe, "%s", h.scpData)
		if err != nil && err != io.EOF {
			h.err = fmt.Errorf("SCP: write error: %s", err.Error())
			return
		}
	}

	// стартуем комманду
	session, h.err = client.NewSession()
	if err != nil {
		return
	}
	defer session.Close()
	if h.err = session.RequestPty("xterm", 80, 40, ssh.TerminalModes{ssh.ECHO: 0}); h.err != nil {
		return
	}
	if stdoutPipe, h.err = session.StdoutPipe(); h.err != nil {
		return
	}
	if stderrPipe, h.err = session.StderrPipe(); err != nil {
		return
	}
	// SendEnv
	if lang := os.Getenv(`LANG`); lang != `` {
		session.Setenv(`LANG`, lang)
	}
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, `LC_`) {
			session.Setenv(env, os.Getenv(env))
		}
	}
	go pipeFeeder(h.visibleHostName+"\t\t", stdoutPipe, stdout)
	go pipeFeeder(h.visibleHostName+"\t\t", stderrPipe, stderr)

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

	time.Sleep(100 * time.Millisecond)
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

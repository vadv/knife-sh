package ssh

import (
	"fmt"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

func makeSignersFromAgent() []ssh.Signer {
	var sshAgent agent.Agent
	signers := []ssh.Signer{}
	agentConn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err == nil {
		sshAgent = agent.NewClient(agentConn)
		agentSigners, err := sshAgent.Signers()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Connect  to ssh agent: %s\n", err.Error())
			os.Exit(1)
		}
		signers = append(signers, agentSigners...)
	}
	return signers
}

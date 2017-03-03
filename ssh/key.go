package ssh

import (
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"
)

func makeSignersFromKey(c config) []ssh.Signer {
	signers := []ssh.Signer{}
	signer, err := ssh.ParsePrivateKey([]byte(c.SshKeyContent()))
	if err == nil {
		signers = append(signers, signer)
	} else {
		fmt.Fprintf(os.Stderr, "Can't make key: %s\n", err.Error())
		os.Exit(1)
	}
	return signers
}

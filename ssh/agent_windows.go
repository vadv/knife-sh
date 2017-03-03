package ssh

import (
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"
)

func makeSignersFromAgent() ssh.Signer {
	fmt.Fprintf(os.Stderr, "Unsupported connect to ssh-agent on this platform")
	os.Exit(1)
}

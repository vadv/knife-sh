package main

import (
	"github.com/vadv/knife-sh/config"
	"github.com/vadv/knife-sh/ssh"
)

func main() {
	c := config.New()
	ssh.Run(c)
}

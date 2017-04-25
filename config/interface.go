package config

func (c *Config) Hosts() map[string]string {
	return c.hosts
}

func (c *Config) ConnectTimeout() int64 {
	return c.timeoutSshConnect
}

func (c *Config) ExecTimeout() int64 {
	return c.timeoutExec
}

func (c *Config) Command() string {
	return c.command
}

func (c *Config) Concurrency() int64 {
	return c.concurrency
}

func (c *Config) SshKeyContent() string {
	return c.sshKey
}

func (c *Config) SshUser() string {
	return c.sshUser
}

func (c *Config) SCPSource() string {
	return c.scpSource
}

func (c *Config) SCPDest() string {
	return c.scpDest
}

func (c *Config) StopOnFirstError() bool {
	return c.stopOnFirstError
}

package scope

import (
	"bytes"
	"fmt"
	"net"

	"golang.org/x/crypto/ssh"
)

type SSHClient struct {
	*ssh.Client
	password string
}

func NewSSHClient(addr, user, password string) (*SSHClient, error) {
	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil }),
	}
	conn, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, err
	}
	return &SSHClient{conn, password}, nil
}

func (conn *SSHClient) RunCommand(cmd string) (string, error) {
	session, err := conn.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	var stdout bytes.Buffer
	session.Stdout = &stdout
	var stderr bytes.Buffer
	session.Stderr = &stderr
	if err := session.Run(cmd); err != nil {
		return stderr.String(), err
	}

	return stdout.String(), nil
}

func (c *SSHClient) RunWithStdin(cmd, stdin string) (string, error) {
	session, err := c.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()
		fmt.Fprint(w, stdin)
	}()

	var stdout bytes.Buffer
	session.Stdout = &stdout
	var stderr bytes.Buffer
	session.Stderr = &stderr
	if err := session.Run(cmd); err != nil {
		return stderr.String(), err
	}

	return stdout.String(), nil
}

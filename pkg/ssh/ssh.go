package ssh

import (
	"context"
	"fmt"
	"io"
	"net"

	"golang.org/x/crypto/ssh"
)

type Client struct {
	addr string

	username string

	command string

	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer

	localPortForwards  []PortForward
	remotePortForwards []PortForward
}

func New(addr string, options ...Option) *Client {
	c := &Client{
		addr: addr,

		username: "root",
	}

	for _, option := range options {
		option(c)
	}

	return c
}

type PortForward struct {
	LocalAddr string
	LocalPort int

	RemoteAddr string
	RemotePort int
}

type Option func(*Client)

func WithStdin(r io.Reader) Option {
	return func(c *Client) {
		c.stdin = r
	}
}

func WithStdout(w io.Writer) Option {
	return func(c *Client) {
		c.stdout = w
	}
}

func WithStderr(w io.Writer) Option {
	return func(c *Client) {
		c.stderr = w
	}
}

func WithCommand(command string) Option {
	return func(c *Client) {
		c.command = command
	}
}

func WithLocalPortForward(p PortForward) Option {
	return func(c *Client) {
		if p.LocalAddr == "" {
			p.LocalAddr = "127.0.0.1"
		}

		if p.RemoteAddr == "" {
			p.RemoteAddr = "127.0.0.1"
		}

		c.localPortForwards = append(c.localPortForwards, p)
	}
}

func WithRemotePortForward(p PortForward) Option {
	return func(c *Client) {
		if p.LocalAddr == "" {
			p.LocalAddr = "127.0.0.1"
		}

		if p.RemoteAddr == "" {
			p.RemoteAddr = "127.0.0.1"
		}

		c.remotePortForwards = append(c.remotePortForwards, p)
	}
}

func (c *Client) Run(ctx context.Context) error {
	config := &ssh.ClientConfig{
		User: c.username,

		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", c.addr, config)

	if err != nil {
		return err
	}

	defer client.Close()

	session, err := client.NewSession()

	if err != nil {
		return err
	}

	session.Stdin = c.stdin
	session.Stdout = c.stdout
	session.Stderr = c.stderr

	defer session.Close()

	for _, p := range c.remotePortForwards {
		listener, err := client.Listen("tcp", fmt.Sprintf("%s:%d", p.RemoteAddr, p.RemotePort))

		if err != nil {
			return err
		}

		defer listener.Close()

		go handleConections(listener, fmt.Sprintf("%s:%d", p.LocalAddr, p.LocalPort))
	}

	if c.command != "" {
		return session.Run(c.command)
	}

	<-ctx.Done()
	return nil
}

func handleConections(l net.Listener, addr string) error {
	for {
		conn, err := l.Accept()

		if err != nil {
			return err
		}

		defer conn.Close()

		go handleConnection(conn, addr)
	}
}

func handleConnection(conn net.Conn, addr string) error {
	l, err := net.Dial("tcp", addr)

	if err != nil {
		return err
	}

	defer l.Close()

	return tunnel(l, conn)
}

func tunnel(local, remote net.Conn) error {
	defer local.Close()
	defer remote.Close()

	result := make(chan error, 2)

	go func() {
		_, err := io.Copy(local, remote)
		result <- err
	}()

	go func() {
		_, err := io.Copy(remote, local)
		result <- err
	}()

	return <-result
}

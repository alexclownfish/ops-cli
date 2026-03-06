package ssh

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"time"
)

type Client struct {
	Host     string
	Port     int
	User     string
	Password string
	client   *ssh.Client
}

func NewClient(host string, port int, user, password string) *Client {
	return &Client{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
	}
}

func (c *Client) Connect() error {
	config := &ssh.ClientConfig{
		User: c.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(c.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}
	
	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return fmt.Errorf("连接失败: %v", err)
	}
	
	c.client = client
	return nil
}

func (c *Client) Execute(cmd string) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("未连接")
	}
	
	session, err := c.client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()
	
	output, err := session.CombinedOutput(cmd)
	return string(output), err
}

func (c *Client) Close() {
	if c.client != nil {
		c.client.Close()
	}
}

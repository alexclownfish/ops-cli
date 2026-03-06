package ssh

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"os"
	"time"
)

type Client struct {
	Host     string
	Port     int
	User     string
	Password string
	KeyPath  string
	KeyPass  string
	client   *ssh.Client
}

func NewClient(host string, port int, user, password, keyPath, keyPass string) *Client {
	return &Client{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		KeyPath:  keyPath,
		KeyPass:  keyPass,
	}
}
func parsePrivateKey(keyPath, passphrase string) (ssh.Signer, error) {
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	if passphrase != "" {
		return ssh.ParsePrivateKeyWithPassphrase(key, []byte(passphrase))
	}
	return ssh.ParsePrivateKey(key)
}

func (c *Client) Connect() error {
	var authMethods []ssh.AuthMethod
	
	if c.KeyPath != "" {
		signer, err := parsePrivateKey(c.KeyPath, c.KeyPass)
		if err != nil {
			return fmt.Errorf("密钥解析失败: %v", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}
	
	if c.Password != "" {
		authMethods = append(authMethods, ssh.Password(c.Password))
	}
	
	config := &ssh.ClientConfig{
		User:            c.User,
		Auth:            authMethods,
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

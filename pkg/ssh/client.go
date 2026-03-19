package ssh

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/ssh"
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

// UploadFile 通过 SCP 上传单个文件到远程服务器
func (c *Client) UploadFile(localPath, remotePath string) error {
	if c.client == nil {
		return fmt.Errorf("未连接")
	}

	// 打开本地文件
	f, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("打开本地文件失败: %v", err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %v", err)
	}

	// 创建 SSH session
	session, err := c.client.NewSession()
	if err != nil {
		return fmt.Errorf("创建session失败: %v", err)
	}
	defer session.Close()

	// 通过 stdin pipe 发送 SCP 协议数据
	w, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("创建stdin pipe失败: %v", err)
	}

	// 启动远端 scp 接收进程
	filename := filepath.Base(localPath)
	if err := session.Start(fmt.Sprintf("scp -t %s", remotePath)); err != nil {
		return fmt.Errorf("启动远端scp失败: %v", err)
	}

	// 发送文件头
	fmt.Fprintf(w, "C%04o %d %s\n", stat.Mode().Perm(), stat.Size(), filename)

	// 发送文件内容
	if _, err := io.Copy(w, f); err != nil {
		return fmt.Errorf("发送文件内容失败: %v", err)
	}

	// 发送结束符
	fmt.Fprint(w, "\x00")
	w.Close()

	return session.Wait()
}

// UploadDir 递归上传目录到远程服务器
func (c *Client) UploadDir(localDir, remoteDir string) error {
	return filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// 计算相对路径
		rel, err := filepath.Rel(localDir, path)
		if err != nil {
			return err
		}

		remotePath := remoteDir + "/" + filepath.ToSlash(rel)

		// 确保远端目录存在
		remoteFileDir := filepath.Dir(remotePath)
		if _, err := c.Execute(fmt.Sprintf("mkdir -p %s", remoteFileDir)); err != nil {
			return fmt.Errorf("创建远端目录失败: %v", err)
		}

		return c.UploadFile(path, remotePath)
	})
}

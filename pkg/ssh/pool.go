package ssh

import (
	"fmt"
	"sync"
	"time"
)

// Pool SSH连接池
type Pool struct {
	mu       sync.RWMutex
	clients  map[string]*Client // key: "host:port"
	maxSize  int
	timeout  time.Duration
}

// PoolClient 带有过期时间的客户端
type PoolClient struct {
	client   *Client
	lastUsed time.Time
}

// NewPool 创建SSH连接池
func NewPool(maxSize int) *Pool {
	return &Pool{
		clients: make(map[string]*Client),
		maxSize: maxSize,
		timeout: 30 * time.Minute, // 默认30分钟过期
	}
}

// Get 从连接池获取客户端，如果不存在则创建
func (p *Pool) Get(host string, port int, user, password, keyPath, keyPass string) (*Client, error) {
	key := fmt.Sprintf("%s:%d", host, port)

	p.mu.RLock()
	if client, ok := p.clients[key]; ok {
		p.mu.RUnlock()
		// 检查连接是否有效
		if p.isAlive(client) {
			return client, nil
		}
		// 连接无效，删除并重建
		p.mu.Lock()
		delete(p.clients, key)
		p.mu.Unlock()
	} else {
		p.mu.RUnlock()
	}

	// 创建新连接
	client := NewClient(host, port, user, password, keyPath, keyPass)
	if err := client.Connect(); err != nil {
		return nil, err
	}

	p.mu.Lock()
	// 如果池满了，清理过期连接
	if len(p.clients) >= p.maxSize {
		p.cleanup()
	}
	p.clients[key] = client
	p.mu.Unlock()

	return client, nil
}

// isAlive 检查连接是否存活
func (p *Pool) isAlive(client *Client) bool {
	if client == nil || client.client == nil {
		return false
	}

	// 尝试创建session来验证连接
	session, err := client.client.NewSession()
	if err != nil {
		return false
	}
	session.Close()
	return true
}

// cleanup 清理无效连接
func (p *Pool) cleanup() {
	for key, client := range p.clients {
		if !p.isAlive(client) {
			if client != nil {
				client.Close()
			}
			delete(p.clients, key)
		}
	}
}

// CloseAll 关闭所有连接
func (p *Pool) CloseAll() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for key, client := range p.clients {
		if client != nil {
			client.Close()
		}
		delete(p.clients, key)
	}
}

// Size 返回连接池大小
func (p *Pool) Size() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.clients)
}

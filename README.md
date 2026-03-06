# ops-cli

> 这是一个AI写的，也不知道能写成啥样 🤖

一个集成了常用运维功能的Go语言工具集

## ✨ 功能特性

### ✅ 已实现
- **SSH单机执行** - 在单台服务器上执行命令
- **SSH批量执行** - 并发在多台服务器上执行命令
- **密钥认证** - 支持SSH密钥登录（含加密密钥）
- **配置文件** - 从YAML文件读取服务器列表
- **并发控制** - 可配置并发数

### 🚧 规划中
- 服务器监控告警
- 日志分析
- 自动化部署

## 📦 安装

### 下载二进制文件```bash
# Linux/macOS
wget https://github.com/alexclownfish/ops-cli/raw/master/ops
chmod +x ops
sudo mv ops /usr/local/bin/

# Windows
# 下载 ops.exe 并添加到PATH
```

### 从源码编译
```bash
git clone https://github.com/alexclownfish/ops-cli.git
cd ops-cli
go build -o ops
```

## 🚀 快速开始

### 单机执行命令

**密码认证：**
```bash
ops exec "uptime" -H 192.168.1.100 -u root -p password
```

**密钥认证：**
```bash
ops exec "df -h" -H 192.168.1.100 -u root -i ~/.ssh/id_rsa
```

**带密码的密钥：**
```bash
ops exec "free -h" -H 192.168.1.100 -u root -i ~/.ssh/id_rsa --key-pass mypass
```

### 批量执行命令

**手动指定服务器列表：**
```bash
ops batch "uptime" -L 192.168.1.100,192.168.1.101,192.168.1.102 -U root -P password
```

**使用配置文件：**
```bash
ops batch "df -h" -c servers.yaml
```

**使用密钥批量执行：**
```bash
ops batch "free -m" -L server1,server2,server3 -U root -K ~/.ssh/id_rsa --batch-key-pass mypass
```

**控制并发数：**
```bash
ops batch "uptime" -c servers.yaml --parallel 5
```

## 📝 配置文件

创建 `servers.yaml`：
```yaml
servers:
  - name: web-server-1
    host: 192.168.1.100
    port: 22
    user: root
    password: password123
  - name: web-server-2
    host: 192.168.1.101
    port: 22
    user: root
    password: password123
  - name: db-server
    host: 192.168.1.102
    port: 22
    user: admin
    password: admin123
```

## 📖 命令参数

### exec 命令
```
-H, --host string       服务器地址 (必需)
-P, --port int          SSH端口 (默认: 22)
-u, --user string       用户名 (默认: root)
-p, --password string   密码
-i, --key string        私钥路径
    --key-pass string   私钥密码
```

### batch 命令
```
-c, --config string           配置文件路径
-L, --hosts strings           服务器列表 (逗号分隔)
-T, --batch-port int          SSH端口 (默认: 22)
-U, --batch-user string       用户名 (默认: root)
-P, --batch-password string   密码
-K, --batch-key string        私钥路径
    --batch-key-pass string   私钥密码
    --parallel int            并发数 (默认: 10)
```

## 🔐 认证方式

**优先级：** 密钥认证 > 密码认证

- 如果同时提供密钥和密码，优先使用密钥
- 密钥认证失败时，自动尝试密码认证
- 支持加密的私钥文件（需提供 `--key-pass`）

## 💡 使用技巧

1. **配置文件 vs 命令行参数**
   - 服务器数量少：直接用 `-L` 参数
   - 服务器数量多：使用配置文件 `-c`

2. **并发控制**
   - 默认并发10个，可根据网络情况调整
   - 服务器多时建议降低并发数避免网络拥堵

3. **安全建议**
   - 生产环境建议使用密钥认证
   - 配置文件中的密码建议加密存储
   - 使用 `chmod 600` 保护配置文件

## 🛠️ 开发

```bash
# 克隆仓库
git clone https://github.com/alexclownfish/ops-cli.git
cd ops-cli

# 安装依赖
go mod tidy

# 编译
go build -o ops

# 运行测试
go test ./...
```

## 📄 License

MIT

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

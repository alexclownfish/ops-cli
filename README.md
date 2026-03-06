# ops-cli

> 这是一个AI写的，也不知道能写成啥样 🤖

一个集成了常用运维功能的Go语言工具集

## ✨ 功能特性

### ✅ SSH管理
- **SSH单机执行** - 在单台服务器上执行命令
- **SSH批量执行** - 并发在多台服务器上执行命令
- **密钥认证** - 支持SSH密钥登录（含加密密钥）
- **配置文件** - 从YAML文件读取服务器列表
- **并发控制** - 可配置并发数

### ✅ 密码管理（新功能）
- **密码生成** - 生成24位强密码
- **加密存储** - AES-256-GCM加密
- **密码轮换** - 自动密码生命周期管理
- **批量改密** - 支持批量修改服务器密码
- **KeePassXC导出** - 导出为KDBX格式

### 🚧 规划中
- 服务器监控告警
- 日志分析
- 自动化部署

## 📦 安装
### 下载二进制文件
```bash
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

### SSH命令执行

**单机执行（密码认证）：**
```bash
ops exec "uptime" -H 192.168.1.100 -u root -p password
```

**单机执行（密钥认证）：**
```bash
ops exec "df -h" -H 192.168.1.100 -u root -i ~/.ssh/id_rsa
```

**批量执行：**
```bash
ops batch "uptime" -L 192.168.1.100,192.168.1.101 -U root -P password
```

**使用配置文件：**
```bash
ops batch "df -h" -c servers.yaml
```

### 密码管理

**生成强密码：**
```bash
ops passwd generate
# 输出: ✅ 生成密码: O#y-XQyEXl9I6@hQKIcMt9Zk
```

**保存密码（即将实现）：**
```bash
ops passwd save vm-001 --host 192.168.1.100 --user root
# 自动生成密码并加密存储
```

**查看密码（即将实现）：**
```bash
ops passwd show vm-001
# 输出解密后的密码
```

**批量改密（即将实现）：**
```bash
ops passwd reset-batch --config servers.yaml
# 批量修改所有服务器密码
```

**导出到KeePassXC（即将实现）：**
```bash
ops passwd export --format kdbx --output passwords.kdbx
# 导出为KeePass数据库格式
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

### passwd 命令
```
generate                生成强密码
save                    保存密码（即将实现）
show                    查看密码（即将实现）
reset                   修改密码（即将实现）
export                  导出密码（即将实现）
```

## 🔐 密码管理特性

### 密码生成规则
- 长度：24位
- 字符集：A-Z、a-z、0-9、#_-@~
- 必须包含所有类型字符

### 加密存储
- 算法：AES-256-GCM
- 存储：BoltDB加密数据库
- 主密钥：32字节随机密钥

### 密码生命周期
- 有效期：90天（可配置）
- 自动轮换：提前5天自动改密
- 审计日志：记录所有操作

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
   - 定期轮换密码

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

## 📚 文档

详细设计文档：[docs/password-manager-design.md](docs/password-manager-design.md)

## 📄 License

MIT

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

---

**更新日志：**
- 2026-03-06: 添加密码管理功能（生成、加密存储）
- 2026-03-05: 添加SSH批量执行和密钥认证
- 2026-03-04: 初始版本

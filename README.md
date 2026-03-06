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

### ✅ 密码管理
- **密码生成** - 生成24位强密码（A-Z、a-z、0-9、#_-@~）
- **加密存储** - AES-256-GCM加密，BoltDB数据库
- **SSH改密** - 使用chpasswd命令改密
- **virsh改密** - 支持OpenStack虚拟机改密（无需原密码）
- **批量改密** - 支持批量修改服务器密码
- **密码查看** - 查看已保存的密码
- **KeePassXC导出** - 导出为KDBX格式，可直接导入KeePassXC

### 🚧 规划中
- 密码生命周期管理（90天自动轮换）
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

**1. 生成强密码：**
```bash
ops passwd generate
# 输出: ✅ 生成密码: O#y-XQyEXl9I6@hQKIcMt9Zk
```

**2. 保存密码到数据库：**
```bash
# 生成主密钥（首次使用）
MASTER_KEY="your-32-byte-master-key-here"

# 保存密码（自动生成新密码）
ops passwd save vm-001 -H 192.168.1.100 -u root --key $MASTER_KEY
```

**3. 查看已保存的密码：**
```bash
ops passwd show vm-001 --key $MASTER_KEY
```

**4. 列出所有服务器：**
```bash
ops passwd list --key $MASTER_KEY
```

**5. 修改密码（单机）：**
```bash
# SSH方式改密
ops passwd reset vm-001 --key $MASTER_KEY

# virsh方式改密（OpenStack）
# 需要在保存时指定reset_method=virsh和物理机信息
```

**6. 批量改密：**
```bash
ops passwd reset-batch --key $MASTER_KEY
# 自动遍历所有服务器并改密
```

**7. 导出到KeePassXC：**
```bash
ops passwd export --key $MASTER_KEY --kdbx-password "keepass-password" --output passwords.kdbx
# 然后在KeePassXC中打开passwords.kdbx
```

## 📝 配置文件

### SSH批量执行配置
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
save                    保存密码
show                    查看密码
list                    列出所有服务器
reset                   修改密码（单机）
reset-batch             批量改密
export                  导出到KeePassXC

通用参数：
--db string             数据库路径 (默认: passwords.db)
--key string            主密钥 (必需)
```

## 🔐 密码管理特性

### 密码生成规则
- **长度**：24位
- **字符集**：A-Z、a-z、0-9、#_-@~
- **验证**：必须包含所有类型字符
- **安全性**：使用crypto/rand生成

### 加密存储
- **算法**：AES-256-GCM（军事级加密）
- **存储**：BoltDB加密数据库
- **主密钥**：32字节随机密钥
- **权限**：数据库文件权限600

### 改密方式

**1. SSH改密（通用）**
- 使用chpasswd命令
- 需要知道当前密码
- 适用于所有Linux服务器

**2. virsh改密（OpenStack）**
- 使用virsh set-user-password
- 无需知道当前密码
- 需要物理机访问权限
- 适用于OpenStack虚拟机

### KeePassXC导出
- **格式**：KDBX（KeePass 2.x）
- **加密**：密码保护
- **导入**：可直接在KeePassXC中打开
- **字段**：标题、用户名、密码、URL、备注

## 💡 使用技巧

### 1. 主密钥管理
```bash
# 生成主密钥（32字节）
openssl rand -base64 32

# 保存到环境变量
export OPS_MASTER_KEY="your-master-key"

# 使用环境变量
ops passwd list --key $OPS_MASTER_KEY
```

### 2. 配置文件 vs 命令行
- **服务器数量少**：直接用 `-L` 参数
- **服务器数量多**：使用配置文件 `-c`
- **密码管理**：使用数据库存储

### 3. 并发控制
- 默认并发10个，可根据网络情况调整
- 服务器多时建议降低并发数避免网络拥堵

### 4. 安全建议
- 生产环境建议使用密钥认证
- 主密钥不要提交到Git
- 数据库文件使用 `chmod 600` 保护
- 定期备份密码数据库
- 导出的KDBX文件及时删除

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

- **设计文档**：[docs/password-manager-design.md](docs/password-manager-design.md)
- **GitHub仓库**：https://github.com/alexclownfish/ops-cli

## 🎯 使用场景

### 场景1：日常运维
```bash
# 批量执行命令
ops batch "df -h" -c servers.yaml
```

### 场景2：密码管理
```bash
# 保存密码
ops passwd save vm-001 -H 192.168.1.100 --key $KEY

# 定期改密
ops passwd reset-batch --key $KEY
```

### 场景3：密码导出
```bash
# 导出到KeePassXC备份
ops passwd export --key $KEY --kdbx-password backup123
```

## 📄 License

MIT

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

---

**更新日志：**
- 2026-03-06: 添加密码管理完整功能（生成、加密存储、SSH改密、virsh改密、批量改密、KeePassXC导出）
- 2026-03-05: 添加SSH批量执行和密钥认证
- 2026-03-04: 初始版本

**Star ⭐ 如果这个项目对你有帮助！**

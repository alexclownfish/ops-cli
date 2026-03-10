# ops-cli

一个集成了SSH管理、密码管理、OpenStack集成的运维工具集

## ✨ 功能特性

### SSH管理
- SSH单机/批量命令执行
- 支持密码和密钥认证
- 支持配置文件批量管理
- 可配置并发数

### 密码管理
- 24位强密码生成（A-Z、a-z、0-9、#_-@~）
- AES-256-GCM加密存储
- 主密钥验证（防止未授权访问）
- 智能改密（SSH优先，virsh回退）
- 批量改密
- KeePassXC导出（KDBX格式）

### 密码生命周期管理
- 密码年龄检查
- 自动改密（密码到期）
- 定时任务支持

### OpenStack集成
- 自动导入虚拟机
- 自动获取物理机IP
- virsh改密支持

---

## 📦 安装

### 下载二进制文件
```bash
# Linux
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

---

## 🚀 快速开始

### 1. SSH命令执行

**单机执行：**
```bash
# 密码认证
ops exec "uptime" -H 192.168.1.100 -u root -p password

# 密钥认证
ops exec "df -h" -H 192.168.1.100 -u root -i ~/.ssh/id_rsa
```

**批量执行：**
```bash
# 命令行指定
ops batch "uptime" -L 192.168.1.100,192.168.1.101 -U root -P password

# 配置文件
ops batch "df -h" -c servers.yaml
```

---

### 2. 密码管理基础

**生成主密钥（首次使用）：**
```bash
# 生成32字节随机密钥
openssl rand -base64 32

# 保存到环境变量
export OPS_MASTER_KEY="eQ9P1niDITFavopNKbgmxxmSc5jIEg5zKUvNnZTUpEs="
```

**基本操作：**
```bash
# 生成强密码
ops passwd generate

# 保存密码（自动生成新密码）
ops passwd save vm-001 -H 192.168.1.100 -u root --key $OPS_MASTER_KEY

# 查看密码
ops passwd show vm-001 --key $OPS_MASTER_KEY

# 列出所有服务器
ops passwd list --key $OPS_MASTER_KEY

# 删除服务器
ops passwd delete vm-001 --key $OPS_MASTER_KEY
```

---

### 3. 密码修改

**单机改密（智能选择方式）：**
```bash
ops passwd reset vm-001 --key $OPS_MASTER_KEY
```

**批量改密：**
```bash
ops passwd reset-batch --key $OPS_MASTER_KEY
```

**智能改密流程：**
1. 优先尝试SSH改密（快速、不依赖物理机）
2. SSH失败自动回退virsh（适合应急场景）
3. 自动记录成功的改密方式

---

### 4. OpenStack集成

**从OpenStack自动导入虚拟机：**
```bash
# 设置环境变量
export OS_AUTH_URL=http://keystone-admin.cty.os:10006/v3
export OS_USERNAME=admin
export OS_PASSWORD=your-password
export OS_PROJECT_NAME=admin
export OPS_MASTER_KEY="your-master-key"

# 一键导入
ops passwd import \
  --hypervisor-port 10000 \
  --hypervisor-user secure \
  --hypervisor-key /root/.ssh/id_rsa \
  --vm-user secure \
  --key $OPS_MASTER_KEY
```

**自动完成：**
- ✅ 获取所有虚拟机信息
- ✅ 自动获取虚拟机IP地址
- ✅ 自动获取OpenStack实例ID
- ✅ 自动获取物理机IP
- ✅ 自动生成密码并保存

---

### 5. 密码生命周期管理

**检查密码年龄：**
```bash
ops passwd check-age --key $OPS_MASTER_KEY --days 85
```

**手动执行自动改密：**
```bash
ops passwd auto-rotate --key $OPS_MASTER_KEY --days 85
```

**配置定时任务（完全自动化）：**
```bash
# 1. 复制脚本（项目根目录下）
cp password-rotate.sh /usr/local/bin/
chmod +x /usr/local/bin/password-rotate.sh

# 2. 设置环境变量
export OPS_MASTER_KEY="your-master-key"

# 3. 配置cron（每天凌晨2点）
echo "0 2 * * * root /usr/local/bin/password-rotate.sh >> /var/log/ops-rotate.log 2>&1" > /etc/cron.d/ops-password-rotate
```

---

### 6. KeePassXC导出

```bash
ops passwd export-keepass \
  --key $OPS_MASTER_KEY \
  --output passwords.kdbx
```

---

### 7. 主密钥管理

**更换主密钥（安全）：**
```bash
ops passwd reset-key \
  --old-key $OLD_MASTER_KEY \
  --new-key $NEW_MASTER_KEY \
  --db passwords.db
```

---

## 📋 命令参考

### 密码管理命令

```bash
ops passwd generate                    # 生成强密码
ops passwd save <id> -H <host> -u <user> --key <key>  # 保存密码
ops passwd show <id> --key <key>       # 查看密码
ops passwd list --key <key>            # 列出所有服务器
ops passwd delete <id> --key <key>     # 删除服务器
ops passwd reset <id> --key <key>      # 修改密码（单机）
ops passwd reset-batch --key <key>     # 批量改密
ops passwd import --key <key>          # 从OpenStack导入
ops passwd export-keepass --key <key> --output <file>  # 导出到KeePassXC
ops passwd check-age --key <key> --days <days>     # 检查密码年龄
ops passwd auto-rotate --key <key> --days <days>   # 自动改密
ops passwd reset-key --old-key <old> --new-key <new>  # 更换主密钥
```

---

## 🔐 安全特性

### 密码生成规则
- **长度**：24位
- **字符集**：A-Z、a-z、0-9、#_-@~
- **验证**：必须包含所有类型字符
- **安全性**：使用crypto/rand生成

### 加密存储
- **算法**：AES-256-GCM（军事级加密）
- **存储**：BoltDB加密数据库
- **主密钥**：32字节随机密钥
- **验证**：主密钥哈希验证（防止未授权访问）

### 改密方式

**SSH改密：**
- 使用 `/usr/sbin/chpasswd` 命令
- 需要sudo权限（免密sudo）
- 适用于所有Linux服务器

**virsh改密：**
- 使用 `virsh set-user-password` 命令
- 无需知道当前密码
- 需要物理机访问权限
- 适用于OpenStack虚拟机

---

## 💡 使用技巧

### 主密钥管理最佳实践

```bash
# 生成主密钥
openssl rand -base64 32 > master.key

# 保存到环境变量
export OPS_MASTER_KEY=$(cat master.key)

# 安全存储
chmod 600 master.key
```

### OpenStack环境配置

```bash
# 创建环境变量文件
cat > ~/.openstack_env << EOF
export OS_AUTH_URL=http://keystone-admin.cty.os:10006/v3
export OS_USERNAME=admin
export OS_PASSWORD=your-password
export OS_PROJECT_NAME=admin
export OPS_MASTER_KEY="your-master-key"
EOF

# 使用时加载
source ~/.openstack_env
ops passwd import ...
```

---

## 🐛 故障排查

### SSH改密失败

**问题：** `Permission denied`

**解决：** 配置免密sudo
```bash
echo "secure ALL=(ALL) NOPASSWD: ALL" | sudo tee /etc/sudoers.d/secure
```

### virsh改密失败

**问题：** `error: Guest agent is not responding`

**解决：** 安装并启动qemu-guest-agent
```bash
yum install -y qemu-guest-agent
systemctl start qemu-guest-agent
systemctl enable qemu-guest-agent
```

---

## 📚 更多信息

- **GitHub**: https://github.com/alexclownfish/ops-cli
- **Issues**: https://github.com/alexclownfish/ops-cli/issues

---

## 📄 许可证

MIT License
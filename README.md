# ops-cli

一个集成了SSH管理、密码管理、OpenStack集成的运维工具集

## 功能特性

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
- 自动备份导出

### OpenStack集成
- 自动导入虚拟机
- 自动获取虚拟机IP、实例ID
- 自动获取物理机IP
- virsh改密支持

---

## 📦 安装

### 方式1：下载二进制文件（推荐）

```bash
# Linux
wget https://github.com/alexclownfish/ops-cli/raw/master/ops
chmod +x ops
sudo mv ops /usr/local/bin/

# 验证安装
ops --version
```

### 方式2：从源码编译

```bash
git clone https://github.com/alexclownfish/ops-cli.git
cd ops-cli
go build -o ops
sudo mv ops /usr/local/bin/
```

---

## 🚀 快速开始

### 第一步：生成主密钥

> ⚠️ 主密钥用于加密所有密码，请妥善保管！

```bash
# 生成32字节随机密钥
openssl rand -base64 32
# 输出示例：eQ9P1niDITFavopNKbgmxxmSc5jIEg5zKUvNnZTUpEs=

# 保存到环境变量（推荐）
export OPS_MASTER_KEY="eQ9P1niDITFavopNKbgmxxmSc5jIEg5zKUvNnZTUpEs="

# 或写入 bashrc 永久生效
echo 'export OPS_MASTER_KEY="eQ9P1niDITFavopNKbgmxxmSc5jIEg5zKUvNnZTUpEs="' >> ~/.bashrc
source ~/.bashrc
```

**安全建议：**
- 将主密钥保存到安全的地方（如KeePassXC）
- 不要将主密钥提交到Git仓库
- 定期更换主密钥（使用 `reset-key` 命令）

---

## 📖 完整使用指南

详细的使用指南、扩展案例和故障排查，请查看：
- [完整文档](https://github.com/alexclownfish/ops-cli/wiki)
- [快速开始](https://github.com/alexclownfish/ops-cli/wiki/Quick-Start)
- [扩展案例](https://github.com/alexclownfish/ops-cli/wiki/Examples)

### 常用命令速查

```bash
# 密码管理
ops passwd generate                                    # 生成强密码
ops passwd save <id> -H <host> -u <user> --key <key>  # 保存密码
ops passwd show <id> --key <key>                       # 查看密码
ops passwd list --key <key>                            # 列出所有服务器
ops passwd reset <id> --key <key>                      # 修改密码（单机）
ops passwd reset-batch --key <key>                     # 批量改密
ops passwd export --key <key> --output <file>          # 导出到KeePassXC

# 密码生命周期
ops passwd check-age --key <key> --days <days>         # 检查密码年龄
ops passwd auto-rotate --key <key> --days <days>       # 自动改密

# OpenStack集成
ops passwd import --key <key> [options]                # 从OpenStack导入

# SSH命令
ops exec <command> -H <host> -u <user> -p <password>  # 单机执行
ops batch <command> -c servers.yaml                    # 批量执行
```

---

## 🔧 扩展功能案例

### 案例1：自动密码轮换（定时任务）

```bash
# 部署脚本
cp password-rotate.sh /usr/local/bin/
chmod +x /usr/local/bin/password-rotate.sh

# 配置cron（每天凌晨2点执行）
cat > /etc/cron.d/ops-password-rotate << 'EOF'
0 2 * * * root OPS_MASTER_KEY="your-key" /usr/local/bin/password-rotate.sh
EOF
```

**自动化流程：**
1. ✅ 检查密码年龄
2. ✅ 自动改密到期服务器
3. ✅ 导出KDBX备份
4. ✅ 清理30天前的备份
5. ✅ 详细日志记录

**日志示例：**
```
[2026-03-11 16:09:34] 开始密码轮换任务
[2026-03-11 16:09:34] 需要改密的服务器（5台）:
[2026-03-11 16:09:35] ✅ 改密完成: 成功 5/5 台
[2026-03-11 16:09:35] 备份导出成功: /opt/oppos/backup/passwords/passwords-20260311.kdbx
[2026-03-11 16:09:35] 密码轮换任务完成
```

---

### 案例2：多项目OpenStack批量管理

```bash
#!/bin/bash
# 从多个OpenStack项目批量导入虚拟机

PROJECTS=("project-a" "project-b" "project-c")
KEY="$OPS_MASTER_KEY"

for PROJECT in "${PROJECTS[@]}"; do
    echo "正在导入项目: $PROJECT"
    export OS_PROJECT_NAME="$PROJECT"
    
    ops passwd import         --hypervisor-port 10000         --hypervisor-user secure         --hypervisor-key /root/.ssh/id_rsa         --vm-user secure         --key "$KEY"
done
```

---

### 案例3：改密后自动验证

```bash
#!/bin/bash
# 改密并验证新密码是否有效

KEY="$OPS_MASTER_KEY"
SERVERS=$(ops passwd list --key "$KEY" | grep "^  名称:" | awk '{print $2}')

for SERVER in $SERVERS; do
    INFO=$(ops passwd show "$SERVER" --key "$KEY")
    IP=$(echo "$INFO" | grep "地址:" | awk '{print $2}')
    USER=$(echo "$INFO" | grep "用户:" | awk '{print $2}')
    PASS=$(echo "$INFO" | grep "密码:" | awk '{print $2}')
    
    # 测试登录
    if sshpass -p "$PASS" ssh -o ConnectTimeout=5         "${USER}@${IP}" "echo ok" &>/dev/null; then
        echo "✅ $SERVER 验证成功"
    else
        echo "❌ $SERVER 验证失败"
    fi
done
```

---

### 案例4：与Ansible集成

```bash
#!/bin/bash
# 使用ops-cli获取密码，然后运行Ansible

KEY="$OPS_MASTER_KEY"
TARGET="vm-web-001"

INFO=$(ops passwd show "$TARGET" --key "$KEY")
IP=$(echo "$INFO" | grep "地址:" | awk '{print $2}')
USER=$(echo "$INFO" | grep "用户:" | awk '{print $2}')
PASS=$(echo "$INFO" | grep "密码:" | awk '{print $2}')

# 生成临时inventory
cat > /tmp/ansible_hosts << EOF
[targets]
$IP ansible_user=$USER ansible_password=$PASS
EOF

# 运行Ansible
ansible-playbook -i /tmp/ansible_hosts your-playbook.yml
rm -f /tmp/ansible_hosts
```

---

### 案例5：密码轮换通知（企业微信/钉钉）

```bash
#!/bin/bash
# 改密完成后发送通知

KEY="$OPS_MASTER_KEY"
WEBHOOK_URL="https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=your-key"

OUTPUT=$(ops passwd auto-rotate --key "$KEY" --days 85 2>&1)
SUCCESS=$(echo "$OUTPUT" | grep -oP '成功 \K\d+')
TOTAL=$(echo "$OUTPUT" | grep -oP '/\K\d+ 台' | tr -d ' 台')

MSG="✅ 密码轮换完成
时间: $(date '+%Y-%m-%d %H:%M:%S')
成功: ${SUCCESS}/${TOTAL} 台"

curl -s -X POST "$WEBHOOK_URL"     -H "Content-Type: application/json"     -d "{"msgtype":"text","text":{"content":"${MSG}"}}"
```

---

### 案例6：分级密码轮换策略

```bash
#!/bin/bash
# 核心服务器30天，普通服务器90天

KEY="$OPS_MASTER_KEY"

# 核心服务器（30天轮换）
ops passwd auto-rotate --key "$KEY" --days 30

# 普通服务器（90天轮换）
ops passwd auto-rotate --key "$KEY" --days 90

# 导出统一备份
ops passwd export     --key "$KEY"     --output "/backup/passwords/all-$(date +%Y%m%d).kdbx"
```

---

## 🔐 安全特性

### 密码生成
- **长度**：24位
- **字符集**：A-Z、a-z、0-9、#_-@~
- **安全性**：crypto/rand（密码学安全随机数）

### 加密存储
- **算法**：AES-256-GCM（军事级加密）
- **数据库**：BoltDB
- **主密钥**：32字节Base64编码
- **权限**：数据库文件600（仅所有者可读写）

### 改密方式

#### SSH改密
- **命令**：`/usr/sbin/chpasswd`（完整路径）
- **权限**：需要sudo（建议配置免密sudo）
- **适用**：所有Linux服务器
- **优势**：速度快，不依赖物理机

#### virsh改密
- **命令**：`virsh set-user-password`
- **权限**：需要物理机访问
- **适用**：OpenStack虚拟机
- **优势**：无需知道当前密码
- **要求**：需安装qemu-guest-agent

---

## 🐛 故障排查

### SSH改密失败

**问题：Permission denied**

```bash
# 配置免密sudo
echo "secure ALL=(ALL) NOPASSWD: ALL" | sudo tee /etc/sudoers.d/secure
chmod 440 /etc/sudoers.d/secure
```

---

### virsh改密失败

**问题：Guest agent is not responding**

```bash
# 安装qemu-guest-agent
apt-get install -y qemu-guest-agent  # Ubuntu/Debian
yum install -y qemu-guest-agent      # CentOS/RHEL

# 启动服务
systemctl enable --now qemu-guest-agent
```

---

### 主密钥错误

```bash
# 检查环境变量
echo $OPS_MASTER_KEY

# 确认密钥文件
cat ~/.ops_master_key
```

---

### OpenStack导入失败

```bash
# 检查环境变量
env | grep OS_

# 测试连接
openstack server list

# 切换项目
export OS_PROJECT_NAME=your-project
```

---

## 💡 最佳实践

### 1. 主密钥管理

```bash
# 生成并安全保存
openssl rand -base64 32 > ~/.ops_master_key
chmod 600 ~/.ops_master_key

# 加载到环境变量
export OPS_MASTER_KEY=$(cat ~/.ops_master_key)
echo 'export OPS_MASTER_KEY=$(cat ~/.ops_master_key)' >> ~/.bashrc
```

### 2. 推荐配置
- **有效期**：85天（留15天缓冲，实际90天轮换）
- **检查频率**：每天凌晨2点
- **备份保留**：30天
- **日志保留**：90天

### 3. 日志管理

```bash
# 配置logrotate
cat > /etc/logrotate.d/ops-rotate << 'EOF'
/var/log/ops-rotate.log {
    daily
    rotate 90
    compress
    missingok
    notifempty
}
EOF
```

---

## 📚 更多信息

- **GitHub**: https://github.com/alexclownfish/ops-cli
- **问题反馈**: https://github.com/alexclownfish/ops-cli/issues
- **Wiki文档**: https://github.com/alexclownfish/ops-cli/wiki

---

## 📄 许可证

MIT License - Copyright (c) 2026 ops-cli

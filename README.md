

# ops-cli

一个集成了SSH管理、密码管理、OpenStack集成的运维工具集

## ✨ 功能特性

### SSH管理
- SSH单机/批量命令执行
- 支持密码和密钥认证
- 支持配置文件批量管理
- 可配置并发数
- **联动密码库**：exec/batch 可直接从密码库读取账密，无需手动输入
- **SSH连接池**：批量执行时复用连接，提升性能

### 文件分发
- 单机/批量文件分发（SCP）
- 支持目录递归上传
- 联动密码库：自动从数据库读取账密

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
# 克隆仓库
git clone https://github.com/alexclownfish/ops-cli.git
cd ops-cli

# 编译
go build -o ops

# 安装
sudo mv ops /usr/local/bin/
```

---

## 🚀 快速开始

### 第一步：生成主密钥

**⚠️ 重要：主密钥用于加密所有密码，请妥善保管！**

```bash
# 生成32字节随机密钥
openssl rand -base64 32

# 输出示例：
# eQ9P1niDITFavopNKbgmxxmSc5jIEg5zKUvNnZTUpEs=

# 保存到环境变量（推荐）
export OPS_MASTER_KEY="eQ9P1niDITFavopNKbgmxxmSc5jIEg5zKUvNnZTUpEs="

# 或保存到配置文件
echo 'export OPS_MASTER_KEY="eQ9P1niDITFavopNKbgmxxmSc5jIEg5zKUvNnZTUpEs="' >> ~/.bashrc
source ~/.bashrc
```

**安全建议：**
- 将主密钥保存到安全的地方（如密码管理器）
- 不要将主密钥提交到Git仓库
- 定期更换主密钥（使用`reset-key`命令）

---

## 📖 使用指南

### 1. SSH命令执行

#### 单机执行

```bash
# 密码认证
ops exec "uptime" -H 192.168.1.100 -u root -p password

# 密钥认证
ops exec "df -h" -H 192.168.1.100 -u root -i ~/.ssh/id_rsa

# 加密密钥认证（密钥有密码）
ops exec "free -h" -H 192.168.1.100 -u root -i ~/.ssh/id_rsa -k keypass
```

#### 批量执行

```bash
# 命令行指定多台服务器
ops batch "uptime" \
  -L 192.168.1.100,192.168.1.101,192.168.1.102 \
  -U root \
  -P password

# 使用配置文件（推荐）
ops batch "df -h" -c servers.yaml
```

#### 配置文件示例

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
    key: /root/.ssh/id_rsa
    key_password: keypass  # 可选
  
  - name: db-server-1
    host: 192.168.1.102
    port: 22
    user: admin
    password: admin123
```

---

### 2. 密码管理

#### 2.1 基本操作

**生成强密码：**

```bash
ops passwd generate

# 输出：
# ✅ 生成密码: O#y-XQyEXl9I6@hQKIcMt9Zk
```

**保存密码（自动生成新密码）：**

```bash
ops passwd save vm-001 \
  -H 192.168.1.100 \
  -u root \
  --key $OPS_MASTER_KEY

# 输出：
# ✅ 生成密码: xxx
# ✅ 已保存
```

**查看密码：**

```bash
ops passwd show vm-001 --key $OPS_MASTER_KEY

# 输出：
# 服务器ID: vm-001
# 名称: vm-001
# 地址: 192.168.1.100
# 用户: root
# 密码: O#y-XQyEXl9I6@hQKIcMt9Zk
# 创建时间: 2026-03-10 14:30:25
```

**列出所有服务器：**

```bash
ops passwd list --key $OPS_MASTER_KEY

# 输出：
# 总共 3 台服务器:
#
# ID: vm-001
#   名称: vm-001
#   地址: 192.168.1.100
#   用户: root
#   更新: 2026-03-10 14:30:25
```

**删除服务器：**

```bash
ops passwd delete vm-001 --key $OPS_MASTER_KEY

# 输出：
# ✅ 已删除服务器: vm-001
```

#### 2.2 密码修改

**单机改密（智能选择方式）：**

```bash
ops passwd reset vm-001 --key $OPS_MASTER_KEY

# 输出：
# 生成新密码: xxx
# 尝试SSH方式改密...
# ✅ SSH改密成功
```

**智能改密流程：**
1. **优先SSH改密**：速度快，不依赖物理机
2. **自动回退virsh**：SSH失败时自动尝试，适合应急场景
3. **自动记录方式**：记录成功的改密方式，优化下次执行

**批量改密：**

```bash
ops passwd reset-batch --key $OPS_MASTER_KEY

# 输出：
# 开始批量改密，共 10 台服务器
# 
# 处理: vm-001 (192.168.1.100)
#   尝试SSH方式改密...
#   ✅ SSH改密成功
# 
# 处理: vm-002 (192.168.1.101)
#   尝试SSH方式改密...
#   ⚠️  SSH改密失败: ...
#   尝试virsh方式改密...
#   ✅ virsh改密成功
# 
# ...
# 
# ✅ 批量改密完成: 成功 10/10 台
```

---

### 3. OpenStack集成

#### 3.1 环境准备

**设置OpenStack环境变量：**

```bash
# 方式1：直接设置
export OS_AUTH_URL=http://keystone-admin.cty.os:10006/v3
export OS_USERNAME=admin
export OS_PASSWORD=your-password
export OS_PROJECT_NAME=admin
export OPS_MASTER_KEY="your-master-key"

# 方式2：创建环境变量文件（推荐）
cat > ~/.openstack_env << 'EOF'
export OS_AUTH_URL=http://keystone-admin.cty.os:10006/v3
export OS_USERNAME=admin
export OS_PASSWORD=your-password
export OS_PROJECT_NAME=admin
export OPS_MASTER_KEY="your-master-key"
EOF

# 使用时加载
source ~/.openstack_env
```

#### 3.2 导入虚拟机

**一键导入所有虚拟机：**

```bash
ops passwd import \
  --hypervisor-port 10000 \
  --hypervisor-user secure \
  --hypervisor-key /root/.ssh/id_rsa \
  --vm-user secure \
  --key $OPS_MASTER_KEY

# 输出：
# 正在连接OpenStack...
# ✅ 认证成功
# 正在获取物理机列表...
# ✅ 找到 3 台物理机
# 正在获取虚拟机列表...
# ✅ 找到 10 台虚拟机
#
# 导入: vm-001 (192.168.1.100)
#   ✅ 已导入
# 导入: vm-002 (192.168.1.101)
#   ✅ 已导入
# ...
# 
# 导入完成！成功 10/10 台
```

**参数说明：**
- `--hypervisor-port`：物理机SSH端口（默认22）
- `--hypervisor-user`：物理机SSH用户
- `--hypervisor-key`：物理机SSH密钥路径
- `--hypervisor-key-pass`：物理机SSH密钥密码（可选）
- `--vm-user`：虚拟机用户名
- `--key`：主密钥

**自动完成的操作：**
- ✅ 从OpenStack API获取所有虚拟机信息
- ✅ 自动获取虚拟机IP地址
- ✅ 自动获取OpenStack实例ID（用于virsh改密）
- ✅ 自动获取物理机IP（每个虚拟机对应的物理机）
- ✅ 自动生成24位强密码
- ✅ 自动保存到加密数据库

**导入后查看：**

```bash
ops passwd list --key $OPS_MASTER_KEY

# 输出：
# 总共 10 台服务器:
#
# ID: vm-001
#   名称: vm-001
#   地址: 192.168.1.100
#   用户: secure
#   虚拟机ID: instance-00001234  ← OpenStack实例ID
#   物理机IP: 10.6.4.37          ← 物理机IP
#   更新: 2026-03-10 14:30:25
```

---

### 4. 密码生命周期管理

#### 4.1 检查密码年龄

```bash
ops passwd check-age --key $OPS_MASTER_KEY --days 85

# 输出：
# 需要改密的服务器（2台）:
# - vm-001 (已使用86天)
# - vm-003 (已使用90天)
```

**参数说明：**
- `--days`：密码有效期（天），默认85天

#### 4.2 手动执行自动改密

```bash
ops passwd auto-rotate --key $OPS_MASTER_KEY --days 85

# 输出：
# 发现 2 台服务器需要改密
#
# 改密: vm-001
#   尝试SSH方式改密...
#   ✅ SSH改密成功
#
# 改密: vm-003
#   尝试SSH方式改密...
#   ⚠️  SSH改密失败: ...
#   尝试virsh方式改密...
#   ✅ virsh改密成功
#
# ✅ 改密完成: 成功 2/2 台
```

#### 4.3 配置定时任务（完全自动化）

**步骤1：复制脚本**

```bash
# 从项目根目录复制
cp password-rotate.sh /usr/local/bin/
chmod +x /usr/local/bin/password-rotate.sh
```

**步骤2：配置环境变量**

```bash
# 方式1：在cron中直接指定
# 见步骤3

# 方式2：在系统环境变量中设置
echo 'export OPS_MASTER_KEY="your-master-key"' >> /etc/environment
```

**步骤3：配置cron定时任务**

```bash
# 创建cron配置文件
cat > /etc/cron.d/ops-password-rotate << 'EOF'
# 每天凌晨2点执行密码轮换
0 2 * * * root OPS_MASTER_KEY="your-master-key" /usr/local/bin/password-rotate.sh
EOF

# 设置权限
chmod 644 /etc/cron.d/ops-password-rotate

# 重启cron服务
systemctl restart cron  # Debian/Ubuntu
# 或
systemctl restart crond  # CentOS/RHEL
```

**自动化流程：**
1. ✅ 每天凌晨2点自动检查密码年龄
2. ✅ 自动改密到期的服务器（SSH优先，virsh回退）
3. ✅ 自动导出KDBX备份到 `/backup/passwords/passwords-YYYYMMDD.kdbx`
4. ✅ 自动清理30天前的备份
5. ✅ 详细日志记录到 `/var/log/ops-rotate.log`

**查看日志：**

```bash
# 实时查看日志
tail -f /var/log/ops-rotate.log

# 查看最近的执行记录
tail -n 100 /var/log/ops-rotate.log

# 查看今天的日志
grep "$(date +%Y-%m-%d)" /var/log/ops-rotate.log
```

**日志示例：**

```
[2026-03-10 02:00:01] =========================================
[2026-03-10 02:00:01] 开始密码轮换任务
[2026-03-10 02:00:01] =========================================
[2026-03-10 02:00:01] 配置信息：
[2026-03-10 02:00:01]   - 密码有效期: 85天
[2026-03-10 02:00:01]   - 备份目录: /backup/passwords
[2026-03-10 02:00:01]   - 日志文件: /var/log/ops-rotate.log
[2026-03-10 02:00:01] 正在检查密码年龄...
[2026-03-10 02:00:02] 需要改密的服务器（2台）:
[2026-03-10 02:00:02] - vm-001 (已使用86天)
[2026-03-10 02:00:02] - vm-003 (已使用90天)
[2026-03-10 02:00:02] 发现需要改密的服务器，开始执行改密...
[2026-03-10 02:00:05] 改密完成: 成功 2/2 台
[2026-03-10 02:00:05] 改密执行成功
[2026-03-10 02:00:05] 正在导出备份到: /backup/passwords/passwords-20260310.kdbx
[2026-03-10 02:00:06] 备份导出成功: /backup/passwords/passwords-20260310.kdbx
[2026-03-10 02:00:06] 正在清理30天前的备份...
[2026-03-10 02:00:06] 已删除旧备份:
[2026-03-10 02:00:06]   - /backup/passwords/passwords-20260208.kdbx
[2026-03-10 02:00:06] =========================================
[2026-03-10 02:00:06] 密码轮换任务完成
[2026-03-10 02:00:06] =========================================
```

#### 联动密码库执行

```bash
# 单机：从密码库读取账密，无需手动输入密码
ops exec "uptime" --from-db vm-001 --master-key $OPS_MASTER_KEY

# 批量：自动加载数据库所有服务器
ops batch "df -h" --from-db --master-key $OPS_MASTER_KEY

# 批量：指定数据库路径
ops batch "free -h" --from-db --master-key $OPS_MASTER_KEY --db /path/to/passwords.db
```

---

### 7. 文件分发

#### 单机分发

```bash
# 密码认证
ops scp ./app.tar.gz -H 192.168.1.100 -u root -p password -r /tmp/

# 密钥认证
ops scp ./app.tar.gz -H 192.168.1.100 -u root -i ~/.ssh/id_rsa -r /opt/app/

# 从密码库读取账密
ops scp ./app.tar.gz --from-db vm-001 --master-key $OPS_MASTER_KEY -r /tmp/
```

#### 批量分发

```bash
# 批量分发到指定服务器列表
ops scp ./config.yaml -L 192.168.1.100,192.168.1.101 --batch-user root --batch-pass password -r /etc/app/

# 批量分发到密码库所有服务器（推荐）
ops scp ./app.tar.gz --all-from-db --master-key $OPS_MASTER_KEY -r /tmp/

# 目录递归上传
ops scp ./dist/ --all-from-db --master-key $OPS_MASTER_KEY -r /opt/app/

# 控制并发数
ops scp ./app.tar.gz --all-from-db --master-key $OPS_MASTER_KEY -r /tmp/ --parallel 20
```

---

**导出密码到KeePassXC：**

```bash
ops passwd export-keepass \
  --kdbx-password "" \
  --key $OPS_MASTER_KEY \
  --output passwords.kdbx

# 输出：
# ✅ 已导出 10 台服务器到 passwords.kdbx
```

**在KeePassXC中打开：**
1. 打开KeePassXC应用
2. 文件 → 打开数据库
3. 选择 `passwords.kdbx`
4. 输入主密钥作为数据库密码
5. 所有密码将显示在KeePassXC中

**导出格式：**
- 标题：服务器ID
- 用户名：SSH用户
- 密码：服务器密码
- URL：服务器IP地址
- 备注：创建时间、更新时间等

---

### 6. 主密钥管理

#### 6.1 更换主密钥

**⚠️ 重要：更换主密钥会重新加密所有密码，请确保操作正确！**

```bash
ops passwd reset-key \
  --old-key $OLD_MASTER_KEY \
  --new-key $NEW_MASTER_KEY \
  --db passwords.db

# 输出：
# ✅ 旧密钥验证成功
# 正在重新加密 10 台服务器的密码...
#   ✅ vm-001
#   ✅ vm-002
#   ✅ vm-003
#   ...
# ✅ 主密钥更换完成！
# ⚠️  请妥善保管新密钥，旧密钥已失效
```

**安全特性：**
- ✅ 必须提供旧密钥才能更换（防止未授权操作）
- ✅ 自动重新加密所有密码
- ✅ 原子操作（失败自动回滚）

**更换后验证：**

```bash
# 使用新密钥查看密码
ops passwd list --key $NEW_MASTER_KEY

# 旧密钥将无法使用
ops passwd list --key $OLD_MASTER_KEY
# 输出：❌ 打开数据库失败: 主密钥错误
```

---

## 📋 完整命令参考

### SSH命令

```bash
# 单机执行
ops exec <command> -H <host> -u <user> [-p <password>] [-i <key>] [-k <key-pass>]

# 批量执行
ops batch <command> -L <host1,host2,...> -U <user> [-P <password>] [-c <config>]
```

### 密码管理命令

```bash
ops passwd generate                                    # 生成强密码
ops passwd save <id> -H <host> -u <user> --key <key>  # 保存密码
ops passwd show <id> --key <key>                       # 查看密码
ops passwd list --key <key>                            # 列出所有服务器

ops passwd update <id> [options] --key <key>              # 更新服务器信息
ops passwd reset <id> --key <key>                      # 修改密码（单机）
ops passwd reset-batch --key <key>                     # 批量改密
ops passwd import --key <key> [options]                # 从OpenStack导入
ops passwd export --kdbx-password "" --key <key> --output <file> # 导出到KeePassXC
```

### 密码生命周期命令

```bash
ops passwd check-age --key <key> --days <days>     # 检查密码年龄
ops passwd auto-rotate --key <key> --days <days>   # 自动改密
```

### 主密钥管理命令

```bash
ops passwd reset-key --old-key <old> --new-key <new> --db <db>  # 更换主密钥
```


---

## 🔐 安全特性

### 密码生成规则
- **长度**：24位
- **字符集**：A-Z、a-z、0-9、#_-@~
- **验证**：必须包含所有类型字符
- **安全性**：使用crypto/rand生成（密码学安全随机数）

### 加密存储
- **算法**：AES-256-GCM（军事级加密）
- **存储**：BoltDB加密数据库
- **主密钥**：32字节随机密钥（Base64编码）
- **权限**：数据库文件权限600（仅所有者可读写）
- **验证**：主密钥SHA256哈希验证（防止未授权访问）

### 改密方式

#### SSH改密
- **命令**：`/usr/sbin/chpasswd`
- **权限**：需要sudo权限（建议配置免密sudo）
- **适用**：所有Linux服务器
- **优势**：速度快，不依赖物理机

#### virsh改密
- **命令**：`virsh set-user-password`
- **权限**：需要物理机访问权限
- **适用**：OpenStack虚拟机
- **优势**：无需知道当前密码，适合应急场景
- **要求**：虚拟机需安装qemu-guest-agent

---

## 💡 最佳实践

### 1. 主密钥管理

**生成和存储：**

```bash
# 生成主密钥
openssl rand -base64 32 > ~/.ops_master_key

# 设置权限（仅所有者可读）
chmod 600 ~/.ops_master_key

# 加载到环境变量
export OPS_MASTER_KEY=$(cat ~/.ops_master_key)

# 添加到shell配置（可选）
echo 'export OPS_MASTER_KEY=$(cat ~/.ops_master_key)' >> ~/.bashrc
```

**备份主密钥：**
- 将主密钥保存到密码管理器（如KeePassXC、1Password）
- 打印纸质备份，存放在安全的地方
- 不要将主密钥提交到Git仓库或发送到网络

**定期更换：**
- 建议每年更换一次主密钥
- 使用`reset-key`命令安全更换
- 更换后立即备份新密钥

### 2. 密码轮换策略

**推荐配置：**
- **有效期**：85天（给5天缓冲期，实际90天轮换）
- **检查频率**：每天凌晨2点
- **备份保留**：30天
- **日志保留**：90天

**分级管理：**
- **核心服务器**：60天轮换
- **普通服务器**：85天轮换
- **测试服务器**：120天轮换

### 3. OpenStack集成

**首次导入流程：**
1. 设置OpenStack环境变量
2. 执行`import`命令导入所有虚拟机
3. 使用`reset-batch`命令统一改密（virsh方式）
4. 配置免密sudo
5. 后续使用SSH方式改密

**物理机密钥管理：**
- 使用SSH密钥认证（更安全）
- 密钥设置密码保护
- 定期轮换物理机密钥

### 4. 备份策略

**自动备份：**
- 每次改密后自动导出KDBX
- 保留最近30天的备份
- 备份文件命名：`passwords-YYYYMMDD.kdbx`

**手动备份：**
```bash
# 定期手动备份
ops passwd export \
  --kdbx-password "keeppass-password" \
  --key $OPS_MASTER_KEY \
  --output backup-$(date +%Y%m%d).kdbx

# 备份到远程
scp backup-$(date +%Y%m%d).kdbx backup-server:/backup/
```

### 5. 日志管理

**日志轮转配置：**

```bash
# 创建logrotate配置
cat > /etc/logrotate.d/ops-rotate << 'EOF'
/var/log/ops-rotate.log {
    daily
    rotate 90
    compress
    missingok
    notifempty
    create 0640 root root
}
EOF
```

**日志监控：**
```bash
# 监控改密失败
grep "ERROR" /var/log/ops-rotate.log

# 统计改密成功率
grep "改密完成" /var/log/ops-rotate.log | tail -n 10
```

---

## 🐛 故障排查

### SSH改密失败

**问题1：Permission denied**

```
⚠️  SSH改密失败: 改密失败: Process exited with status 1
```

**原因：** 用户没有sudo权限或需要输入sudo密码

**解决方案：** 配置免密sudo

```bash
# 在虚拟机上执行
echo "secure ALL=(ALL) NOPASSWD: ALL" | sudo tee /etc/sudoers.d/secure
chmod 440 /etc/sudoers.d/secure

# 验证
sudo -l
```

---

**问题2：chpasswd: command not found**

```
⚠️  SSH改密失败: sudo: chpasswd: command not found
```

**原因：** 已修复，现在使用完整路径 `/usr/sbin/chpasswd`

**解决方案：** 更新到最新版本

```bash
wget https://github.com/alexclownfish/ops-cli/raw/master/ops -O ops
chmod +x ops
```

---

### virsh改密失败

**问题1：Guest agent is not responding**

```
⚠️  virsh改密失败: error: Guest agent is not responding
```

**原因：** 虚拟机未安装或未启动qemu-guest-agent

**解决方案：** 安装并启动qemu-guest-agent

```bash
# 在虚拟机内执行
# CentOS/RHEL
yum install -y qemu-guest-agent
systemctl start qemu-guest-agent
systemctl enable qemu-guest-agent

# Ubuntu/Debian
apt-get install -y qemu-guest-agent
systemctl start qemu-guest-agent
systemctl enable qemu-guest-agent

# 验证
systemctl status qemu-guest-agent
```

---

**问题2：Connection refused**

```
⚠️  virsh改密失败: 连接物理机失败: ssh: connect to host 10.6.4.37 port 10000: Connection refused
```

**原因：** 物理机SSH端口不正确或防火墙阻止

**解决方案：** 检查物理机SSH配置

```bash
# 检查物理机SSH端口
netstat -tlnp | grep sshd

# 检查防火墙
firewall-cmd --list-all

# 测试连接
ssh -p 10000 secure@10.6.4.37
```

---

### 主密钥错误

**问题：主密钥错误**

```
❌ 打开数据库失败: 主密钥错误
```

**原因：** 使用了错误的主密钥

**解决方案：**
1. 确认使用正确的主密钥
2. 检查环境变量是否正确设置
3. 如果主密钥丢失，需要重新导入数据

```bash
# 检查环境变量
echo $OPS_MASTER_KEY

# 如果主密钥丢失，重新导入
rm passwords.db
ops passwd import ...
```

---

### OpenStack导入失败

**问题：认证失败**

```
❌ 认证失败: Unauthorized
```

**原因：** OpenStack环境变量不正确

**解决方案：** 检查环境变量

```bash
# 检查环境变量
env | grep OS_

# 测试OpenStack连接
openstack server list
```

---

**问题：找不到虚拟机**

```
✅ 找到 0 台虚拟机
```

**原因：** 当前项目没有虚拟机或权限不足

**解决方案：** 检查项目和权限

```bash
# 列出所有项目
openstack project list

# 切换项目
export OS_PROJECT_NAME=your-project

# 列出虚拟机
openstack server list
```

---

### openstack环境实测

计划任务

```shell
[root@aqc-bqjaq-jumpserver-10e6e7e248 oppos]# cat /etc/cron.d/ops-password-rotate
# 每天凌晨2点执行密码轮换
5 17 * * * root /usr/local/bin/password-rotate.sh
```

脚本

'cat /usr/local/bin/password-rotate.sh'

```shell
[root@aqc-bqjaq-jumpserver-10e6e7e248 oppos]# cat /usr/local/bin/password-rotate.sh
#!/bin/bash
# password-rotate.sh - 自动密码轮换脚本
# 用途：定期检查密码年龄并自动改密

set -e

# ==================== 配置 ====================
export PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
cd /opt/oppos && . admin-openrc.sh
KEY="${OPS_MASTER_KEY}"
DAYS=80
BACKUP_DIR="/opt/oppos/backup/passwords"
LOG_FILE="/var/log/ops-rotate.log"
TODAY=$(date +%Y%m%d)

# ==================== 函数 ====================
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

log_error() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] [ERROR] $1" | tee -a "$LOG_FILE" >&2
}

# ==================== 主流程 ====================
log "========================================="
log "开始密码轮换任务"
log "========================================="

# 检查主密钥
if [ -z "$KEY" ]; then
    log_error "主密钥未设置，请设置 OPS_MASTER_KEY 环境变量"
    exit 1
fi

log "配置信息："
log "  - 密码有效期: ${DAYS}天"
log "  - 备份目录: ${BACKUP_DIR}"
log "  - 日志文件: ${LOG_FILE}"

# 检查密码年龄
log "正在检查密码年龄..."
CHECK_OUTPUT=$(/usr/local/bin/ops passwd check-age --key "$KEY" --days "$DAYS" 2>&1)
log "$CHECK_OUTPUT"

# 判断是否需要改密
if echo "$CHECK_OUTPUT" | grep -q "需要改密"; then
    log "发现需要改密的服务器，开始执行改密..."

    # 执行改密
    ROTATE_OUTPUT=$(/usr/local/bin/ops passwd auto-rotate --key "$KEY" --days "$DAYS" 2>&1)
    log "$ROTATE_OUTPUT"

    # 检查改密结果
    if echo "$ROTATE_OUTPUT" | grep -q "改密完成"; then
        log "改密执行成功"

        # 创建备份目录
        mkdir -p "$BACKUP_DIR"

        # 导出全量备份
        BACKUP_FILE="${BACKUP_DIR}/passwords-${TODAY}.kdbx"
        log "正在导出备份到: ${BACKUP_FILE}"

        if /usr/local/bin/ops passwd export --kdbx-password "pass-${TODAY}" --key "$KEY" --output "$BACKUP_FILE" 2>&1 | tee -a "$LOG_FILE"; then
            log "备份导出成功: ${BACKUP_FILE}"

            # 清理30天前的备份
            log "正在清理30天前的备份..."
            DELETED=$(find "$BACKUP_DIR" -name "passwords-*.kdbx" -mtime +30 -delete -print 2>&1)
            if [ -n "$DELETED" ]; then
                log "已删除旧备份:"
                echo "$DELETED" | while read file; do
                    log "  - $file"
                done
            else
                log "无需清理旧备份"
            fi
        else
            log_error "备份导出失败"
            exit 1
        fi
    else
        log_error "改密执行失败"
        exit 1
    fi
else
    log "所有服务器密码都在有效期内，无需改密"
fi

log "========================================="
log "密码轮换任务完成"
log "========================================="
```

日志输出
```
以一天过期时间作为测试的日志

[2026-03-11 17:05:01] 需要改密的服务器（1台）:
- yangwenzhe-test4 (已使用1天)
[2026-03-11 17:05:01] 发现需要改密的服务器，开始执行改密...
[2026-03-11 17:05:01] 发现 1 台服务器需要改密

改密: yangwenzhe-test4
  尝试SSH方式改密...
  ✅ SSH改密成功

✅ 改密完成: 成功 1/1 台
[2026-03-11 17:05:01] 改密执行成功
[2026-03-11 17:05:01] 正在导出备份到: /opt/oppos/backup/passwords/passwords-20260311.kdbx
✅ 已导出 6 台服务器到 /opt/oppos/backup/passwords/passwords-20260311.kdbx
[2026-03-11 17:05:01] 备份导出成功: /opt/oppos/backup/passwords/passwords-20260311.kdbx
[2026-03-11 17:05:01] 正在清理30天前的备份...
[2026-03-11 17:05:01] 无需清理旧备份
[2026-03-11 17:05:01] =========================================
[2026-03-11 17:05:01] 密码轮换任务完成
[2026-03-11 17:05:01] =========================================

```



## 📚 更多信息

- **GitHub**: https://github.com/alexclownfish/ops-cli
- **问题反馈**: https://github.com/alexclownfish/ops-cli/issues
- **Wiki文档**: https://github.com/alexclownfish/ops-cli/wiki

---

## 📄 许可证

MIT License

Copyright (c) 2026 ops-cli

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

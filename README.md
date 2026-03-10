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

### ✅ 密码生命周期管理
- **密码年龄检查** - 检查密码使用天数
- **自动改密** - 密码到期自动轮换
- **智能改密** - SSH优先，virsh回退
- **定时任务** - 配合cron实现自动化

### ✅ 安全特性
- **主密钥验证** - 防止未授权访问
- **密钥更换** - 安全更换主密钥
- **OpenStack集成** - 自动导入虚拟机

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

**5. 删除服务器：**
```bash
ops passwd delete vm-001 --key $MASTER_KEY
```

**6. 修改密码（单机）：**
```bash
ops passwd list --key $MASTER_KEY
```


```bash
# SSH方式改密
ops passwd reset vm-001 --key $MASTER_KEY

# virsh方式改密（OpenStack）
# 需要在保存时指定reset_method=virsh和物理机信息
```

**7. 批量改密：**
```bash
ops passwd reset-batch --key $MASTER_KEY
# 自动遍历所有服务器并改密
```

**8. 导出到KeePassXC：**
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
delete                  删除服务器
reset                   修改密码（单机）
reset-batch             批量改密
import                  从OpenStack导入虚拟机
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

## 📖 详细使用手册

### 一、密码管理完整工作流程

#### 步骤1：初始化（首次使用）

**1.1 生成主密钥**
```bash
# 方法1：使用openssl生成
MASTER_KEY=$(openssl rand -base64 32)
echo $MASTER_KEY

# 方法2：使用ops-cli生成（即将支持）
# ops passwd init-key

# 保存主密钥到安全位置
echo $MASTER_KEY > ~/.ops-master-key
chmod 600 ~/.ops-master-key
```

**1.2 设置环境变量（推荐）**
```bash
# 添加到 ~/.bashrc 或 ~/.zshrc
export OPS_MASTER_KEY="your-master-key-here"
export OPS_DB_PATH="$HOME/.ops-cli/passwords.db"

# 重新加载配置
source ~/.bashrc
```

#### 步骤2：添加服务器

**2.1 添加普通Linux服务器（SSH改密）**
```bash
ops passwd save web-01 \
  -H 192.168.1.100 \
  -u root \
  --key $OPS_MASTER_KEY

# 输出示例：
# ✅ 生成新密码
# ✅ 密码已保存到数据库
# 服务器: web-01
# 密码: xT9#mK2_pLqW8@nR5vYzA3Bc
```

**2.2 添加OpenStack虚拟机（virsh改密）**
```bash
# 注意：virsh改密需要手动编辑数据库或使用配置文件
# 当前版本需要先保存，然后手动修改数据库中的字段：
# - reset_method: "virsh"
# - instance_id: "虚拟机实例ID"
# - hypervisor_host: "物理机IP"
# - hypervisor_port: 22
# - hypervisor_user: "root"
# - hypervisor_pass: "物理机密码"
```

#### 步骤3：查看和管理密码

**3.1 查看单个服务器密码**
```bash
ops passwd show web-01 --key $OPS_MASTER_KEY

# 输出示例：
# 服务器ID: web-01
# 名称: web-01
# 地址: 192.168.1.100
# 用户: root
# 密码: xT9#mK2_pLqW8@nR5vYzA3Bc
# 创建时间: 2026-03-06 14:00:00
```

**3.2 列出所有服务器**
```bash
ops passwd list --key $OPS_MASTER_KEY

# 输出示例：
# 总共 3 台服务器:
#
# ID: web-01
#   名称: web-01
#   地址: 192.168.1.100
#   用户: root
#   更新: 2026-03-06 14:00:00
#
# ID: web-02
#   名称: web-02
#   地址: 192.168.1.101
#   用户: root
#   更新: 2026-03-06 14:05:00
```

#### 步骤4：修改密码

**4.1 单机改密**
```bash
ops passwd reset web-01 --key $OPS_MASTER_KEY

# 输出示例：
# 生成新密码: yH7@nK9_qMwP2#vR6xZaB4Cd
# 使用SSH方式改密...
# ✅ 改密成功
```

**4.2 批量改密**
```bash
ops passwd reset-batch --key $OPS_MASTER_KEY

# 输出示例：
# 开始批量改密，共 3 台服务器
#
# 处理: web-01 (192.168.1.100)
#   ✅ 改密成功
#
# 处理: web-02 (192.168.1.101)
#   ✅ 改密成功
#
# 处理: web-03 (192.168.1.102)
#   ✅ 改密成功
```

#### 步骤5：导出备份

**5.1 导出到KeePassXC**
```bash
ops passwd export \
  --key $OPS_MASTER_KEY \
  --kdbx-password "your-keepass-password" \
  --output ~/backup/passwords-$(date +%Y%m%d).kdbx

# 输出示例：
# ✅ 已导出 3 台服务器到 /home/user/backup/passwords-20260306.kdbx
```

**5.2 在KeePassXC中导入**
```bash
# 1. 打开KeePassXC
# 2. 文件 -> 打开数据库
# 3. 选择 passwords-20260306.kdbx
# 4. 输入密码: your-keepass-password
# 5. 完成！所有密码已导入
```

### 二、SSH管理详细说明

#### 2.1 单机执行

**基础用法**
```bash
# 密码认证
ops exec "uptime" -H 192.168.1.100 -u root -p password

# 密钥认证
ops exec "df -h" -H 192.168.1.100 -u root -i ~/.ssh/id_rsa

# 带密码的密钥
ops exec "free -m" -H 192.168.1.100 -u root -i ~/.ssh/id_rsa --key-pass mypass
```

**高级用法**
```bash
# 执行多条命令
ops exec "cd /var/log && tail -n 20 syslog" -H 192.168.1.100 -u root -p pass

# 执行脚本
ops exec "bash /tmp/deploy.sh" -H 192.168.1.100 -u root -i ~/.ssh/id_rsa

# 查看系统信息
ops exec "uname -a && cat /etc/os-release" -H 192.168.1.100 -u root -p pass
```

#### 2.2 批量执行

**方式1：命令行指定服务器列表**
```bash
ops batch "uptime" \
  -L 192.168.1.100,192.168.1.101,192.168.1.102 \
  -U root \
  -P password

# 输出示例：
# ✅ 192.168.1.100:
#  14:30:25 up 10 days,  3:45,  1 user,  load average: 0.15, 0.10, 0.08
# ---
# ✅ 192.168.1.101:
#  14:30:26 up 5 days,  2:30,  2 users,  load average: 0.25, 0.20, 0.15
# ---
```

**方式2：使用配置文件**
```bash
# 创建 servers.yaml
cat > servers.yaml << EOF
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
    password: password456
  - name: db-server
    host: 192.168.1.102
    port: 22
    user: admin
    password: admin123
EOF

# 执行批量命令
ops batch "df -h" -c servers.yaml
```

**方式3：使用密钥批量执行**
```bash
ops batch "systemctl status nginx" \
  -L server1,server2,server3 \
  -U root \
  -K ~/.ssh/id_rsa \
  --batch-key-pass mypass
```

**控制并发数**
```bash
# 默认并发10个
ops batch "uptime" -c servers.yaml

# 设置并发数为5
ops batch "uptime" -c servers.yaml --parallel 5

# 串行执行（并发数为1）
ops batch "systemctl restart nginx" -c servers.yaml --parallel 1
```

### 三、常见问题（FAQ）

#### Q1: 如何生成主密钥？
```bash
# 使用openssl生成32字节密钥
openssl rand -base64 32

# 或者使用Python
python3 -c "import secrets; print(secrets.token_urlsafe(32))"
```

#### Q2: 主密钥丢失了怎么办？
**答：** 主密钥丢失后，数据库中的密码将无法解密。建议：
- 将主密钥保存在多个安全位置
- 定期导出到KeePassXC备份
- 使用密码管理器（如1Password）保存主密钥

#### Q3: 如何备份密码数据库？
```bash
# 备份数据库文件
cp passwords.db passwords-backup-$(date +%Y%m%d).db

# 或导出到KeePassXC
ops passwd export --key $KEY --kdbx-password backup123 --output backup.kdbx
```

#### Q4: 批量改密失败了部分服务器怎么办？
**答：** 
- 查看错误日志，确认失败原因
- 对失败的服务器单独执行 `ops passwd reset`
- 检查网络连接和SSH权限

#### Q5: 如何在OpenStack环境使用virsh改密？
**答：** 当前版本需要手动配置，步骤：
1. 先用 `ops passwd save` 保存服务器
2. 手动编辑数据库，添加virsh相关字段
3. 执行 `ops passwd reset` 自动识别virsh方式

#### Q6: 密码数据库可以多人共享吗？
**答：** 可以，但需要注意：
- 所有人使用相同的主密钥
- 数据库文件权限设置为600
- 建议使用Git管理数据库（加密后）

#### Q7: 如何迁移到新机器？
```bash
# 在旧机器上
cp passwords.db /path/to/backup/
echo $OPS_MASTER_KEY > master-key.txt

# 在新机器上
cp /path/to/backup/passwords.db ~/.ops-cli/
export OPS_MASTER_KEY=$(cat master-key.txt)
ops passwd list --key $OPS_MASTER_KEY
```

#### Q8: SSH连接超时怎么办？
**答：** 
- 检查网络连接：`ping <host>`
- 检查SSH端口：`telnet <host> 22`
- 增加超时时间（当前固定10秒，未来版本支持配置）
- 检查防火墙规则

### 四、最佳实践

#### 4.1 主密钥管理
```bash
# ✅ 推荐：使用环境变量
export OPS_MASTER_KEY="your-key"

# ✅ 推荐：保存到文件（权限600）
echo "your-key" > ~/.ops-master-key
chmod 600 ~/.ops-master-key
export OPS_MASTER_KEY=$(cat ~/.ops-master-key)

# ❌ 不推荐：直接在命令行输入（会留在历史记录）
ops passwd list --key "your-key-in-plain-text"
```

#### 4.2 密码轮换策略
```bash
# 定期改密（建议每90天）
# 方法1：手动执行
ops passwd reset-batch --key $KEY

# 方法2：设置cron任务（未来版本支持）
# 0 2 1 */3 * ops passwd reset-batch --key $KEY
```

#### 4.3 备份策略
```bash
# 每周备份一次
#!/bin/bash
DATE=$(date +%Y%m%d)
ops passwd export \
  --key $OPS_MASTER_KEY \
  --kdbx-password "backup-password" \
  --output ~/backups/passwords-$DATE.kdbx

# 保留最近4周的备份
find ~/backups -name "passwords-*.kdbx" -mtime +28 -delete
```

#### 4.4 安全建议
```bash
# 1. 数据库文件权限
chmod 600 passwords.db

# 2. 主密钥文件权限
chmod 600 ~/.ops-master-key

# 3. 定期更换主密钥（高级操作，需要重新加密所有密码）
# 当前版本不支持，建议导出后重新导入

# 4. 审计日志（未来版本支持）
# ops passwd audit --since 7d
```

#### 4.5 团队协作
```bash
# 1. 使用Git管理配置文件
git init
git add servers.yaml
git commit -m "Add server config"

# 2. 不要提交密码数据库
echo "passwords.db" >> .gitignore
echo "*.kdbx" >> .gitignore

# 3. 团队共享主密钥（使用密码管理器）
# 将主密钥保存在1Password/LastPass等团队密码管理器中
```

### 五、故障排查

#### 5.1 常见错误

**错误1：连接失败**
```bash
❌ 连接失败: dial tcp 192.168.1.100:22: i/o timeout

# 解决方法：
# 1. 检查网络连接
ping 192.168.1.100

# 2. 检查SSH端口
telnet 192.168.1.100 22

# 3. 检查防火墙
sudo iptables -L | grep 22
```

**错误2：认证失败**
```bash
❌ 连接失败: ssh: handshake failed: ssh: unable to authenticate

# 解决方法：
# 1. 检查密码是否正确
# 2. 检查用户名是否正确
# 3. 检查密钥文件权限
chmod 600 ~/.ssh/id_rsa

# 4. 检查密钥是否正确
ssh -i ~/.ssh/id_rsa root@192.168.1.100
```

**错误3：改密失败**
```bash
❌ 改密失败: chpasswd: (user root) pam_chauthtok() failed

# 解决方法：
# 1. 检查是否有root权限
# 2. 检查PAM配置
# 3. 手动测试chpasswd命令
echo "root:newpassword" | sudo chpasswd
```

**错误4：数据库打开失败**
```bash
❌ 打开数据库失败: timeout

# 解决方法：
# 1. 检查数据库文件是否存在
ls -la passwords.db

# 2. 检查文件权限
chmod 600 passwords.db

# 3. 检查是否被其他进程占用
lsof passwords.db
```

#### 5.2 调试技巧

**启用详细日志（未来版本支持）**
```bash
# ops passwd reset vm-001 --key $KEY --verbose
# ops batch "uptime" -c servers.yaml --debug
```

**测试SSH连接**
```bash
# 使用系统ssh命令测试
ssh -v root@192.168.1.100

# 测试密钥认证
ssh -i ~/.ssh/id_rsa -v root@192.168.1.100
```

### 六、高级用法

#### 6.1 脚本集成

**Bash脚本示例**
```bash
#!/bin/bash
# auto-deploy.sh - 自动部署脚本

set -e

# 配置
MASTER_KEY=$(cat ~/.ops-master-key)
SERVERS="web-01,web-02,web-03"

# 1. 批量执行部署命令
echo "开始部署..."
ops batch "cd /app && git pull && systemctl restart app" \
  -L $SERVERS \
  -U deploy \
  -K ~/.ssh/deploy_key

# 2. 验证部署结果
echo "验证部署..."
ops batch "systemctl status app | grep Active" \
  -L $SERVERS \
  -U deploy \
  -K ~/.ssh/deploy_key

echo "部署完成！"
```

**Python脚本示例**
```python
#!/usr/bin/env python3
# auto-reset-password.py - 自动改密脚本

import subprocess
import os
from datetime import datetime

def reset_passwords():
    master_key = os.environ.get('OPS_MASTER_KEY')
    if not master_key:
        print("错误: 未设置OPS_MASTER_KEY环境变量")
        return
    
    # 执行批量改密
    result = subprocess.run([
        'ops', 'passwd', 'reset-batch',
        '--key', master_key
    ], capture_output=True, text=True)
    
    # 记录日志
    with open('password-reset.log', 'a') as f:
        f.write(f"{datetime.now()}: {result.stdout}\n")
    
    print("改密完成！")

if __name__ == '__main__':
    reset_passwords()
```

#### 6.2 定时任务

**Cron任务示例**
```bash
# 编辑crontab
crontab -e

# 每月1号凌晨2点执行批量改密
0 2 1 * * /usr/local/bin/ops passwd reset-batch --key $(cat ~/.ops-master-key) >> /var/log/ops-reset.log 2>&1

# 每周日凌晨3点备份密码数据库
0 3 * * 0 /usr/local/bin/ops passwd export --key $(cat ~/.ops-master-key) --kdbx-password backup123 --output ~/backups/passwords-$(date +\%Y\%m\%d).kdbx
```

#### 6.3 与其他工具集成

**与Ansible集成**
```yaml
# playbook.yml
- name: 使用ops-cli批量改密
  hosts: localhost
  tasks:
    - name: 执行批量改密
      shell: ops passwd reset-batch --key {{ master_key }}
      register: result
    
    - name: 显示结果
      debug:
        msg: "{{ result.stdout }}"
```

**与Jenkins集成**
```groovy
// Jenkinsfile
pipeline {
    agent any
    environment {
        OPS_MASTER_KEY = credentials('ops-master-key')
    }
    stages {
        stage('Reset Passwords') {
            steps {
                sh 'ops passwd reset-batch --key $OPS_MASTER_KEY'
            }
        }
    }
}
```

### virsh改密详细步骤（OpenStack环境）

#### 适用场景
- OpenStack虚拟机密码忘记
- 需要批量重置虚拟机密码
- 无法通过SSH登录虚拟机

#### 前置条件
1. 有物理机（宿主机）的SSH访问权限
2. 知道虚拟机的实例ID（instance ID）
3. 物理机上已安装libvirt和virsh工具

#### 步骤1：查找虚拟机所在的物理机

**方法1：通过OpenStack API查询**
```bash
# 使用OpenStack CLI
openstack server show <vm-uuid> -f json | jq -r '.["OS-EXT-SRV-ATTR:hypervisor_hostname"]'

# 或使用nova命令
nova show <vm-uuid> | grep hypervisor_hostname
```

**方法2：通过配置文件映射**
```bash
# 如果有物理机IP映射表
# 例如：hypervisor-01 -> 172.28.188.89
```

#### 步骤2：获取虚拟机实例ID

**在物理机上查询：**
```bash
# SSH登录物理机
ssh root@172.28.188.89

# 列出所有虚拟机
virsh list --all

# 输出示例：
# Id    Name                           State
# ----------------------------------------------------
# 1     instance-00000001              running
# 2     instance-00000002              running
```

**实例ID就是Name列的值**（例如：instance-00000001）

#### 步骤3：使用ops-cli保存服务器信息

**当前版本需要手动配置数据库：**

1. 先保存基本信息：
```bash
ops passwd save vm-openstack-001 \
  -H 192.168.1.100 \
  -u root \
  --key $OPS_MASTER_KEY
```

2. 手动编辑数据库添加virsh信息（使用BoltDB工具或直接修改代码）：
```json
{
  "id": "vm-openstack-001",
  "name": "vm-openstack-001",
  "host": "192.168.1.100",
  "user": "root",
  "reset_method": "virsh",
  "instance_id": "instance-00000001",
  "hypervisor_host": "172.28.188.89",
  "hypervisor_port": 22,
  "hypervisor_user": "root",
  "hypervisor_pass": "物理机密码"
}
```

#### 步骤4：执行改密

```bash
# 单机改密
ops passwd reset vm-openstack-001 --key $OPS_MASTER_KEY

# 输出示例：
# 生成新密码: xT9#mK2_pLqW8@nR5vYzA3Bc
# 使用virsh方式改密...
# ✅ 改密成功
```

#### 步骤5：验证新密码

```bash
# 使用新密码SSH登录虚拟机
ssh root@192.168.1.100

# 或查看保存的密码
ops passwd show vm-openstack-001 --key $OPS_MASTER_KEY
```

#### 工作原理

virsh改密的底层命令：
```bash
# 在物理机上执行
virsh set-user-password <instance-id> <username> <new-password>

# 例如：
virsh set-user-password instance-00000001 root "xT9#mK2_pLqW8@nR5vYzA3Bc"
```

#### 注意事项

1. **权限要求**：需要物理机root权限
2. **虚拟机状态**：虚拟机必须处于running状态
3. **QEMU Guest Agent**：虚拟机内需要安装qemu-guest-agent
4. **密码复杂度**：某些系统可能有密码策略限制
5. **安全性**：物理机密码需要妥善保管

#### 故障排查

**问题1：virsh命令找不到**
```bash
# 安装libvirt
yum install libvirt libvirt-client  # CentOS/RHEL
apt install libvirt-clients          # Ubuntu/Debian
```

**问题2：改密失败 - QEMU guest agent is not connected**
```bash
# 在虚拟机内安装qemu-guest-agent
yum install qemu-guest-agent         # CentOS/RHEL
apt install qemu-guest-agent         # Ubuntu/Debian

# 启动服务
systemctl start qemu-guest-agent
systemctl enable qemu-guest-agent
```

**问题3：找不到虚拟机实例**
```bash
# 检查实例ID是否正确
virsh list --all | grep instance

# 检查是否在正确的物理机上
nova show <vm-uuid> | grep hypervisor
```

#### 未来改进

计划在下个版本中添加：
- 自动查询物理机位置（通过OpenStack API）
- 图形化配置virsh参数
- 批量virsh改密支持
- 改密前自动检查qemu-guest-agent状态


### OpenStack批量导入（新功能）

**从OpenStack自动导入所有虚拟机：**

```bash
# 步骤1：设置OpenStack环境变量
export OS_AUTH_URL=http://keystone-admin.cty.os:10006/v3
export OS_USERNAME=admin
export OS_PASSWORD=your-password
export OS_PROJECT_NAME=admin
export OPS_MASTER_KEY="your-master-key"

# 步骤2：一键导入所有虚拟机
ops passwd import \
  --hypervisor-port 10000 \
  --hypervisor-user secure \
  --hypervisor-key /root/.ssh/id_rsa \
  --vm-user secure \
  --key $OPS_MASTER_KEY

# 输出示例：
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
# 导入完成！成功 10/10 台
```

**自动完成：**
- ✅ 从OpenStack API获取所有虚拟机信息
- ✅ 自动获取虚拟机IP地址
- ✅ 自动获取虚拟机实例ID
- ✅ 自动获取物理机IP（无需手动指定）
- ✅ 自动生成密码并保存到数据库

**优势：**
- 不再需要手动复制粘贴虚拟机信息
- 不再需要查询物理机IP
- 一条命令导入所有虚拟机
- 支持大规模虚拟机批量管理



## 🔄 密码生命周期管理

### 检查密码年龄

```bash
ops passwd check-age --key $MASTER_KEY --days 85
```

**输出示例：**
```
需要改密的服务器（2台）:
- yangwenzhe-test1 (已使用86天)
- yangwenzhe-test3 (已使用90天)
```

### 自动改密

```bash
ops passwd auto-rotate --key $MASTER_KEY --days 85
```

**输出示例：**
```
发现 2 台服务器需要改密

改密: yangwenzhe-test1
  尝试SSH方式改密...
  ✅ SSH改密成功

改密: yangwenzhe-test3
  尝试SSH方式改密...
  ⚠️  SSH改密失败: ...
  尝试virsh方式改密...
  ✅ virsh改密成功

✅ 改密完成: 成功 2/2 台
```

### 配置定时任务（完全自动化）

**步骤1：复制脚本**
```bash
cp password-rotate.sh /usr/local/bin/
chmod +x /usr/local/bin/password-rotate.sh
```

**步骤2：设置环境变量**
```bash
# 在 /etc/environment 或 ~/.bashrc 中添加
export OPS_MASTER_KEY="your-master-key-here"
```

**步骤3：配置cron**
```bash
# 每天凌晨2点执行
echo "0 2 * * * root /usr/local/bin/password-rotate.sh >> /var/log/ops-rotate.log 2>&1" > /etc/cron.d/ops-password-rotate
```

**自动化流程：**
1. ✅ 每天自动检查密码年龄
2. ✅ 自动改密到期的服务器（SSH优先，virsh回退）
3. ✅ 自动导出KDBX备份到 `/backup/passwords/`
4. ✅ 自动清理30天前的备份

---

## 🔐 主密钥管理

### 更换主密钥

**安全更换主密钥（需要旧密钥验证）：**

```bash
ops passwd reset-key \
  --old-key $OLD_MASTER_KEY \
  --new-key $NEW_MASTER_KEY \
  --db passwords.db
```

**输出示例：**
```
✅ 旧密钥验证成功
正在重新加密 6 台服务器的密码...
  ✅ yangwenzhe-test1
  ✅ yangwenzhe-test2
  ✅ yangwenzhe-test3
  ✅ yangwenzhe-test4
  ✅ yangwenzhe-test5
  ✅ yangwenzhe-test6

✅ 主密钥更换完成！
⚠️  请妥善保管新密钥，旧密钥已失效
```

**安全特性：**
- ✅ 必须提供旧密钥才能更换
- ✅ 自动重新加密所有密码
- ✅ 防止未授权更换密钥

---

## 🎯 智能改密方案

**自动选择最佳改密方式：**

1. **优先SSH改密**
   - 速度快
   - 不依赖物理机
   - 适合日常使用

2. **自动回退virsh**
   - SSH失败时自动尝试
   - 适合应急场景
   - 无需原密码

3. **自动记录方式**
   - 记录成功的改密方式
   - 优化下次执行

**使用示例：**
```bash
# 单机改密（自动选择方式）
ops passwd reset yangwenzhe-test1 --key $MASTER_KEY

# 输出：
# 生成新密码: xxx
# 尝试SSH方式改密...
# ✅ SSH改密成功
```


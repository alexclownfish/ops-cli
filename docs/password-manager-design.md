# 密码管理器设计方案

## 一、需求概述

### 1.1 核心功能
- 生成符合规则的强密码
- 修改虚拟机密码（支持多种平台）
- 加密存储密码
- 解密查看密码
- 支持单机/批量操作

### 1.2 密码规则
- 字符集：A-Z、a-z、0-9、#_-@~
- 长度：24位
- 必须包含至少一个大写、小写、数字、特殊字符

### 1.3 使用场景
- OpenStack虚拟机（virsh）
- 云平台虚拟机（阿里云、腾讯云等）
- 物理机/普通虚拟机（SSH）
- VMware虚拟机

---

## 二、技术方案

### 2.1 密码生成

**算法：**
```
1. 使用 crypto/rand 生成随机字节
2. 从字符集中随机选择24个字符
3. 验证是否包含所有类型字符
4. 不满足则重新生成
```

**Go实现：**
```go
func GeneratePassword() string {
    charset := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789#_-@~"
    // 使用crypto/rand确保安全性
}
```

### 2.2 密码加密存储

**方案选择：AES-256-GCM**

**优点：**
- 安全性高（军事级加密）
- Go标准库支持
- 带认证的加密（防篡改）

**密钥管理：**
```
方案1：主密钥文件（推荐）
- 生成32字节主密钥
- 存储在 ~/.ops-cli/master.key
- 权限设置为 600

方案2：环境变量
- export OPS_MASTER_KEY="xxx"
- 适合CI/CD环境

方案3：密码派生
- 用户输入密码
- 使用PBKDF2派生密钥
- 更安全但每次需要输入
```

**存储格式：**
```json
{
  "version": "1.0",
  "servers": [
    {
      "id": "vm-001",
      "name": "web-server-1",
      "host": "192.168.1.100",
      "user": "root",
      "password_encrypted": "base64(encrypted_data)",
      "platform": "openstack",
      "updated_at": "2026-03-06T10:00:00Z"
    }
  ]
}
```

### 2.3 改密方案

#### 2.3.1 OpenStack (virsh)
```bash
# 在物理机上执行
virsh set-user-password <instance_id> <user> <password>
```

**优点：** 不需要知道原密码  
**缺点：** 需要物理机访问权限

#### 2.3.2 SSH方式（通用）
```bash
# 方式1：使用chpasswd（推荐）
echo "username:newpassword" | chpasswd

# 方式2：使用passwd命令
echo -e "newpassword\nnewpassword" | passwd username

# 方式3：直接修改shadow文件
# 需要先生成密码hash
```

**优点：** 通用性强  
**缺点：** 需要root权限或sudo

#### 2.3.3 云平台API
```
阿里云：ResetInstancePassword
腾讯云：ResetInstancesPassword
AWS：ModifyInstanceAttribute
```

**优点：** 官方支持，安全  
**缺点：** 需要API密钥，平台相关

### 2.4 命令设计

#### 生成密码
```bash
ops passwd generate
ops passwd generate --length 32
ops passwd generate --save vm-001
```

#### 修改密码（单机）
```bash
# OpenStack方式
ops passwd reset vm-001 --method virsh --host 192.168.1.1

# SSH方式
ops passwd reset vm-001 --method ssh --host 192.168.1.100 -u root -i ~/.ssh/id_rsa

# 云平台API
ops passwd reset vm-001 --method aliyun --instance-id i-xxx
```

#### 批量修改
```bash
ops passwd reset-batch --config servers.yaml
ops passwd reset-batch --hosts vm-001,vm-002,vm-003
```

#### 查看密码
```bash
ops passwd show vm-001
ops passwd show --all
ops passwd export vm-001 --output plain
```

---

## 三、实现步骤

### 3.1 第一阶段（核心功能）
1. 密码生成器
2. AES加密/解密模块
3. 本地存储（JSON文件）
4. SSH改密（通用方式）

### 3.2 第二阶段（OpenStack支持）
1. virsh改密支持
2. OpenStack API集成
3. 批量操作

### 3.3 第三阶段（云平台扩展）
1. 阿里云API
2. 腾讯云API
3. AWS API

---

## 四、安全建议

### 4.1 密钥管理
- 主密钥文件权限必须是600
- 定期轮换主密钥
- 不要将密钥提交到Git

### 4.2 密码存储
- 加密文件权限600
- 定期备份
- 审计日志记录所有操作

### 4.3 改密操作
- 改密前备份旧密码
- 改密后验证新密码可用
- 失败自动回滚

---

## 五、技术选型

### 5.1 加密库
- **crypto/aes** - Go标准库，AES加密
- **crypto/rand** - 安全随机数生成
- **encoding/base64** - Base64编码

### 5.2 存储方案对比

| 方案 | 优点 | 缺点 | 推荐度 |
|------|------|------|--------|
| JSON文件 | 简单，易调试 | 并发写入问题 | ⭐⭐⭐ |
| SQLite | 支持查询，事务 | 需要额外依赖 | ⭐⭐⭐⭐ |
| BoltDB | 纯Go，高性能 | 学习成本 | ⭐⭐⭐⭐⭐ |

**推荐：BoltDB**
- 纯Go实现，无CGO依赖
- 支持事务
- 性能优秀
- 单文件存储

### 5.3 改密方式优先级

```
1. virsh (OpenStack) - 最可靠，不需要原密码
2. SSH + chpasswd - 通用性强
3. 云平台API - 官方支持
4. SSH + passwd - 备选方案
```

---

## 六、风险评估

### 6.1 安全风险
| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 主密钥泄露 | 高 | 文件权限600，定期轮换 |
| 密码文件被盗 | 中 | 加密存储，审计日志 |
| 改密失败导致无法登录 | 高 | 改密前备份，失败回滚 |
| 批量改密部分失败 | 中 | 事务处理，详细日志 |

### 6.2 技术风险
| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| virsh命令不可用 | 中 | 提供多种改密方式 |
| SSH连接失败 | 中 | 重试机制，超时控制 |
| 并发写入冲突 | 低 | 使用数据库事务 |

---

## 七、实现建议

### 7.1 最小可行方案（MVP）
**第一周实现：**
1. 密码生成器
2. AES加密/解密
3. JSON文件存储
4. SSH改密（chpasswd方式）
5. 单机操作

**命令：**
```bash
ops passwd generate --save vm-001
ops passwd reset vm-001 --host 192.168.1.100
ops passwd show vm-001
```

### 7.2 完整方案（2-3周）
**增强功能：**
1. BoltDB存储
2. virsh改密支持
3. 批量操作
4. 改密验证和回滚
5. 审计日志

### 7.3 扩展方案（1个月+）
**云平台集成：**
1. 阿里云API
2. 腾讯云API
3. AWS API
4. Web管理界面

---

## 八、总结

### 8.1 推荐方案
**存储：** BoltDB（加密）  
**改密：** virsh（OpenStack） + SSH（通用）  
**加密：** AES-256-GCM  
**密钥：** 主密钥文件 + 环境变量备选

### 8.2 开发优先级
1. **高优先级**：密码生成、加密存储、SSH改密
2. **中优先级**：virsh改密、批量操作、回滚机制
3. **低优先级**：云平台API、Web界面

### 8.3 预期效果
- 密码强度：24位混合字符，安全性高
- 改密速度：单机<5秒，批量100台<2分钟
- 存储安全：AES-256加密，军事级安全
- 易用性：命令行操作，简单直观

---

**文档版本：** v1.0  
**创建时间：** 2026-03-06  
**作者：** AI Assistant

---

## 九、问题解决方案

### 9.1 virsh改密的物理机定位问题

#### 问题分析
- 每个虚拟机在不同的物理机上
- 不同池子的物理机IP不同
- 需要先定位虚拟机在哪台物理机

#### 解决方案

**方案1：OpenStack API查询（推荐）**
```bash
# 通过OpenStack API获取虚拟机所在物理机
openstack server show <vm-id> -f json | jq '.["OS-EXT-SRV-ATTR:host"]'
```

**数据结构：**
```json
{
  "pools": [
    {
      "name": "pool-1",
      "openstack_api": "http://controller1:5000",
      "auth": {
        "username": "admin",
        "password": "xxx",
        "project": "admin"
      },
      "hypervisors": [
        {
          "hostname": "compute-01",
          "ip": "192.168.1.10",
          "ssh_user": "root",
          "ssh_key": "/root/.ssh/id_rsa"
        }
      ]
    }
  ]
}
```

**改密流程：**
```
1. 读取虚拟机ID和所属池子
2. 调用OpenStack API查询虚拟机所在物理机hostname
3. 从配置中查找该hostname对应的物理机IP
4. SSH登录物理机
5. 执行virsh set-user-password命令
```

**命令示例：**
```bash
# 配置文件包含池子和物理机映射
ops passwd reset vm-001 --method virsh --pool pool-1

# 自动流程：
# 1. 查询vm-001在哪台物理机 -> compute-01
# 2. 查找compute-01的IP -> 192.168.1.10
# 3. SSH到192.168.1.10执行virsh命令
```

**方案2：预配置映射表**
```yaml
# vm-hypervisor-mapping.yaml
mappings:
  - vm_id: vm-001
    instance_id: i-xxxxx
    hypervisor: compute-01
    hypervisor_ip: 192.168.1.10
  - vm_id: vm-002
    instance_id: i-yyyyy
    hypervisor: compute-02
    hypervisor_ip: 192.168.1.11
```

**优点：** 不依赖OpenStack API  
**缺点：** 需要手动维护映射关系

### 9.2 密码生命周期自动管理

#### 需求分析
- 密码有效期：90天
- 提前5天自动改密（第85天）
- 从加密存储读取当前密码
- 自动生成新密码并更新

#### 数据结构扩展
```json
{
  "servers": [
    {
      "id": "vm-001",
      "host": "192.168.1.100",
      "user": "root",
      "password_encrypted": "xxx",
      "password_created_at": "2026-01-01T00:00:00Z",
      "password_expires_at": "2026-04-01T00:00:00Z",
      "password_lifetime_days": 90,
      "auto_rotate": true
    }
  ]
}
```

#### 实现方案

**方案1：定时任务（推荐）**
```bash
# 添加cron任务
ops passwd cron-install

# 每天检查一次
0 2 * * * /usr/local/bin/ops passwd check-rotate
```

**check-rotate逻辑：**
```
1. 读取所有服务器配置
2. 计算密码剩余天数
3. 如果 <= 5天，触发自动改密
4. 生成新密码
5. 使用当前密码SSH登录
6. 执行改密命令
7. 验证新密码可用
8. 更新加密存储
9. 发送通知（邮件/钉钉）
```

**命令示例：**
```bash
# 手动检查并轮换
ops passwd check-rotate

# 查看即将过期的密码
ops passwd list-expiring --days 7

# 强制轮换指定服务器
ops passwd rotate vm-001 --force
```

**方案2：守护进程**
```bash
# 启动守护进程
ops passwd daemon start

# 后台运行，每小时检查一次
# 日志输出到 /var/log/ops-passwd-daemon.log
```

#### 自动改密流程

```
┌─────────────────┐
│  定时检查任务    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ 读取密码数据库   │
│ 计算剩余天数     │
└────────┬────────┘
         │
         ▼
    剩余 <= 5天？
         │
    ┌────┴────┐
    │   是    │   否 → 结束
    └────┬────┘
         │
         ▼
┌─────────────────┐
│ 解密当前密码     │
│ 生成新密码       │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ SSH登录虚拟机    │
│ (使用当前密码)   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ 执行改密命令     │
│ chpasswd        │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ 验证新密码       │
│ (新建SSH连接)   │
└────────┬────────┘
         │
    ┌────┴────┐
    │  成功？  │
    └────┬────┘
         │
    ┌────┴────┐
    │   是    │   否 → 回滚
    └────┬────┘
         │
         ▼
┌─────────────────┐
│ 更新数据库       │
│ 记录新密码       │
│ 更新时间戳       │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ 发送通知         │
│ (邮件/钉钉)     │
└─────────────────┘
```

#### 配置示例

**完整配置文件：**
```yaml
# password-config.yaml
global:
  password_lifetime_days: 90
  rotate_before_days: 5
  notification:
    enabled: true
    type: dingtalk
    webhook: "https://oapi.dingtalk.com/robot/send?access_token=xxx"

pools:
  - name: pool-1
    openstack_api: "http://controller1:5000/v3"
    auth:
      username: admin
      password: admin123
      project: admin
    hypervisors:
      - hostname: compute-01
        ip: 192.168.1.10
        ssh_user: root
        ssh_key: /root/.ssh/id_rsa

servers:
  - id: vm-001
    name: web-server-1
    pool: pool-1
    instance_id: i-xxxxx
    host: 192.168.1.100
    user: root
    auto_rotate: true
    rotate_method: ssh  # ssh | virsh
```

#### 注意事项

**1. 改密时机选择**
- 建议在业务低峰期（凌晨2-4点）
- 避免在备份、巡检时间段
- 分批执行，避免同时改密大量服务器

**2. 失败处理**
- 改密失败不影响当前密码
- 记录失败日志，人工介入
- 连续失败3次，暂停自动轮换

**3. 通知机制**
```
成功：发送简要通知
失败：发送详细错误信息
即将过期：提前7天提醒
```

**4. 审计日志**
```json
{
  "timestamp": "2026-03-06T02:00:00Z",
  "action": "password_rotate",
  "server_id": "vm-001",
  "old_password_age": 85,
  "result": "success",
  "method": "ssh"
}
```

---

## 十、问题解答总结

### 10.1 virsh改密物理机定位
**解决方案：** OpenStack API + 配置映射
- 通过API查询虚拟机所在物理机
- 配置文件维护物理机IP映射
- 自动SSH到物理机执行virsh命令

### 10.2 密码自动轮换
**解决方案：** 定时任务 + 生命周期管理
- 存储密码创建时间和有效期
- 定时检查（每天或每小时）
- 提前5天自动改密
- 使用当前密码登录后改密
- 验证成功后更新存储

**关键点：**
- 必须先用旧密码登录才能改密
- 改密后立即验证新密码
- 失败自动回滚或告警

---

## 十一、实施建议（更新）

### 11.1 MVP阶段（2周）
**核心功能：**
1. 密码生成器（24位强密码）
2. AES-256加密存储（BoltDB）
3. SSH改密（chpasswd方式）
4. 密码生命周期记录
5. 手动轮换命令

**命令：**
```bash
ops passwd generate --save vm-001
ops passwd reset vm-001 --host 192.168.1.100
ops passwd show vm-001
ops passwd list-expiring --days 7
```

### 11.2 完整版（4周）
**增强功能：**
1. OpenStack API集成（查询物理机）
2. virsh改密支持
3. 自动轮换（定时任务）
4. 批量操作
5. 钉钉通知

### 11.3 生产级（6周）
**企业功能：**
1. 守护进程模式
2. 审计日志
3. 权限管理
4. Web管理界面
5. 多池子支持

---

**文档版本：** v1.1  
**更新时间：** 2026-03-06 10:50  
**更新内容：** 
- 添加virsh物理机定位方案
- 添加密码自动轮换方案
- 完善配置文件结构

---

## 十二、KeePassXC导出支持

### 12.1 KeePassXC导入格式

KeePassXC支持多种导入格式：

**1. CSV格式（推荐）**
```csv
"Group","Title","Username","Password","URL","Notes"
"Root","web-server-1","root","password123","ssh://192.168.1.100","vm-001"
"Root","web-server-2","root","password456","ssh://192.168.1.101","vm-002"
```

**字段说明：**
- Group: 分组名称
- Title: 条目标题
- Username: 用户名
- Password: 密码（明文）
- URL: 连接地址
- Notes: 备注信息

**2. XML格式（KeePass 2.x）**
```xml
<?xml version="1.0" encoding="utf-8"?>
<database>
  <entry>
    <title>web-server-1</title>
    <username>root</username>
    <password>password123</password>
    <url>ssh://192.168.1.100</url>
    <notes>vm-001</notes>
  </entry>
</database>
```

### 12.2 实现方案

#### 命令设计
```bash
# 导出所有密码到CSV
ops passwd export --format keepass --output passwords.csv

# 导出指定服务器
ops passwd export vm-001,vm-002 --format keepass --output servers.csv

# 导出指定分组
ops passwd export --group production --format keepass --output prod.csv

# 导出为XML格式
ops passwd export --format keepass-xml --output passwords.xml
```

#### CSV生成逻辑
```go
func ExportToKeePassCSV(servers []Server, output string) error {
    file, _ := os.Create(output)
    defer file.Close()
    
    writer := csv.NewWriter(file)
    defer writer.Flush()
    
    // 写入表头
    writer.Write([]string{"Group", "Title", "Username", "Password", "URL", "Notes"})
    
    // 写入数据
    for _, srv := range servers {
        password := decrypt(srv.PasswordEncrypted)
        writer.Write([]string{
            srv.Group,           // 分组
            srv.Name,            // 标题
            srv.User,            // 用户名
            password,            // 明文密码
            "ssh://" + srv.Host, // URL
            srv.ID,              // 备注
        })
    }
    return nil
}
```

#### 数据结构扩展
```json
{
  "servers": [
    {
      "id": "vm-001",
      "name": "web-server-1",
      "group": "Production",
      "host": "192.168.1.100",
      "user": "root",
      "password_encrypted": "xxx",
      "tags": ["web", "nginx"]
    }
  ]
}
```

### 12.3 KeePassXC导入步骤

**1. 导出密码**
```bash
ops passwd export --format keepass --output /tmp/passwords.csv
```

**2. 在KeePassXC中导入**
```
1. 打开KeePassXC
2. 数据库 -> 导入 -> CSV文件
3. 选择 /tmp/passwords.csv
4. 映射字段：
   - 列1 -> Group
   - 列2 -> Title
   - 列3 -> Username
   - 列4 -> Password
   - 列5 -> URL
   - 列6 -> Notes
5. 导入完成
```

**3. 安全清理**
```bash
# 导入后立即删除CSV文件
shred -u /tmp/passwords.csv
```

### 12.4 安全建议

**1. 导出文件保护**
```bash
# 设置文件权限为仅所有者可读
chmod 600 passwords.csv

# 导出到加密分区
ops passwd export --format keepass --output /encrypted/passwords.csv
```

**2. 临时文件处理**
- 导出后立即导入KeePassXC
- 导入完成后使用 `shred` 安全删除
- 不要通过网络传输明文CSV

**3. 审计日志**
```json
{
  "timestamp": "2026-03-06T10:00:00Z",
  "action": "export_keepass",
  "user": "admin",
  "servers_count": 50,
  "output_file": "/tmp/passwords.csv"
}
```

### 12.5 高级功能

**分组导出**
```bash
# 按池子分组
ops passwd export --group-by pool --format keepass

# 按标签分组
ops passwd export --group-by tags --format keepass
```

**过滤导出**
```bash
# 只导出生产环境
ops passwd export --filter "group=Production" --format keepass

# 只导出特定标签
ops passwd export --filter "tags=web" --format keepass
```

### 12.6 实现优先级

**MVP阶段：**
- CSV格式导出（基础功能）
- 全量导出

**完整版：**
- 分组导出
- 过滤导出
- XML格式支持

**生产级：**
- 自动清理临时文件
- 导出审计日志
- 加密导出（GPG）

---

## 十三、最终功能清单

### 13.1 核心功能
- [x] 密码生成（24位强密码）
- [x] AES-256加密存储
- [x] SSH改密
- [x] virsh改密（OpenStack）
- [x] 密码生命周期管理
- [x] 自动轮换（85天触发）
- [x] KeePassXC导出

### 13.2 命令总览
```bash
# 密码生成
ops passwd generate --save vm-001

# 改密
ops passwd reset vm-001 --host 192.168.1.100
ops passwd reset vm-001 --method virsh --pool pool-1

# 批量改密
ops passwd reset-batch --config servers.yaml

# 查看密码
ops passwd show vm-001
ops passwd list-expiring --days 7

# 自动轮换
ops passwd check-rotate
ops passwd daemon start

# 导出
ops passwd export --format keepass --output passwords.csv
```

---

**文档版本：** v1.2  
**最后更新：** 2026-03-06 10:57  
**新增内容：** KeePassXC导出支持

---

## 十四、KeePass数据库格式（.kdbx）支持

### 14.1 格式说明

**KeePass数据库格式：**
- `.kdb` - KeePass 1.x格式（已过时）
- `.kdbx` - KeePass 2.x格式（推荐，KeePassXC使用）

**优势：**
- 直接导入，无需CSV中转
- 保留分组结构
- 支持附件、图标等高级功能
- 加密存储（AES-256）

### 14.2 技术实现

**Go语言库：**
```
github.com/tobischo/gokeepasslib/v3
```

**功能：**
- 创建.kdbx数据库
- 添加条目（Entry）
- 创建分组（Group）
- 设置主密码

### 14.3 实现示例

**Go代码示例：**
```go
import "github.com/tobischo/gokeepasslib/v3"

func ExportToKeePass(servers []Server, output, masterPassword string) error {
    // 创建新数据库
    db := gokeepasslib.NewDatabase()
    db.Options = gokeepasslib.NewDatabaseOptions()
    
    // 创建根分组
    rootGroup := gokeepasslib.NewGroup()
    rootGroup.Name = "ops-cli-passwords"
    
    // 添加条目
    for _, srv := range servers {
        entry := gokeepasslib.NewEntry()
        entry.Values = append(entry.Values, 
            gokeepasslib.ValueData{Key: "Title", Value: gokeepasslib.V{Content: srv.Name}},
            gokeepasslib.ValueData{Key: "UserName", Value: gokeepasslib.V{Content: srv.User}},
            gokeepasslib.ValueData{Key: "Password", Value: gokeepasslib.V{Content: decrypt(srv.PasswordEncrypted), Protected: true}},
            gokeepasslib.ValueData{Key: "URL", Value: gokeepasslib.V{Content: "ssh://" + srv.Host}},
            gokeepasslib.ValueData{Key: "Notes", Value: gokeepasslib.V{Content: srv.ID}},
        )
        rootGroup.Entries = append(rootGroup.Entries, entry)
    }
    
    db.Content.Root.Groups = append(db.Content.Root.Groups, rootGroup)
    
    // 加密并保存
    db.LockWithPassword(masterPassword)
    file, _ := os.Create(output)
    defer file.Close()
    
    encoder := gokeepasslib.NewEncoder(file)
    return encoder.Encode(db)
}
```

### 14.4 命令设计

**导出为.kdbx格式：**
```bash
# 导出为KeePass数据库（需要设置主密码）
ops passwd export --format kdbx --output passwords.kdbx --master-password "your-strong-password"

# 交互式输入主密码（更安全）
ops passwd export --format kdbx --output passwords.kdbx

# 按分组导出
ops passwd export --format kdbx --output prod.kdbx --group production
```

### 14.5 KeePassXC导入步骤

**1. 导出数据库**
```bash
ops passwd export --format kdbx --output /tmp/passwords.kdbx
# 输入主密码: ********
```

**2. 在KeePassXC中打开**
```
1. 打开KeePassXC
2. 文件 -> 打开数据库
3. 选择 /tmp/passwords.kdbx
4. 输入主密码
5. 完成！所有密码已导入
```

**优势：**
- 一步到位，无需CSV中转
- 保留分组结构
- 密码已加密，更安全

### 14.6 格式对比

| 格式 | 优点 | 缺点 | 推荐度 |
|------|------|------|--------|
| **CSV** | 简单，通用 | 明文，需手动映射字段 | ⭐⭐⭐ |
| **XML** | 结构化 | 明文，较复杂 | ⭐⭐ |
| **KDBX** | 加密，直接导入，保留结构 | 需要额外库 | ⭐⭐⭐⭐⭐ |

**推荐：KDBX格式**
- 最安全（加密传输）
- 最方便（一步导入）
- 最完整（保留分组）

### 14.7 实现优先级

**MVP阶段：**
- CSV格式（快速实现）

**完整版：**
- KDBX格式（推荐）
- 分组支持

**最终命令：**
```bash
# 推荐方式：导出为KDBX
ops passwd export --format kdbx --output passwords.kdbx

# 备选方式：导出为CSV
ops passwd export --format csv --output passwords.csv
```

---

**文档版本：** v1.3  
**最后更新：** 2026-03-06 11:00  
**新增内容：** KDBX格式支持（推荐）

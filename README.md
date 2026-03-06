# ops-cli

> 这是一个AI写的，也不知道能写成啥样 🤖

一个集成了常用运维功能的Go语言工具集

## 功能规划

### ✅ 已规划
1. **SSH批量执行命令** - 多服务器并发执行
2. **服务器监控告警** - CPU/内存/磁盘监控
3. **日志分析** - 关键词搜索和错误统计
4. **自动化部署** - 应用发布和回滚

### 🚧 开发中
- [ ] SSH批量执行命令

## 安装

```bash
go install github.com/alexclownfish/ops-cli@latest
```

## 使用

```bash
# 查看帮助
ops --help

# SSH批量执行命令（开发中）
ops server exec "uptime"
```

## 开发

```bash
git clone https://github.com/alexclownfish/ops-cli.git
cd ops-cli
go build
```

## License

MIT

#!/bin/bash
# password-rotate.sh - 自动密码轮换脚本

KEY="${OPS_MASTER_KEY}"
DAYS=85
BACKUP_DIR="/backup/passwords"
TODAY=$(date +%Y%m%d)

echo "[$(date)] 开始检查密码年龄..."

# 检查是否需要改密
NEED_ROTATE=$(ops passwd check-age --key $KEY --days $DAYS 2>&1 | grep "需要改密")

if [ -n "$NEED_ROTATE" ]; then
    echo "[$(date)] 发现需要改密的服务器"
    
    # 执行改密
    ops passwd auto-rotate --key $KEY --days $DAYS
    
    # 导出全量备份
    mkdir -p $BACKUP_DIR
    ops passwd export-keepass --key $KEY --output $BACKUP_DIR/passwords-$TODAY.kdbx
    
    echo "[$(date)] ✅ 改密完成，已导出备份: passwords-$TODAY.kdbx"
    
    # 清理30天前的备份
    find $BACKUP_DIR -name "passwords-*.kdbx" -mtime +30 -delete
else
    echo "[$(date)] ✅ 无需改密"
fi
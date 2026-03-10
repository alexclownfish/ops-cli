#!/bin/bash
# password-rotate.sh - 自动密码轮换脚本
# 用途：定期检查密码年龄并自动改密

set -e

# ==================== 配置 ====================
KEY="${OPS_MASTER_KEY}"
DAYS=85
BACKUP_DIR="/backup/passwords"
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
CHECK_OUTPUT=$(ops passwd check-age --key "$KEY" --days "$DAYS" 2>&1)
log "$CHECK_OUTPUT"

# 判断是否需要改密
if echo "$CHECK_OUTPUT" | grep -q "需要改密"; then
    log "发现需要改密的服务器，开始执行改密..."
    
    # 执行改密
    ROTATE_OUTPUT=$(ops passwd auto-rotate --key "$KEY" --days "$DAYS" 2>&1)
    log "$ROTATE_OUTPUT"
    
    # 检查改密结果
    if echo "$ROTATE_OUTPUT" | grep -q "改密完成"; then
        log "改密执行成功"
        
        # 创建备份目录
        mkdir -p "$BACKUP_DIR"
        
        # 导出全量备份
        BACKUP_FILE="${BACKUP_DIR}/passwords-${TODAY}.kdbx"
        log "正在导出备份到: ${BACKUP_FILE}"
        
        if ops passwd export-keepass --key "$KEY" --output "$BACKUP_FILE" 2>&1 | tee -a "$LOG_FILE"; then
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
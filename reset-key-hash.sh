#!/bin/bash
# reset-key-hash.sh - 重置主密钥哈希

# 删除旧的key哈希
sqlite3 passwords.db "DELETE FROM key_hash;" 2>/dev/null || echo "使用BoltDB，需要手动删除"

# 或者直接删除数据库重新导入
# rm passwords.db
# ./ops passwd import ...

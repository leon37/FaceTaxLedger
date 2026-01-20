#!/bin/bash

# 遇到错误立即停止
set -e

echo "Starting Swagger generation..."

# 切换到脚本所在目录的上一级（即项目根目录）
cd "$(dirname "$0")/.."

# 检查是否安装了 swag
if ! command -v swag &> /dev/null; then
    echo "Swag tool not found. Installing..."
    go install github.com/swaggo/swag/cmd/swag@latest
fi

# 执行生成命令
# -g 指定入口文件
# -o 指定输出目录
# --parseInternal 解析 internal 包中的结构体定义
swag init -g cmd/server/main.go -o docs --parseInternal

echo "Swagger generation complete."
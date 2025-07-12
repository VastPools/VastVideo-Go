#!/bin/bash

# VastVideo-Go 启动脚本
# 先编译，再启动

set -e  # 遇到错误立即退出

echo "🚀 VastVideo-Go 启动脚本开始执行..."
echo "=================================="

# 检查 Go 环境
if ! command -v go &> /dev/null; then
    echo "❌ 错误: 未找到 Go 环境，请先安装 Go"
    exit 1
fi

echo "✅ Go 环境检查通过: $(go version)"

# 清理旧的编译文件
echo "🧹 清理旧的编译文件..."
rm -f vastvideo-go
rm -f *.exe

# 编译项目
echo "🔨 开始编译项目..."
if go build -o vastvideo-go .; then
    echo "✅ 编译成功!"
else
    echo "❌ 编译失败!"
    exit 1
fi

# 检查编译后的文件
if [ ! -f "vastvideo-go" ]; then
    echo "❌ 错误: 编译后的文件不存在"
    exit 1
fi

echo "📦 编译完成，文件大小: $(du -h vastvideo-go | cut -f1)"

# 设置执行权限
chmod +x vastvideo-go

echo "=================================="
echo "🎯 开始启动 VastVideo-Go 服务..."
echo "=================================="

# 启动服务
./vastvideo-go "$@"

echo "=================================="
echo "👋 VastVideo-Go 服务已退出"
echo "==================================" 
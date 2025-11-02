#!/bin/bash

# Gravital Core Web 开发环境启动脚本

echo "🚀 启动 Gravital Core Web 开发环境..."

# 检查 Node.js
if ! command -v node &> /dev/null; then
    echo "❌ Node.js 未安装，请先安装 Node.js 18+"
    exit 1
fi

# 检查 npm
if ! command -v npm &> /dev/null; then
    echo "❌ npm 未安装"
    exit 1
fi

# 检查依赖
if [ ! -d "node_modules" ]; then
    echo "📦 安装依赖..."
    npm install
fi

# 创建环境变量文件
if [ ! -f ".env.local" ]; then
    echo "📝 创建 .env.local 文件..."
    cp .env.example .env.local
fi

# 启动开发服务器
echo "✨ 启动开发服务器..."
npm run dev


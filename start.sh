#!/bin/bash

# 评茶员初赛理论题库系统 - 启动脚本

set -e

echo "=========================================="
echo "  评茶员初赛理论题库系统 - 启动脚本"
echo "=========================================="
echo ""

# 加载 .env 文件（如果存在）
if [ -f ".env" ]; then
    echo "加载 .env 配置文件..."
    export $(grep -v '^#' .env | xargs)
fi

# 检查 MySQL 是否运行
echo "[1/5] 检查 MySQL 服务..."
if ! mysqladmin ping -h localhost --silent 2>/dev/null; then
    echo "⚠️  警告: 无法连接到 MySQL，请确保 MySQL 服务已启动"
    echo "    macOS: brew services start mysql"
    echo "    Linux: sudo systemctl start mysql"
    exit 1
fi
echo "✓ MySQL 服务正常"

# 创建数据库（如果不存在）
echo ""
echo "[2/5] 初始化数据库..."
cd backend

# 获取环境变量或使用默认值
DB_USER=${DB_USER:-root}
DB_PASSWORD=${DB_PASSWORD:-}
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-3306}
DB_NAME=${DB_NAME:-tea_exam}

if [ -n "$DB_PASSWORD" ]; then
    mysql -u$DB_USER -p$DB_PASSWORD -e "CREATE DATABASE IF NOT EXISTS $DB_NAME CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;" 2>/dev/null || true
else
    mysql -u$DB_USER -e "CREATE DATABASE IF NOT EXISTS $DB_NAME CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;" 2>/dev/null || true
fi

echo "✓ 数据库初始化完成"

# 安装后端依赖
echo ""
echo "[3/5] 安装后端依赖..."
if [ ! -d "vendor" ]; then
    go mod tidy
    go mod download
fi
echo "✓ 后端依赖已安装"

# 编译后端
echo ""
echo "[4/5] 编译后端程序..."
go build -o tea-exam-server ./cmd/main.go
echo "✓ 后端编译完成"

# 启动后端（后台运行）
echo ""
echo "启动后端服务..."
export DB_USER=$DB_USER
export DB_PASSWORD=$DB_PASSWORD
export DB_HOST=$DB_HOST
export DB_PORT=$DB_PORT
export DB_NAME=$DB_NAME
export SERVER_PORT=8080

./tea-exam-server &
BACKEND_PID=$!
echo "✓ 后端服务已启动 (PID: $BACKEND_PID)"

cd ..

# 安装前端依赖
echo ""
echo "[5/5] 启动前端服务..."
cd frontend
if [ ! -d "node_modules" ]; then
    echo "安装前端依赖..."
    npm install
fi

echo "启动前端开发服务器..."
npm run dev &
FRONTEND_PID=$!
echo "✓ 前端服务已启动 (PID: $FRONTEND_PID)"

cd ..

echo ""
echo "=========================================="
echo "  系统启动成功！"
echo "=========================================="
echo ""
echo "访问地址："
echo "  答题页:    http://localhost:3000"
echo "  管理页:    http://localhost:3000/admin/login"
echo ""
echo "默认账号："
echo "  管理员:    admin / 123456"
echo "  答题用户:  张三 / 123456  或  李四 / 123456"
echo ""
echo "后端API:     http://localhost:8080/api"
echo ""
echo "按 Ctrl+C 停止服务"
echo ""

# 等待用户中断
trap "kill $BACKEND_PID $FRONTEND_PID 2>/dev/null; exit 0" INT
wait

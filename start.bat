@echo off
chcp 65001 >nul

echo ==========================================
echo   评茶员初赛理论题库系统 - 启动脚本
echo ==========================================
echo.

echo [1/5] 检查环境...
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo 错误: 未找到 Go，请安装 Go 1.21 或更高版本
    pause
    exit /b 1
)

where npm >nul 2>nul
if %errorlevel% neq 0 (
    echo 错误: 未找到 npm，请安装 Node.js 18 或更高版本
    pause
    exit /b 1
)

echo Go 和 Node.js 环境正常
echo.

echo [2/5] 初始化数据库...
echo 请确保 MySQL 服务已启动，然后按任意键继续...
pause >nul

cd backend

set DB_USER=root
set DB_PASSWORD=
set DB_HOST=localhost
set DB_PORT=3306
set DB_NAME=tea_exam

echo 创建数据库（如果不存在）...
mysql -u%DB_USER% -e "CREATE DATABASE IF NOT EXISTS %DB_NAME% CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;" 2>nul

echo 数据库初始化完成
echo.

echo [3/5] 安装后端依赖...
if not exist "go.sum" (
    go mod tidy
)
echo 后端依赖已安装
echo.

echo [4/5] 编译后端程序...
go build -o tea-exam-server.exe ./cmd/main.go
echo 后端编译完成
echo.

echo 启动后端服务...
start "Tea Exam Server" tea-exam-server.exe

cd ..
echo 后端服务已启动
echo.

echo [5/5] 启动前端服务...
cd frontend
if not exist "node_modules" (
    echo 安装前端依赖...
    call npm install
)

start "Tea Exam Frontend" npm run dev

cd ..
echo 前端服务已启动
echo.

echo ==========================================
echo   系统启动成功！
echo ==========================================
echo.
echo 访问地址：
echo   答题页:    http://localhost:3000
echo   管理页:    http://localhost:3000/admin/login
echo.
echo 默认账号：
echo   管理员:    admin / 123456
echo   答题用户:  张三 / 123456  或  李四 / 123456
echo.
echo 后端API:     http://localhost:8080/api
echo.
echo 关闭两个命令行窗口即可停止服务
echo.
pause

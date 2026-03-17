# 评茶员初赛理论题库系统

一个轻量级的本地部署题库系统，专为评茶员初赛理论考试设计。

## 技术栈

- **前端**: Vue 3 + Element Plus + Vite
- **后端**: Go + Gin + GORM
- **数据库**: MySQL

## 功能特性

### 答题页
- 姓名 + 密码登录
- 欢迎页展示考试统计
- 分页答题（每页10题）
- **支持单选题和多选题**
- 答题进度网格标记
- 已答题锁定（不可修改）
- 答题后立即显示对错
- 计时功能
- 中断续答
- 自动提交（完成所有题目后）
- 成绩结果页

### 题型支持
- **单选题**：点击选项直接提交，自动判定对错
- **多选题**：使用复选框选择多个选项，需点击"提交本题"按钮提交
  - 判题规则：所选答案与标准答案完全一致才算正确
  - 少选、多选、错选均判错，不支持部分得分
  - 多选题每题仍按 1 分计算

### 管理页
- 仅密码登录
- 题库导入（支持 .xlsx / .csv）
- 自动识别 UTF-8 / GBK 编码
- 覆盖/追加两种导入模式
- 历史考试记录查看
- 题库统计信息

## 快速开始

### 环境要求

- Go 1.21+
- Node.js 18+
- MySQL 5.7+

### 启动步骤

#### macOS / Linux

```bash
# 1. 确保 MySQL 已启动
brew services start mysql  # macOS
# 或
sudo systemctl start mysql  # Linux

# 2. 进入项目目录
cd TeaTest

# 3. 运行启动脚本
chmod +x start.sh
./start.sh
```

#### Windows

```cmd
# 1. 确保 MySQL 服务已启动
# 2. 双击运行 start.bat
```

### 访问地址

- **答题页**: http://localhost:3000
- **管理页**: http://localhost:3000/admin/login

### 默认账号

| 角色 | 账号 | 密码 |
|------|------|------|
| 管理员 | admin | 123456 |
| 答题用户 | 张三 | 123456 |
| 答题用户 | 李四 | 123456 |
| 答题用户 | 王五 | 123456 |

## 题库导入

### CSV 格式要求

CSV 文件应包含以下列：

| 列名 | 说明 | 必填 |
|------|------|------|
| 题号 | 题目序号 | 是 |
| 题库名称 | 题库分类名称 | 否 |
| 题型编码 | 题型代码（single_choice/multiple_choice） | 否 |
| 题型名称 | 题型描述 | 否 |
| 题干 | 题目内容 | 是 |
| 选项A | 选项A内容 | 是 |
| 选项B | 选项B内容 | 是 |
| 选项C | 选项C内容 | 否 |
| 选项D | 选项D内容 | 否 |
| 选项E | 选项E内容 | 否 |
| 正确答案 | 单选题填 A/B/C/D/E，多选题填组合如 AB、ACD、ABCDE | 是 |

### 编码支持

- **UTF-8**（推荐，带或不带 BOM 均可）
- **GBK / GB2312 / GB18030**（中文 Windows 默认编码）

系统会自动检测编码格式，无需手动转换。

### 导入模式

1. **覆盖模式**: 清空原有题库，导入新题目
2. **追加模式**: 保留原有题目，追加新题目

## 数据库配置

默认使用以下配置，可通过环境变量修改：

```bash
DB_HOST=localhost      # 数据库地址
DB_PORT=3306           # 数据库端口
DB_USER=root           # 数据库用户
DB_PASSWORD=           # 数据库密码
DB_NAME=tea_exam       # 数据库名
SERVER_PORT=8080       # 后端端口
```

## 项目结构

```
TeaTest/
├── backend/             # 后端代码
│   ├── cmd/            # 入口
│   ├── internal/       # 内部代码
│   │   ├── config/     # 配置
│   │   ├── handlers/   # HTTP 处理器
│   │   ├── middleware/ # 中间件
│   │   ├── models/     # 数据模型
│   │   ├── services/   # 业务逻辑
│   │   └── utils/      # 工具函数
│   ├── pkg/            # 公共包
│   │   ├── csvutil/    # CSV 工具
│   │   └── response/   # 响应封装
│   └── go.mod          # Go 依赖
├── frontend/           # 前端代码
│   ├── src/
│   │   ├── api/        # API 接口
│   │   ├── components/ # 组件
│   │   ├── router/     # 路由
│   │   ├── stores/     # Pinia 状态
│   │   ├── views/      # 页面
│   │   └── App.vue     # 根组件
│   ├── package.json    # Node 依赖
│   └── vite.config.js  # Vite 配置
├── start.sh            # 启动脚本(macOS/Linux)
├── start.bat           # 启动脚本(Windows)
└── README.md           # 本文件
```

## API 接口

### 认证相关
- `POST /api/auth/user/login` - 用户登录
- `POST /api/auth/admin/login` - 管理员登录
- `GET /api/auth/me` - 获取当前用户

### 考试相关
- `GET /api/exam/in-progress` - 获取进行中的考试
- `POST /api/exam/start` - 开始考试
- `GET /api/exam/questions` - 获取分页题目
- `GET /api/exam/all-status` - 获取所有题目状态
- `POST /api/exam/:exam_id/answer` - 提交答案
- `GET /api/exam/:exam_id/unanswered` - 获取未答题列表
- `GET /api/exam/:exam_id/result` - 获取考试结果
- `GET /api/exam/stats` - 获取考试统计

### 管理相关
- `GET /api/admin/stats` - 获取题库统计
- `GET /api/admin/records` - 获取考试记录
- `POST /api/admin/questions/import` - 导入题库

## 开发说明

### 单独启动后端

```bash
cd backend
go run ./cmd/main.go
```

### 单独启动前端

```bash
cd frontend
npm install
npm run dev
```

### 构建生产版本

```bash
# 前端构建
cd frontend
npm run build

# 后端构建
cd ../backend
go build -o tea-exam-server ./cmd/main.go
```

## 数据维护

### 添加答题用户

直接在 MySQL 中执行：

```sql
USE tea_exam;
INSERT INTO exam_users (name, password, status, created_at, updated_at)
VALUES ('新用户', '密码', 1, NOW(), NOW());
```

### 修改管理员密码

```sql
USE tea_exam;
UPDATE admin_config SET admin_password = '新密码', updated_at = NOW();
```

## 注意事项

1. 支持单选题和多选题，题型编码为 `multiple_choice` 时自动识别为多选题
2. 答题用户账号直接在数据库维护，管理页不提供增删改功能
3. 密码明文存储（按需求）
4. 支持多用户同时在线答题
5. 同一账号可在多个设备登录（不互斥）
6. 多选题判题规则：所选答案与标准答案完全一致才算正确，少选、多选、错选均判错

## License

MIT

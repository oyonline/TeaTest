# 阿里云 ECS 单机部署

本文对应当前方案：中国内地 ECS、暂时没有域名、前端/后端/MySQL 放在同一台机器，通过公网 IP 访问。

## 一、服务器与安全组

建议从 Ubuntu 24.04 LTS、2 核 4 GB、40 GB 系统盘起步。安全组只开放：

- TCP 22：来源限制为你自己的公网 IP，不要对所有地址开放。
- TCP 80：测试阶段尽量只允许已知使用者的公网 IP；确需公开测试时再允许 `0.0.0.0/0`。
- 不开放 8080 和 3306。后端与 MySQL 只在 Docker 内部通信。

在 ECS 上按照 [Docker 官方 Ubuntu 安装文档](https://docs.docker.com/engine/install/ubuntu/) 安装 Docker Engine 和 Compose 插件，然后确认：

```bash
docker --version
docker compose version
```

## 二、上传代码并创建生产配置

将仓库克隆到 ECS，例如：

```bash
git clone <你的仓库地址> TeaTest
cd TeaTest
cp .env.production.example .env.production
chmod 600 .env.production
```

生成三个彼此不同的随机值：

```bash
openssl rand -hex 32
openssl rand -hex 32
openssl rand -hex 32
```

编辑 `.env.production`：

- `MYSQL_ROOT_PASSWORD`：第一个随机值。
- `MYSQL_PASSWORD`：第二个随机值。
- `JWT_SECRET`：第三个随机值。
- `ADMIN_INITIAL_PASSWORD`：你自己保存的管理员强密码，至少 12 个字符。

`.env.production` 已被 Git 忽略，不要把它发到聊天或提交到仓库。

## 三、首次启动

先检查 Compose 展开后的配置，再构建并启动：

```bash
docker compose --env-file .env.production config --quiet
docker compose --env-file .env.production up -d --build
docker compose --env-file .env.production ps
```

在 ECS 内检查：

```bash
curl --fail http://127.0.0.1/api/health
```

返回 `{"status":"ok"}` 后，在浏览器访问：

- 答题页：`http://ECS公网IP/`
- 管理页：`http://ECS公网IP/admin/login`

全新数据库不会创建示例答题人。用 `.env.production` 中的管理员初始密码登录管理后台，再从“答题人管理”新增账号。

## 四、迁移本机现有数据（需要时执行）

如果要把本机当前题库、答题人和考试记录一起迁到 ECS，先在本机导出：

```bash
mysqldump -uroot -p --single-transaction --routines --triggers tea_exam > ~/tea_exam-import.sql
scp ~/tea_exam-import.sql <ECS用户名>@<ECS公网IP>:~/tea_exam-import.sql
rm ~/tea_exam-import.sql
```

然后在 ECS 的仓库目录导入：

```bash
docker compose --env-file .env.production exec -T mysql \
  sh -c 'exec mysql -uroot -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE"' < ~/tea_exam-import.sql
docker compose --env-file .env.production restart backend
rm ~/tea_exam-import.sql
```

导入旧数据库后，`ADMIN_INITIAL_PASSWORD` 不会覆盖原管理员密码。旧版明文管理员和答题人密码仍能登录，并会在各自第一次成功登录后自动升级为 bcrypt 哈希。

## 五、日常运维

查看状态和日志：

```bash
docker compose --env-file .env.production ps
docker compose --env-file .env.production logs --tail=200 -f web backend mysql
```

更新版本前先备份，再拉取并重建：

```bash
mkdir -p ~/tea-exam-backups
docker compose --env-file .env.production exec -T mysql \
  sh -c 'exec mysqldump -uroot -p"$MYSQL_ROOT_PASSWORD" --single-transaction "$MYSQL_DATABASE"' \
  > "$HOME/tea-exam-backups/tea_exam-$(date +%F-%H%M%S).sql"
git pull --ff-only
docker compose --env-file .env.production up -d --build
curl --fail http://127.0.0.1/api/health
```

数据库存放在 Docker 卷 `tea-exam_mysql_data` 中。不要执行 `docker compose down -v`，其中的 `-v` 会删除数据库卷。

## 六、没有域名时的限制

当前只能使用 HTTP 公网 IP。HTTP 不会加密登录密码、JWT 和题库内容，所以它适合部署验证或受限范围试用，不适合直接长期公开使用。正式组织考试前，建议购买域名、完成中国内地 ICP 备案，并配置 HTTPS；届时只需在现有架构前补 TLS，不需要拆分服务器。

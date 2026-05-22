# 部署与运维指南

## 1. 部署形态

Stellaris 有两类节点：

- **Core**（调度中心）：单实例服务，依赖 postgres / redis / emqx。推荐 docker-compose 部署。
- **Planet**（代理）：每台需要接入 Stellaris 的机器都部署一个 `stellaris-cli` 守护进程。推荐 `install.sh` + systemd。

## 2. Core 部署（Docker Compose）

### 2.1 准备

```bash
git clone <repo> stellaris
cd stellaris
cp deploy/.env.example deploy/.env
# 编辑 deploy/.env，把 JWT secrets 改成 32 字节随机串：
openssl rand -hex 32   # 每个 secret 生成一个
```

### 2.2 构建镜像

```bash
make docker-build   # 构建 stellaris-core / stellaris-cli / stellaris-web 三个镜像
```

> `docker-build-web` 会先在宿主机跑 `cd web && npm ci && npm run build`，把 `web/dist` COPY 进 nginx 镜像。容器内不再跑 npm install，避开网络慢的坑。

### 2.3 启动

```bash
cd deploy && docker compose -f docker-compose.prod.yml --env-file .env up -d
docker compose -f docker-compose.prod.yml ps
```

启动后：

- Web 面板：`http://<host>:8080`
- Core API：`http://<host>:4228`
- MQTT：`<host>:1883`（agent 接入需要外部可达）

### 2.4 日志与排查

```bash
docker compose -f docker-compose.prod.yml logs -f core
docker compose -f docker-compose.prod.yml logs -f web
```

postgres 数据卷在 `deploy_postgres_data` volume。备份：

```bash
docker exec deploy-postgres-1 pg_dump -U stellaris stellaris > backup-$(date +%F).sql
```

## 3. Planet 部署

### 3.1 一键脚本（推荐）

```bash
curl -fsSL https://stellaris.dev/install.sh | sudo bash
```

或本地：

```bash
sudo bash deploy/install.sh
```

环境变量：
- `STELLARIS_VERSION` 指定版本（默认 latest）
- `STELLARIS_PREFIX` 安装前缀（默认 `/usr/local`）
- `STELLARIS_MIRROR` GitHub release 镜像 URL 前缀

### 3.2 加入星系

```bash
stellaris-cli orbit <core-ip>:4228 <gid> --token <node-token>
```

`<node-token>` 来自面板"星系管理"页创建星系时返回（仅一次显示）。

### 3.3 声明本机 Agent

```bash
stellaris-cli agent add claw-1 --type openclaw --binary /usr/local/bin/openclaw --args "chat --stream"
stellaris-cli agent list
```

### 3.4 启动守护进程

```bash
sudo systemctl enable --now stellaris-cli
sudo systemctl status stellaris-cli
journalctl -u stellaris-cli -f
```

## 4. 升级

### 4.1 Core 升级

```bash
git pull
make docker-build
cd deploy && docker compose -f docker-compose.prod.yml --env-file .env up -d
```

数据库 migration 自动应用到首次启动的新 volume；存量库需要手动跑增量：

```bash
docker exec -i deploy-postgres-1 psql -U stellaris -d stellaris < server/migrations/003_xxx.up.sql
```

### 4.2 Planet 升级

```bash
sudo bash deploy/install.sh   # 重跑同一脚本即可覆盖
sudo systemctl restart stellaris-cli
```

## 5. 卸载

### 5.1 Core

```bash
cd deploy && docker compose -f docker-compose.prod.yml down -v   # -v 会删数据卷
```

### 5.2 Planet

```bash
sudo bash deploy/uninstall.sh
sudo rm -rf /etc/stellaris ~/.stellaris   # 手动清理配置
```

## 6. 监控与可观测性

M4 范围内未集成 prometheus / grafana。M5+ 计划：

- prom metrics（task latency / chunk throughput / ws conn count）
- grafana dashboard
- 告警（task failed / planet offline）

当前手段：

- `docker logs` 看运行时日志
- `journalctl -u stellaris-cli` 看 agent 端日志
- `SELECT * FROM tasks WHERE status='failed' ORDER BY created_at DESC LIMIT 50;` 看失败任务

## 7. 安全建议

- **JWT secrets**：生产必须用 `openssl rand -hex 32` 生成的强随机串，不要用 `.env.example` 的默认值
- **MQTT**：当前 EMQX 配置 `allow_anonymous=true`。生产应启 auth + ACL，限制 `stellaris/+/task/+` topic 只有 Core 能发
- **nginx**：M4 默认 HTTP。生产前置 Caddy / Cloudflare 或自己加 SSL 证书
- **防火墙**：4228 与 1883 端口应只对受信网段开放；8080 看部署形态决定
- **WS ticket**：M3 引入了一次性 ticket（5 分钟 TTL），生产无需额外配置；多实例部署需要把 `internal/ticket` 换成 Redis 后端

# Stellaris

[English](README.en.md) | **中文**

**跨平台分布式异构 Agent 统一调度平台。将分散在不同机器上的 AI Agent 纳入统一调度，支持并行、串行、DAG 编排，通过 Web 面板或 API 管理会话与输出流。**

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![CI](https://github.com/Veni222987/stellaris/actions/workflows/ci.yml/badge.svg)](https://github.com/Veni222987/stellaris/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)

## 快速启动

### 开发环境

```bash
make up      # 启动 postgres / redis / emqx
make build   # 编译 bin/stellaris-core + bin/stellaris-cli

bin/stellaris-core -f server/etc/core.yaml
```

前端面板：

```bash
cd web && npm install && npm run dev
# http://localhost:5173
```

### 生产部署（Docker）

```bash
cp deploy/.env.example deploy/.env   # 填入 JWT secrets（openssl rand -hex 32）
make docker-build
cd deploy && docker compose -f docker-compose.prod.yml --env-file .env up -d
# Web 面板: http://localhost:8080  Core API: http://localhost:4228
```

### 安装 Agent（Planet）

```bash
curl -fsSL https://raw.githubusercontent.com/Veni222987/stellaris/main/deploy/install.sh | sh
stellaris-cli orbit <core-host>:4228 <gid> --token <node-token>
stellaris-cli agent add my-agent --type openclaw --binary /usr/local/bin/openclaw
stellaris-cli start
```

## 模块

| 模块 | 路径 | 职责 |
|------|------|------|
| Core | `server/` | 调度中心：鉴权、星系管理、MQTT 任务下发、WebSocket 输出流 |
| Planet | `agent/` | 守护进程 `stellaris-cli`：注册 Agent、消费任务、回传 stdout |
| Web | `web/` | Vue 3 控制台：三栏聊天、实时输出、星系管理 |
| Protocol | `shared/protocol/` | MQTT topic 规范与消息结构 |
| Token | `shared/token/` | JWT 工具（UserJwt / PlanetJwt） |

## 文档

- [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) — 生产部署与运维指南

# Stellaris

**English** | [中文](README.md)

**A unified scheduling platform for heterogeneous AI Agents across distributed machines. Connect agents such as OpenClaw, HermesAgent, and Workbuddy on any host; dispatch tasks in parallel, relay, or DAG mode; stream results to a web console or API clients in real time.**

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![CI](https://github.com/Veni222987/stellaris/actions/workflows/ci.yml/badge.svg)](https://github.com/Veni222987/stellaris/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)

## Quick Start

### Development

```bash
make up      # start postgres / redis / emqx
make build   # compile bin/stellaris-core + bin/stellaris-cli

bin/stellaris-core -f server/etc/core.yaml
```

Web panel:

```bash
cd web && npm install && npm run dev
# http://localhost:5173
```

### Production (Docker)

```bash
cp deploy/.env.example deploy/.env   # set JWT secrets (openssl rand -hex 32)
make docker-build
cd deploy && docker compose -f docker-compose.prod.yml --env-file .env up -d
# Web panel: http://localhost:8080  Core API: http://localhost:4228
```

### Install an Agent (Planet)

```bash
curl -fsSL https://raw.githubusercontent.com/Veni222987/stellaris/main/deploy/install.sh | sh
stellaris-cli orbit <core-host>:4228 <gid> --token <node-token>
stellaris-cli agent add my-agent --type openclaw --binary /usr/local/bin/openclaw
stellaris-cli start
```

## Modules

| Module | Path | Role |
|--------|------|------|
| Core | `server/` | Scheduling brain: auth, galaxy management, MQTT task dispatch, WebSocket streaming |
| Planet | `agent/` | `stellaris-cli` daemon: register agents, consume tasks, stream stdout back to Core |
| Web | `web/` | Vue 3 console: three-pane chat, live output, galaxy management |
| Protocol | `shared/protocol/` | MQTT topic conventions and message schemas |
| Token | `shared/token/` | JWT utilities (UserJwt / PlanetJwt) |

## Docs

- [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) — Production deployment and operations guide

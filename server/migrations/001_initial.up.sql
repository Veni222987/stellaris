-- migrations/001_initial.up.sql

-- 用户
CREATE TABLE users (
    id           BIGSERIAL PRIMARY KEY,
    email        VARCHAR(128) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 星系
CREATE TABLE galaxies (
    id              BIGSERIAL PRIMARY KEY,
    gid             VARCHAR(16) UNIQUE NOT NULL,           -- 16位十六进制
    name            VARCHAR(128) NOT NULL,
    node_token_hash TEXT NOT NULL,                         -- bcrypt
    owner_user_id   BIGINT NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 行星
CREATE TABLE planets (
    id              BIGSERIAL PRIMARY KEY,
    planet_uuid     VARCHAR(64) UNIQUE NOT NULL,
    gid             VARCHAR(16) NOT NULL REFERENCES galaxies(gid),
    ip              VARCHAR(64) NOT NULL,
    hostname        VARCHAR(128) NOT NULL,
    os              VARCHAR(64),
    status          VARCHAR(16) NOT NULL DEFAULT 'offline',
    last_heartbeat  TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_planets_gid ON planets(gid);

-- Agent
CREATE TABLE agents (
    id                BIGSERIAL PRIMARY KEY,
    agent_uuid        VARCHAR(64) UNIQUE NOT NULL,
    planet_id         BIGINT NOT NULL REFERENCES planets(id),
    type              VARCHAR(32) NOT NULL,
    name              VARCHAR(128) NOT NULL,
    models_json       JSONB NOT NULL DEFAULT '[]',
    capabilities_json JSONB NOT NULL DEFAULT '{}',
    status            VARCHAR(16) NOT NULL DEFAULT 'unknown',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_agents_planet ON agents(planet_id);

-- 会话
CREATE TABLE sessions (
    id          BIGSERIAL PRIMARY KEY,
    session_uuid VARCHAR(64) UNIQUE NOT NULL,
    gid         VARCHAR(16) NOT NULL REFERENCES galaxies(gid),
    user_id     BIGINT NOT NULL REFERENCES users(id),
    mode        VARCHAR(16) NOT NULL DEFAULT 'single',
    agent_uuids JSONB NOT NULL DEFAULT '[]',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 消息
CREATE TABLE messages (
    id          BIGSERIAL PRIMARY KEY,
    session_id  BIGINT NOT NULL REFERENCES sessions(id),
    role        VARCHAR(16) NOT NULL,
    content     TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_messages_session ON messages(session_id);

-- 任务
CREATE TABLE tasks (
    id           BIGSERIAL PRIMARY KEY,
    task_uuid    VARCHAR(64) UNIQUE NOT NULL,
    message_id   BIGINT NOT NULL REFERENCES messages(id),
    agent_id     BIGINT NOT NULL REFERENCES agents(id),
    planet_id    BIGINT NOT NULL REFERENCES planets(id),
    status       VARCHAR(16) NOT NULL DEFAULT 'pending',
    started_at   TIMESTAMPTZ,
    finished_at  TIMESTAMPTZ,
    error        TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_tasks_message ON tasks(message_id);

-- 任务输出分片
CREATE TABLE task_chunks (
    id        BIGSERIAL PRIMARY KEY,
    task_id   BIGINT NOT NULL REFERENCES tasks(id),
    seq       INTEGER NOT NULL,
    chunk     TEXT NOT NULL,
    chunk_type VARCHAR(16) NOT NULL DEFAULT 'stdout',
    ts        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (task_id, seq)
);

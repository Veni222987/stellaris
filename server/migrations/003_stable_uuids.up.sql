-- 保证同一星系内同 hostname 只有一个 planet 行，UUID 跨 orbit 重启保持稳定
ALTER TABLE planets ADD CONSTRAINT uq_planets_gid_hostname UNIQUE (gid, hostname);

-- 保证同一 planet 内同 name 只有一个 agent 行，UUID 跨 start 重启保持稳定
ALTER TABLE agents ADD CONSTRAINT uq_agents_planet_name UNIQUE (planet_id, name);

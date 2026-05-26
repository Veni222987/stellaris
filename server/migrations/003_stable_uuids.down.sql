ALTER TABLE agents DROP CONSTRAINT IF EXISTS uq_agents_planet_name;
ALTER TABLE planets DROP CONSTRAINT IF EXISTS uq_planets_gid_hostname;

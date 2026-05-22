package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoad(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("STELLARIS_CONFIG_DIR", dir)

	in := &Config{
		CoreAddr:   "http://10.0.0.1:4228",
		GID:        "ab12cd34ef567890",
		PlanetUUID: "puuid",
		PlanetJwt:  "jwt-content",
		MQTTBroker: "tcp://10.0.0.1:1883",
		AgentsDir:  "/etc/stellaris/agents.d",
	}
	if err := Save(in); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "config.yaml")); err != nil {
		t.Fatalf("config.yaml not written: %v", err)
	}

	out, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if out.GID != in.GID || out.PlanetJwt != in.PlanetJwt || out.AgentsDir != in.AgentsDir {
		t.Fatalf("mismatch: %+v vs %+v", out, in)
	}
}

func TestLoad_NotJoined(t *testing.T) {
	t.Setenv("STELLARIS_CONFIG_DIR", t.TempDir())
	if _, err := Load(); err == nil {
		t.Fatal("expected error when config absent")
	}
}

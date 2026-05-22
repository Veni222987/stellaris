package scheduler

import "testing"

func TestParseDSL_HappyPath(t *testing.T) {
	raw := []byte(`{"nodes":[
		{"id":"a","agent_uuid":"u1","prompt":"{{user}}"},
		{"id":"b","agent_uuid":"u2","depends_on":["a"],"prompt":"polish: {{node:a}}"}
	]}`)
	d, err := ParseDSL(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Nodes) != 2 {
		t.Fatalf("got %d nodes", len(d.Nodes))
	}
}

func TestParseDSL_DetectsCycle(t *testing.T) {
	raw := []byte(`{"nodes":[
		{"id":"a","agent_uuid":"u1","depends_on":["b"]},
		{"id":"b","agent_uuid":"u2","depends_on":["a"]}
	]}`)
	if _, err := ParseDSL(raw); err == nil {
		t.Fatal("expected cycle error")
	}
}

func TestParseDSL_RejectsDanglingDep(t *testing.T) {
	raw := []byte(`{"nodes":[{"id":"a","agent_uuid":"u1","depends_on":["ghost"]}]}`)
	if _, err := ParseDSL(raw); err == nil {
		t.Fatal("expected dangling dep error")
	}
}

package adapter

import (
	"context"
	"strings"
	"testing"
)

// TestHermesAdapter_Defaults 验证 HermesAdapter 默认类型与能力标志。
func TestHermesAdapter_Defaults(t *testing.T) {
	a := NewHermesAdapter(HermesConfig{Name: "h1"})
	if a.Type() != "hermes" {
		t.Fatalf("type=%s", a.Type())
	}
	if !a.Capabilities().Streaming {
		t.Fatal("expected streaming")
	}
}

// TestWorkbuddyAdapter_Defaults 验证 WorkbuddyAdapter 默认类型与能力标志。
// Workbuddy 与 Hermes 的关键区别：ModelSwitch=false（不支持动态切模型）。
func TestWorkbuddyAdapter_Defaults(t *testing.T) {
	a := NewWorkbuddyAdapter(WorkbuddyConfig{Name: "w1"})
	if a.Type() != "workbuddy" {
		t.Fatalf("type=%s", a.Type())
	}
	if !a.Capabilities().Streaming {
		t.Fatal("expected streaming=true")
	}
	if a.Capabilities().ModelSwitch {
		t.Fatal("expected modelSwitch=false（workbuddy 不支持动态切模型）")
	}
}

// TestHermesAdapter_ExecMock 用内联 shell 脚本模拟 hermes CLI，验证流式输出。
func TestHermesAdapter_ExecMock(t *testing.T) {
	a := NewHermesAdapter(HermesConfig{
		Name:   "h-mock",
		Binary: "/bin/sh",
		Args:   []string{"-c", "read -r p; echo \"Hermes echo: $p\""},
	})
	ch, err := a.Chat(context.Background(), ChatRequest{Prompt: "ping"})
	if err != nil {
		t.Fatal(err)
	}
	var got strings.Builder
	for c := range ch {
		got.WriteString(c.Chunk)
	}
	if !strings.Contains(got.String(), "Hermes echo: ping") {
		t.Fatalf("got %q", got.String())
	}
}

// TestWorkbuddyAdapter_ExecMock 用内联 shell 脚本模拟 workbuddy CLI，验证流式输出。
func TestWorkbuddyAdapter_ExecMock(t *testing.T) {
	a := NewWorkbuddyAdapter(WorkbuddyConfig{
		Name:   "w-mock",
		Binary: "/bin/sh",
		Args:   []string{"-c", "read -r p; echo \"Workbuddy got: $p\""},
	})
	ch, err := a.Chat(context.Background(), ChatRequest{Prompt: "ping"})
	if err != nil {
		t.Fatal(err)
	}
	var got strings.Builder
	for c := range ch {
		got.WriteString(c.Chunk)
	}
	if !strings.Contains(got.String(), "Workbuddy got: ping") {
		t.Fatalf("got %q", got.String())
	}
}

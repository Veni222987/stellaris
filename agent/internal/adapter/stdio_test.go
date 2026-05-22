package adapter

import (
	"context"
	"strings"
	"testing"
)

func TestStdioAdapter_EchoBack(t *testing.T) {
	a := NewStdioAdapter(StdioConfig{
		Name:   "echo-1",
		Type:   "echo",
		Binary: "/bin/sh",
		Args:   []string{"-c", "while IFS= read -r line; do echo \"reply: $line\"; done"},
	})
	chunks, err := a.Chat(context.Background(), ChatRequest{Prompt: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	var buf strings.Builder
	sawDone := false
	for c := range chunks {
		if c.Type == "done" {
			sawDone = true
		}
		buf.WriteString(c.Chunk)
	}
	if !strings.Contains(buf.String(), "reply: hello") {
		t.Fatalf("missing reply line; got %q", buf.String())
	}
	if !sawDone {
		t.Fatal("missing done chunk")
	}
}

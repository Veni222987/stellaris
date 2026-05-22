package daemon

import (
	"errors"
	"os"
	"strconv"
	"testing"
)

func TestWriteReadRemovePID(t *testing.T) {
	t.Setenv("STELLARIS_CONFIG_DIR", t.TempDir())

	if err := WritePID(); err != nil {
		t.Fatal(err)
	}
	pid, err := ReadPID()
	if err != nil {
		t.Fatal(err)
	}
	if pid != os.Getpid() {
		t.Fatalf("pid=%d expected %d", pid, os.Getpid())
	}
	if err := RemovePID(); err != nil {
		t.Fatal(err)
	}
}

func TestWritePID_DetectsAlive(t *testing.T) {
	t.Setenv("STELLARIS_CONFIG_DIR", t.TempDir())
	if err := WritePID(); err != nil {
		t.Fatal(err)
	}
	defer RemovePID()

	err := WritePID()
	if !errors.Is(err, ErrAlreadyRunning) {
		t.Fatalf("expected ErrAlreadyRunning, got %v", err)
	}
}

// 陈旧 pid 文件（pid 已死）应被自动覆盖。
func TestWritePID_OverwritesStale(t *testing.T) {
	t.Setenv("STELLARIS_CONFIG_DIR", t.TempDir())
	// pid=1 在容器/macOS 上一般都是 init，活着；这里挑一个不可能存在的高位 pid。
	if err := os.WriteFile(PIDPath(), []byte("2147483646"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := WritePID(); err != nil {
		t.Fatalf("expected stale pid to be overwritten, got %v", err)
	}
	defer RemovePID()
	got, err := ReadPID()
	if err != nil {
		t.Fatal(err)
	}
	if got != os.Getpid() {
		t.Fatalf("pid=%d expected %d", got, os.Getpid())
	}
}

// pid 文件内容带 trailing newline 应能正常解析。
func TestReadPID_TolerantOfWhitespace(t *testing.T) {
	t.Setenv("STELLARIS_CONFIG_DIR", t.TempDir())
	want := os.Getpid()
	if err := os.MkdirAll(t.TempDir(), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(PIDPath(), []byte(strconv.Itoa(want)+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := ReadPID()
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("pid=%d expected %d", got, want)
	}
}

// pid 文件损坏（非数字）应返回 error，且 WritePID 能恢复（覆盖）。
func TestWritePID_RecoversFromCorruptFile(t *testing.T) {
	t.Setenv("STELLARIS_CONFIG_DIR", t.TempDir())
	if err := os.WriteFile(PIDPath(), []byte("abc\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadPID(); err == nil {
		t.Fatal("expected ReadPID to error on corrupt content")
	}
	if err := WritePID(); err != nil {
		t.Fatalf("WritePID should recover from corrupt pid file, got %v", err)
	}
	defer RemovePID()
}

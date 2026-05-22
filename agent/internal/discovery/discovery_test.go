package discovery

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScan(t *testing.T) {
	dir := t.TempDir()
	yaml := `name: claw-1
type: openclaw
binary: /usr/local/bin/openclaw
args: ["chat", "--stream"]
`
	if err := os.WriteFile(filepath.Join(dir, "claw.yaml"), []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}
	// 非 yaml 文件应该被忽略
	_ = os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("ignore"), 0o644)

	decls, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(decls) != 1 {
		t.Fatalf("expected 1 decl, got %d", len(decls))
	}
	if decls[0].Name != "claw-1" || decls[0].Type != "openclaw" {
		t.Fatalf("decl mismatch: %+v", decls[0])
	}
}

func TestScan_MissingDir(t *testing.T) {
	decls, err := Scan(filepath.Join(t.TempDir(), "does-not-exist"))
	if err != nil {
		t.Fatal(err)
	}
	if decls != nil {
		t.Fatalf("expected nil decls, got %v", decls)
	}
}

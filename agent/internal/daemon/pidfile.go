// pid 文件管理：用于 stop/status 子命令探测守护进程状态。
package daemon

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// ErrAlreadyRunning 已存在活跃的 daemon 实例。
var ErrAlreadyRunning = errors.New("daemon already running")

// PIDPath 返回 daemon.pid 路径，跟 config 目录保持一致。
func PIDPath() string {
	if d := os.Getenv("STELLARIS_CONFIG_DIR"); d != "" {
		return filepath.Join(d, "daemon.pid")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".stellaris", "daemon.pid")
}

// WritePID 用 O_EXCL 原子创建 pid 文件；若已存在则按 stale 与否决定覆盖或报错。
// 这样两个并发 start 中只有一个能拿到锁，避免 ReadPID→processAlive→WriteFile 之间的 TOCTOU。
func WritePID() error {
	if err := os.MkdirAll(filepath.Dir(PIDPath()), 0o700); err != nil {
		return err
	}
	for range 2 {
		f, err := os.OpenFile(PIDPath(), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err == nil {
			_, werr := f.WriteString(strconv.Itoa(os.Getpid()))
			cerr := f.Close()
			if werr != nil {
				return werr
			}
			return cerr
		}
		if !errors.Is(err, os.ErrExist) {
			return err
		}
		// 文件已存在：活的就拒，stale 就清掉再试一次。
		pid, rerr := ReadPID()
		if rerr == nil && pid > 0 && processAlive(pid) {
			return fmt.Errorf("%w: pid=%d", ErrAlreadyRunning, pid)
		}
		if rerr := os.Remove(PIDPath()); rerr != nil && !errors.Is(rerr, os.ErrNotExist) {
			return rerr
		}
	}
	return fmt.Errorf("WritePID: pid 文件被并发竞争，请重试")
}

// RemovePID 删除 pid 文件。
func RemovePID() error {
	return os.Remove(PIDPath())
}

// ReadPID 读 pid 文件并解析成整型；容忍 trailing newline 和空白。
func ReadPID() (int, error) {
	data, err := os.ReadFile(PIDPath())
	if err != nil {
		return 0, err
	}
	s := strings.TrimSpace(string(data))
	if s == "" {
		return 0, fmt.Errorf("pid 文件为空")
	}
	pid, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return pid, nil
}

// Running 读 pid 文件并探测进程是否存活，返回 (pid, 是否在跑)。
func Running() (int, bool) {
	pid, err := ReadPID()
	if err != nil {
		return 0, false
	}
	return pid, processAlive(pid)
}

// processAlive 用 Signal(0) 探测进程是否存活（Unix 语义）。
func processAlive(pid int) bool {
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return p.Signal(syscall.Signal(0)) == nil
}

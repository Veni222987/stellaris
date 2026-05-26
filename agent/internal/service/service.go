// Package service 负责把 stellaris-cli 守护进程托管到操作系统服务管理器，
// 实现开机自启与被杀恢复：Linux 用 systemd（root 走 system unit，否则 user unit），
// macOS 用 launchd 用户级 LaunchAgent。无服务管理器或安装失败时，退化为
// SpawnBackground —— 脱离终端的后台监控进程（崩溃自恢复靠进程内 Supervise）。
package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/stellaris/stellaris/agent/internal/config"
	"github.com/stellaris/stellaris/agent/internal/daemon"
)

type Kind string

const (
	KindSystemdSystem Kind = "systemd (system)"
	KindSystemdUser   Kind = "systemd (user)"
	KindLaunchd       Kind = "launchd"
	KindNone          Kind = "none"
)

const (
	unitName  = "stellaris-cli.service"
	plistName = "dev.stellaris.cli"
)

func selfExe() (string, error) { return os.Executable() }

// Detect 选择当前平台/权限下最合适的服务托管方式。
func Detect() Kind {
	switch runtime.GOOS {
	case "linux":
		if _, err := exec.LookPath("systemctl"); err != nil {
			return KindNone
		}
		if os.Geteuid() == 0 {
			return KindSystemdSystem
		}
		return KindSystemdUser
	case "darwin":
		if _, err := exec.LookPath("launchctl"); err != nil {
			return KindNone
		}
		return KindLaunchd
	default:
		return KindNone
	}
}

// InstalledKind 返回已落盘的服务单元类型；都不存在时返回 KindNone。
func InstalledKind() Kind {
	if p, err := systemUnitPath(); err == nil {
		if _, err := os.Stat(p); err == nil {
			return KindSystemdSystem
		}
	}
	if p, err := userUnitPath(); err == nil {
		if _, err := os.Stat(p); err == nil {
			return KindSystemdUser
		}
	}
	if p, err := plistPath(); err == nil {
		if _, err := os.Stat(p); err == nil {
			return KindLaunchd
		}
	}
	return KindNone
}

// Install 写入服务单元并启用+启动；返回实际采用的类型。
func Install() (Kind, error) {
	switch Detect() {
	case KindSystemdSystem:
		return KindSystemdSystem, installSystemd(false)
	case KindSystemdUser:
		return KindSystemdUser, installSystemd(true)
	case KindLaunchd:
		return KindLaunchd, installLaunchd()
	default:
		return KindNone, fmt.Errorf("当前平台无可用服务管理器")
	}
}

// Stop 通过已安装的服务停止守护进程；未托管时返回 false。
func Stop() (bool, error) {
	switch InstalledKind() {
	case KindSystemdSystem:
		return true, run("systemctl", "stop", unitName)
	case KindSystemdUser:
		return true, run("systemctl", "--user", "stop", unitName)
	case KindLaunchd:
		p, err := plistPath()
		if err != nil {
			return true, err
		}
		return true, run("launchctl", "unload", p)
	default:
		return false, nil
	}
}

// SpawnBackground 兜底：脱离终端启动 `start --foreground` 后台进程，返回子进程 pid。
// 子进程 stdout 接 /dev/null，stderr 接日志文件（捕获 panic 栈）；进程内 Supervise
// 再把结构化日志写入同一文件。
func SpawnBackground() (int, error) {
	exe, err := selfExe()
	if err != nil {
		return 0, err
	}
	logf, err := daemon.OpenLog()
	if err != nil {
		return 0, err
	}
	defer logf.Close()

	cmd := exec.Command(exe, "start", "--foreground")
	cmd.Env = os.Environ()
	cmd.Stdout = nil // os/exec 把 nil 连到 /dev/null
	cmd.Stderr = logf
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return 0, err
	}
	pid := cmd.Process.Pid
	_ = cmd.Process.Release()
	return pid, nil
}

// ---- systemd ----

func systemUnitPath() (string, error) {
	return "/etc/systemd/system/" + unitName, nil
}

func userUnitPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "systemd", "user", unitName), nil
}

func installSystemd(user bool) error {
	exe, err := selfExe()
	if err != nil {
		return err
	}
	unit := systemdUnit(exe, user)

	var unitPath string
	if user {
		unitPath, err = userUnitPath()
	} else {
		unitPath, err = systemUnitPath()
	}
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(unitPath), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(unitPath, []byte(unit), 0o644); err != nil {
		return err
	}

	if user {
		if err := run("systemctl", "--user", "daemon-reload"); err != nil {
			return err
		}
		if err := run("systemctl", "--user", "enable", "--now", unitName); err != nil {
			return err
		}
		// 让 user 服务在未登录时也能开机自启（best-effort，失败不致命）。
		if u := os.Getenv("USER"); u != "" {
			_ = run("loginctl", "enable-linger", u)
		}
		return nil
	}
	if err := run("systemctl", "daemon-reload"); err != nil {
		return err
	}
	return run("systemctl", "enable", "--now", unitName)
}

func systemdUnit(exe string, user bool) string {
	wantedBy := "multi-user.target"
	if user {
		wantedBy = "default.target"
	}
	var env strings.Builder
	for k, v := range serviceEnv() {
		fmt.Fprintf(&env, "Environment=%s=%s\n", k, v)
	}
	return fmt.Sprintf(`[Unit]
Description=Stellaris Planet 代理
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=%s start --foreground
ExecStop=%s stop
Restart=on-failure
RestartSec=5
%sStandardError=append:%s

[Install]
WantedBy=%s
`, exe, exe, env.String(), daemon.LogPath(), wantedBy)
}

// ---- launchd ----

func plistPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "LaunchAgents", plistName+".plist"), nil
}

func installLaunchd() error {
	exe, err := selfExe()
	if err != nil {
		return err
	}
	p, err := plistPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(p, []byte(launchdPlist(exe)), 0o644); err != nil {
		return err
	}
	// 先 unload 容忍「已加载」状态，再 load -w 持久启用。
	_ = run("launchctl", "unload", p)
	return run("launchctl", "load", "-w", p)
}

func launchdPlist(exe string) string {
	var env strings.Builder
	for k, v := range serviceEnv() {
		fmt.Fprintf(&env, "    <key>%s</key><string>%s</string>\n", k, v)
	}
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key><string>%s</string>
  <key>ProgramArguments</key>
  <array>
    <string>%s</string>
    <string>start</string>
    <string>--foreground</string>
  </array>
  <key>RunAtLoad</key><true/>
  <key>KeepAlive</key><true/>
  <key>EnvironmentVariables</key>
  <dict>
%s  </dict>
  <key>StandardOutPath</key><string>/dev/null</string>
  <key>StandardErrorPath</key><string>%s</string>
</dict>
</plist>
`, plistName, exe, env.String(), daemon.LogPath())
}

// ---- 通用 ----

// serviceEnv 收集要透传给服务的环境变量：始终带 STELLARIS_CONFIG_DIR（否则服务
// 在干净环境里找不到 config.yaml），并透传当前已设置的 AGENTS_DIR / LOG_FILE。
func serviceEnv() map[string]string {
	env := map[string]string{"STELLARIS_CONFIG_DIR": config.Dir()}
	for _, k := range []string{"STELLARIS_AGENTS_DIR", "STELLARIS_LOG_FILE"} {
		if v := os.Getenv(k); v != "" {
			env[k] = v
		}
	}
	return env
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s: %v: %s", name, strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return nil
}

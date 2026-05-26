package daemon

import (
	"context"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/stellaris/stellaris/agent/internal/config"
)

// LogPath 返回守护进程日志文件路径，与 config / pid 目录保持一致。
// 优先级：STELLARIS_LOG_FILE > <STELLARIS_CONFIG_DIR>/daemon.log > ~/.stellaris/daemon.log。
func LogPath() string {
	if p := os.Getenv("STELLARIS_LOG_FILE"); p != "" {
		return p
	}
	if d := os.Getenv("STELLARIS_CONFIG_DIR"); d != "" {
		return filepath.Join(d, "daemon.log")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".stellaris", "daemon.log")
}

const maxLogSize = 10 << 20 // 10MB

// OpenLog 打开日志文件（追加）。打开前若已超过 maxLogSize 则轮转为 .1（单备份，简单够用）。
// 返回的 *os.File 由调用方负责 Close。
func OpenLog() (*os.File, error) {
	p := LogPath()
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return nil, err
	}
	if fi, err := os.Stat(p); err == nil && fi.Size() > maxLogSize {
		_ = os.Rename(p, p+".1")
	}
	return os.OpenFile(p, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
}

// Supervise 是守护进程的常驻入口：把日志接到文件，循环跑 daemon 并在异常退出时退避重启，
// 直到 ctx 取消（收到 SIGTERM/SIGINT）才正常返回。
//
// 这样 orbit 时即便 Core/broker 还没起，daemon 也能反复重试直到连上；运行期单次崩溃
// 也会自动拉起，无需人工 start。OS 级硬崩溃/重启由上层 systemd/launchd 兜底。
func Supervise(ctx context.Context) error {
	f, err := OpenLog()
	if err != nil {
		return err
	}
	defer f.Close()
	// 日志同时进文件与 stdout：stdout 交给 systemd journald / launchd，文件用于本地排障。
	log.SetOutput(io.MultiWriter(os.Stdout, f))

	cfg, err := config.Load()
	if err != nil {
		// 配置缺失属致命错误（尚未 orbit），直接返回让上层报错退出。
		return err
	}

	const minBackoff = time.Second
	const maxBackoff = 30 * time.Second
	const healthyRun = time.Minute // 跑满这么久视为健康，重启计数清零
	backoff := minBackoff

	for {
		if ctx.Err() != nil {
			return nil
		}
		started := time.Now()
		runErr := runOnce(ctx, cfg)
		if ctx.Err() != nil { // ctx 取消导致的退出属正常
			return nil
		}
		if time.Since(started) >= healthyRun {
			backoff = minBackoff
		}
		log.Printf("[supervise] daemon 退出: %v（%s 后重启）", runErr, backoff)

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(backoff):
		}
		if backoff < maxBackoff {
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}
}

// runOnce 跑一轮 daemon，返回其退出原因（nil 表示 ctx 取消的正常退出）。
func runOnce(ctx context.Context, cfg *config.Config) error {
	d, err := New(cfg)
	if err != nil {
		return err
	}
	return d.Run(ctx)
}

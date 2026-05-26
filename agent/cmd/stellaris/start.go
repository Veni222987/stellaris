package stellaris

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/stellaris/stellaris/agent/internal/daemon"
	"github.com/stellaris/stellaris/agent/internal/service"
)

var startForeground bool

func init() {
	startCmd.Flags().BoolVar(&startForeground, "foreground", false,
		"前台运行 supervisor（供 systemd/launchd ExecStart 或调试用）")
	rootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "启动守护进程（默认后台：优先注册 OS 服务，失败则脱离终端后台运行）",
	RunE: func(cmd *cobra.Command, args []string) error {
		if startForeground {
			return runForeground()
		}
		return launch()
	},
}

// runForeground 是常驻 supervisor 入口：写 pid、收信号、跑 Supervise（崩溃自恢复 + 日志落盘）。
func runForeground() error {
	if err := daemon.WritePID(); err != nil {
		return err
	}
	defer daemon.RemovePID()
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	return daemon.Supervise(ctx)
}

// launch 是智能启动器：已在跑→报告；否则优先装并启用 OS 服务，失败则后台 spawn。
func launch() error {
	if pid, ok := daemon.Running(); ok {
		fmt.Printf("守护进程已在运行 (pid=%d)\n", pid)
		return nil
	}
	if service.Detect() != service.KindNone {
		kind, err := service.Install()
		if err == nil {
			fmt.Printf("✓ 已通过 %s 启用并启动守护进程\n  日志: %s\n", kind, daemon.LogPath())
			return nil
		}
		fmt.Printf("OS 服务安装失败（%v），改用脱离终端的后台进程\n", err)
	}
	pid, err := service.SpawnBackground()
	if err != nil {
		return err
	}
	fmt.Printf("✓ 守护进程已在后台启动 (pid=%d)\n  日志: %s\n", pid, daemon.LogPath())
	return nil
}

package stellaris

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/stellaris/stellaris/agent/internal/config"
	"github.com/stellaris/stellaris/agent/internal/daemon"
)

func init() { rootCmd.AddCommand(startCmd) }

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "前台运行 Stellaris Planet 守护进程（M2 加 daemonize / systemd unit）",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		// 写入 pid 文件用于 stop / status；进程退出时清理。
		if err := daemon.WritePID(); err != nil {
			return err
		}
		defer daemon.RemovePID()

		d, err := daemon.New(cfg)
		if err != nil {
			return err
		}
		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()
		fmt.Println("Stellaris daemon 已启动，Ctrl+C 退出。")
		return d.Run(ctx)
	},
}

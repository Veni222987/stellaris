package stellaris

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/stellaris/stellaris/agent/internal/daemon"
)

var logsFollow bool

func init() {
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "持续跟随新增日志（Ctrl+C 退出）")
	rootCmd.AddCommand(logsCmd)
}

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "查看守护进程日志文件",
	RunE: func(cmd *cobra.Command, args []string) error {
		p := daemon.LogPath()
		f, err := os.Open(p)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("暂无日志文件: %s", p)
			}
			return err
		}
		defer f.Close()

		if _, err := io.Copy(os.Stdout, f); err != nil {
			return err
		}
		if !logsFollow {
			return nil
		}
		// 跟随模式：轮询读取追加内容，直到 Ctrl+C。
		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(500 * time.Millisecond):
			}
			if _, err := io.Copy(os.Stdout, f); err != nil {
				return err
			}
		}
	},
}

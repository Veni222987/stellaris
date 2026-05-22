package stellaris

import (
	"errors"
	"fmt"
	"os"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/stellaris/stellaris/agent/internal/daemon"
)

func init() { rootCmd.AddCommand(stopCmd) }

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "向守护进程发 SIGTERM 让它优雅退出",
	RunE: func(cmd *cobra.Command, args []string) error {
		pid, err := daemon.ReadPID()
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("daemon 没在跑（pid 文件不存在）")
			}
			return fmt.Errorf("读取 pid 文件失败: %w", err)
		}
		// 进程已死时只清陈旧 pid 文件，不发信号——避免误杀同 pid 复用的无关进程。
		if syscall.Kill(pid, syscall.Signal(0)) != nil {
			_ = daemon.RemovePID()
			return fmt.Errorf("daemon 没在跑（清理了陈旧 pid=%d）", pid)
		}
		if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
			return fmt.Errorf("发送 SIGTERM 失败: %w", err)
		}
		fmt.Printf("✓ 已向 pid=%d 发送 SIGTERM\n", pid)
		return nil
	},
}

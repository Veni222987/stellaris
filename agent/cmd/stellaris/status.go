package stellaris

import (
	"fmt"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/stellaris/stellaris/agent/internal/config"
	"github.com/stellaris/stellaris/agent/internal/daemon"
)

func init() { rootCmd.AddCommand(statusCmd) }

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "查看守护进程与当前配置",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Println("配置: 未加入任何星系")
		} else {
			fmt.Printf("配置: gid=%s planet=%s core=%s\n", cfg.GID, cfg.PlanetUUID, cfg.CoreAddr)
		}
		pid, err := daemon.ReadPID()
		if err != nil {
			fmt.Println("守护进程: 未运行")
			return
		}
		if syscall.Kill(pid, 0) == nil {
			fmt.Printf("守护进程: 运行中 (pid=%d)\n", pid)
		} else {
			fmt.Printf("守护进程: pid 文件 %d 但进程不存在（陈旧 pid 文件）\n", pid)
		}
	},
}

package stellaris

import (
	"fmt"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/stellaris/stellaris/agent/internal/config"
	"github.com/stellaris/stellaris/agent/internal/daemon"
	"github.com/stellaris/stellaris/agent/internal/service"
)

func init() { rootCmd.AddCommand(statusCmd) }

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "查看守护进程、服务托管、日志与已发现的 agent",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Println("配置: 未加入任何星系")
		} else {
			fmt.Printf("配置: gid=%s planet=%s core=%s\n", cfg.GID, cfg.PlanetUUID, cfg.CoreAddr)
		}

		pid, err := daemon.ReadPID()
		switch {
		case err != nil:
			fmt.Println("守护进程: 未运行")
		case syscall.Kill(pid, 0) == nil:
			fmt.Printf("守护进程: 运行中 (pid=%d)\n", pid)
		default:
			fmt.Printf("守护进程: pid 文件 %d 但进程不存在（陈旧 pid 文件）\n", pid)
		}

		if k := service.InstalledKind(); k != service.KindNone {
			fmt.Printf("服务: 由 %s 托管\n", k)
		} else {
			fmt.Println("服务: 未托管（后台/手动）")
		}
		fmt.Printf("日志: %s\n", daemon.LogPath())

		resolved, err := daemon.ResolveAgents(config.AgentsDirPath())
		if err != nil {
			fmt.Printf("agent: 解析失败: %v\n", err)
			return
		}
		if len(resolved) == 0 {
			fmt.Println("agent: 未发现（已知类型不在 PATH，agents.d 也为空）")
			return
		}
		fmt.Println("agent:")
		for _, ra := range resolved {
			fmt.Printf("  - %s (type=%s) [%s]\n", ra.Decl.Name, ra.Decl.Type, ra.Source)
		}
	},
}

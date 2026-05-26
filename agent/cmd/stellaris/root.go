// Package stellaris 是 stellaris-cli 所有 Cobra 子命令的容器。
//
// 命令一览：
//   stellaris-cli orbit <ip>:<port> <gid> --token <node-token>   加入星系并自动拉起守护进程
//   stellaris-cli start [--foreground]                           启动守护进程（默认后台/OS 服务）
//   stellaris-cli stop / status                                  停止 / 查看状态
//   stellaris-cli logs [-f]                                      查看守护进程日志
//   stellaris-cli agent list / add / remove                      查看（自动发现）/ 覆盖 Agent
package stellaris

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "stellaris-cli",
	Short: "Stellaris Planet 代理 —— 把本机 Agent 接入 Stellaris 星系",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

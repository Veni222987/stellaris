// Package stellaris 是 stellaris-cli 所有 Cobra 子命令的容器。
//
// 命令一览：
//   stellaris-cli orbit <ip>:<port> <gid> --token <node-token>   加入星系
//   stellaris-cli start                                          前台运行守护进程
//   stellaris-cli stop / status                                  （M2 实现）
//   stellaris-cli agent list / discover / add                    （M2 实现）
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

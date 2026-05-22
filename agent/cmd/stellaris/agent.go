package stellaris

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stellaris/stellaris/agent/internal/config"
	"github.com/stellaris/stellaris/agent/internal/discovery"
	"gopkg.in/yaml.v3"
)

func init() {
	agentCmd.AddCommand(agentListCmd, agentAddCmd, agentRemoveCmd)
	rootCmd.AddCommand(agentCmd)
}

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "管理本机声明的 Agent 配置（位于 agents.d 目录）",
}

var agentListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出当前已声明的 Agent",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := config.AgentsDirPath()
		decls, err := discovery.Scan(dir)
		if err != nil {
			return err
		}
		if len(decls) == 0 {
			fmt.Printf("（%s 为空）\n", dir)
			return nil
		}
		fmt.Printf("# %s\n", dir)
		for _, d := range decls {
			fmt.Printf("- %s (type=%s binary=%s)\n", d.Name, d.Type, d.Binary)
		}
		return nil
	},
}

var (
	addType   string
	addBinary string
	addArgs   string
)

func init() {
	agentAddCmd.Flags().StringVar(&addType, "type", "", "agent 类型 (openclaw/hermes/workbuddy/...)")
	agentAddCmd.Flags().StringVar(&addBinary, "binary", "", "二进制路径（缺省时使用 type 同名）")
	agentAddCmd.Flags().StringVar(&addArgs, "args", "", "启动参数，用空格分隔")
}

var agentAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "声明一个新的 Agent（生成 agents.d/<name>.yaml）",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if addType == "" {
			return fmt.Errorf("--type 不能为空")
		}
		if _, err := config.EnsureAgentsDir(); err != nil {
			return err
		}
		path := config.AgentFilePath(name)
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("%s 已存在，先 remove 再 add", path)
		}
		// 用 yaml.Marshal 生成文件，避免手拼字符串遇到特殊字符的转义陷阱。
		// 注意：--args 仍按空白拆分，含空格的 arg 暂不支持；后续可改 StringArray flag。
		var parts []string
		if addArgs != "" {
			parts = strings.Fields(addArgs)
		}
		body, err := yaml.Marshal(discovery.AgentDecl{
			Name:   name,
			Type:   addType,
			Binary: addBinary,
			Args:   parts,
		})
		if err != nil {
			return err
		}
		if err := os.WriteFile(path, body, 0o644); err != nil {
			return err
		}
		fmt.Printf("✓ 已写入 %s\n", path)
		return nil
	},
}

var agentRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "删除某个 Agent 的声明文件",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := filepath.Join(config.AgentsDirPath(), args[0]+".yaml")
		if err := os.Remove(path); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("%s 未声明", args[0])
			}
			return err
		}
		fmt.Printf("✓ 已删除 %s\n", path)
		return nil
	},
}

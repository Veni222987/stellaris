package stellaris

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/stellaris/stellaris/agent/internal/config"
)

var orbitToken string

func init() {
	orbitCmd.Flags().StringVar(&orbitToken, "token", "", "node-token (必填，来自调度中心 galaxy/create 返回)")
	rootCmd.AddCommand(orbitCmd)
}

var orbitCmd = &cobra.Command{
	Use:   "orbit <ip>:<port> <gid>",
	Short: "进入指定星系的轨道（即加入星系）",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if orbitToken == "" {
			return fmt.Errorf("--token 不能为空")
		}
		addr, gid := args[0], args[1]
		coreURL := "http://" + addr

		body, _ := json.Marshal(map[string]string{
			"gid":        gid,
			"node_token": orbitToken,
			"hostname":   hostname(),
			"ip":         localIPv4(),
			"os":         runtime.GOOS + "/" + runtime.GOARCH,
		})

		resp, err := (&http.Client{Timeout: 10 * time.Second}).Post(
			coreURL+"/api/planet/join", "application/json", bytes.NewReader(body))
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		data, _ := io.ReadAll(resp.Body)
		if resp.StatusCode >= 300 {
			return fmt.Errorf("加入星系失败 (HTTP %d): %s", resp.StatusCode, data)
		}
		var r struct {
			PlanetUUID string `json:"planet_uuid"`
			PlanetJwt  string `json:"planet_jwt"`
		}
		if err := json.Unmarshal(data, &r); err != nil {
			return err
		}

		brokerHost := strings.SplitN(addr, ":", 2)[0]
		// AgentsDir 留空，由 daemon 在 cfg 缺省时落到 /etc/stellaris/agents.d。
		// 用户想换目录直接编辑 config.yaml 加 agents_dir: 即可。
		if err := config.Save(&config.Config{
			CoreAddr:   coreURL,
			GID:        gid,
			PlanetUUID: r.PlanetUUID,
			PlanetJwt:  r.PlanetJwt,
			MQTTBroker: "tcp://" + brokerHost + ":1883",
		}); err != nil {
			return err
		}
		fmt.Printf("✓ 已加入星系 %s，本机 planet_uuid=%s\n", gid, r.PlanetUUID)
		return nil
	},
}

func hostname() string {
	h, _ := os.Hostname()
	return h
}

func localIPv4() string {
	addrs, _ := net.InterfaceAddrs()
	for _, a := range addrs {
		if ipn, ok := a.(*net.IPNet); ok && !ipn.IP.IsLoopback() && ipn.IP.To4() != nil {
			return ipn.IP.String()
		}
	}
	return "127.0.0.1"
}

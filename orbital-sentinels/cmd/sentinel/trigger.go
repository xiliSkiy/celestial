package main

import (
	"context"
	"fmt"
	"time"

	"github.com/celestial/orbital-sentinels/internal/plugin"
	pingplugin "github.com/celestial/orbital-sentinels/plugins/ping"
	"github.com/spf13/cobra"
)

var triggerCmd = &cobra.Command{
	Use:   "trigger",
	Short: "手动触发采集任务",
	Long:  `手动触发一次采集任务，用于测试或立即执行巡检`,
}

var triggerPingCmd = &cobra.Command{
	Use:   "ping [host]",
	Short: "触发 Ping 采集",
	Long:  `手动触发一次 Ping 采集任务`,
	Args:  cobra.ExactArgs(1),
	Run:   runTriggerPing,
}

func init() {
	triggerCmd.AddCommand(triggerPingCmd)

	// Ping 命令参数
	triggerPingCmd.Flags().IntP("count", "n", 4, "Ping 次数")
	triggerPingCmd.Flags().DurationP("interval", "i", 1*time.Second, "Ping 间隔")
	triggerPingCmd.Flags().DurationP("timeout", "t", 5*time.Second, "超时时间")
}

func runTriggerPing(cmd *cobra.Command, args []string) {
	host := args[0]
	count, _ := cmd.Flags().GetInt("count")
	interval, _ := cmd.Flags().GetDuration("interval")
	timeout, _ := cmd.Flags().GetDuration("timeout")

	fmt.Printf("🚀 触发 Ping 采集: %s\n", host)
	fmt.Printf("   参数: count=%d, interval=%v, timeout=%v\n\n", count, interval, timeout)

	// 创建并注册 Ping 插件
	p := pingplugin.NewPlugin()
	if err := p.Init(nil); err != nil {
		fmt.Printf("❌ 初始化插件失败: %v\n", err)
		return
	}
	defer p.Close()

	// 创建采集任务
	task := &plugin.CollectionTask{
		TaskID:     fmt.Sprintf("manual-ping-%d", time.Now().Unix()),
		DeviceID:   host,
		PluginName: "ping",
		DeviceConfig: map[string]interface{}{
			"host":     host,
			"count":    count,
			"interval": interval.String(),
			"timeout":  timeout.String(),
		},
		Timeout: timeout + 5*time.Second,
	}

	// 执行采集
	ctx, cancel := context.WithTimeout(context.Background(), task.Timeout)
	defer cancel()

	fmt.Println("⏳ 执行采集中...")
	startTime := time.Now()

	metrics, err := p.Collect(ctx, task)
	duration := time.Since(startTime)

	if err != nil {
		fmt.Printf("\n❌ 采集失败: %v\n", err)
		return
	}

	// 显示结果
	fmt.Printf("\n✅ 采集成功! 耗时: %v\n\n", duration.Round(time.Millisecond))
	fmt.Println("📊 采集指标:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	for i, m := range metrics {
		fmt.Printf("\n指标 #%d:\n", i+1)
		fmt.Printf("  名称: %s\n", m.Name)
		fmt.Printf("  值:   %.2f %s\n", m.Value, getUnit(m.Name))
		fmt.Printf("  类型: %s\n", m.Type)
		fmt.Printf("  时间: %s\n", time.Unix(m.Timestamp, 0).Format("2006-01-02 15:04:05"))

		if len(m.Labels) > 0 {
			fmt.Println("  标签:")
			for k, v := range m.Labels {
				fmt.Printf("    %s: %s\n", k, v)
			}
		}
	}

	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("\n💡 提示: 数据已采集但未发送到数据库\n")
	fmt.Printf("   如需自动发送，请使用 'sentinel start' 启动服务\n\n")
}

func getUnit(metricName string) string {
	switch metricName {
	case "ping_rtt_ms":
		return "ms"
	case "ping_packet_loss_percent":
		return "%"
	case "ping_status":
		return ""
	default:
		return ""
	}
}

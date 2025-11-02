package main

import (
	"fmt"
	"os"

	"github.com/celestial/orbital-sentinels/internal/agent"
	"github.com/celestial/orbital-sentinels/internal/pkg/config"
	"github.com/celestial/orbital-sentinels/internal/pkg/logger"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	configFile string
	version    = "1.0.0"
	buildTime  = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "sentinel",
		Short: "Celestial Orbital Sentinel - Data Collection Agent",
		Long:  `A distributed data collection agent for the Celestial monitoring system.`,
	}

	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "config/config.yaml", "config file path")

	// start 命令
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start the sentinel",
		Run:   runStart,
	}

	// version 命令
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Sentinel version %s (built at %s)\n", version, buildTime)
		},
	}

	rootCmd.AddCommand(startCmd, versionCmd, triggerCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runStart(cmd *cobra.Command, args []string) {
	// 加载配置
	cfg, err := config.Load(configFile)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	if err := logger.Init(
		cfg.Logging.Level,
		cfg.Logging.Format,
		cfg.Logging.Output,
		cfg.Logging.FilePath,
	); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting Sentinel",
		zap.String("version", version),
		zap.String("name", cfg.Sentinel.Name),
		zap.String("region", cfg.Sentinel.Region))

	// 创建并启动 Agent
	ag := agent.NewAgent(cfg)
	if err := ag.Start(); err != nil {
		logger.Fatal("Failed to start agent", zap.Error(err))
	}

	// 阻塞等待
	select {}
}

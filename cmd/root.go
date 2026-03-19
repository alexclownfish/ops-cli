package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	// 版本信息
	Version   = "1.1.0"
	BuildDate = "2026-03-19"
	GitCommit = "v1.1.0"
)

var rootCmd = &cobra.Command{
	Use:   "ops",
	Short: "运维工具集",
	Long:  `ops-cli: 集成SSH批量执行、密码管理、OpenStack集成等功能的运维工具`,
	Version: fmt.Sprintf("%s\nBuild: %s\nCommit: %s\nGo: %s\nOS/Arch: %s/%s",
		Version, BuildDate, GitCommit, runtime.Version(), runtime.GOOS, runtime.GOARCH),
}

// Execute 执行根命令
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

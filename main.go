package main

import (
	"fmt"
	"os"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ops",
	Short: "运维工具集 - 整合常用运维功能",
	Long:  `ops-cli: 一个集成了服务器管理、监控、部署、日志分析等功能的运维工具集`,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

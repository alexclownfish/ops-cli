package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "服务器管理",
	Long:  `服务器管理：SSH连接、批量执行命令、文件传输`,
}

var serverListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有服务器",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("服务器列表功能")
	},
}

var serverExecCmd = &cobra.Command{
	Use:   "exec [command]",
	Short: "批量执行命令",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("执行命令: %s\n", args[0])
	},
}

func init() {
	serverCmd.AddCommand(serverListCmd)
	serverCmd.AddCommand(serverExecCmd)
	rootCmd.AddCommand(serverCmd)
}

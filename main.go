package main

import (
	"fmt"
	"os"
	"github.com/spf13/cobra"
	"ops-cli/pkg/ssh"
)

var (
	host     string
	port     int
	user     string
	password string
)

var rootCmd = &cobra.Command{
	Use:   "ops",
	Short: "运维工具集",
	Long:  `ops-cli: 集成SSH批量执行、监控、部署等功能的运维工具`,
}

var execCmd = &cobra.Command{
	Use:   "exec [command]",
	Short: "SSH执行命令",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := ssh.NewClient(host, port, user, password)
		
		fmt.Printf("连接到 %s@%s:%d...\n", user, host, port)
		if err := client.Connect(); err != nil {
			fmt.Printf("❌ 连接失败: %v\n", err)
			os.Exit(1)
		}
		defer client.Close()
		
		fmt.Printf("执行命令: %s\n", args[0])
		output, err := client.Execute(args[0])
		if err != nil {
			fmt.Printf("❌ 执行失败: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Println("--- 输出 ---")
		fmt.Println(output)
	},
}

func init() {
	execCmd.Flags().StringVarP(&host, "host", "H", "", "服务器地址")
	execCmd.Flags().IntVarP(&port, "port", "P", 22, "SSH端口")
	execCmd.Flags().StringVarP(&user, "user", "u", "root", "用户名")
	execCmd.Flags().StringVarP(&password, "password", "p", "", "密码")
	execCmd.MarkFlagRequired("host")
	execCmd.MarkFlagRequired("password")
	
	rootCmd.AddCommand(execCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

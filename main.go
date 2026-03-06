package main

import (
	"fmt"
	"os"
	"sync"
	"github.com/spf13/cobra"
	"ops-cli/pkg/ssh"
	"ops-cli/pkg/config"
)

var (
	configFile string
	keyPath    string
	keyPass    string
	host     string
	port     int
	user     string
	password string
	hosts    []string
	batchPort int
	batchUser string
	batchPass string
	batchKeyPath string
	batchKeyPass string
	parallel int
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
		client := ssh.NewClient(host, port, user, password, keyPath, keyPass)
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

var batchExecCmd = &cobra.Command{
	Use:   "batch [command]",
	Short: "批量执行命令",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		command := args[0]
		
		// 验证：配置文件或命令行参数至少提供一个
		if configFile == "" && len(hosts) == 0 {
			fmt.Println("❌ 错误: 必须提供配置文件(-c)或服务器列表(-L)")
			os.Exit(1)
		}
		
		// 如果指定了配置文件，从配置文件读取服务器列表
		if configFile != "" {
			cfg, err := config.LoadConfig(configFile)
			if err != nil {
				fmt.Printf("❌ 读取配置文件失败: %v\n", err)
				os.Exit(1)
			}
			hosts = []string{}
			for _, srv := range cfg.Servers {
				hosts = append(hosts, srv.Host)
			}
			if len(cfg.Servers) > 0 {
				batchPort = cfg.Servers[0].Port
				batchUser = cfg.Servers[0].User
				batchPass = cfg.Servers[0].Password
			}
		}
		var wg sync.WaitGroup
		results := make(chan string, len(hosts))
		for _, h := range hosts {
			wg.Add(1)
			go func(host string) {
				defer wg.Done()
				client := ssh.NewClient(host, batchPort, batchUser, batchPass, batchKeyPath, batchKeyPass)
				if err := client.Connect(); err != nil {
					results <- fmt.Sprintf("❌ %s: 连接失败 - %v", host, err)
					return
				}
				defer client.Close()
				output, err := client.Execute(command)
				if err != nil {
					results <- fmt.Sprintf("❌ %s: 执行失败 - %v", host, err)
					return
				}
				results <- fmt.Sprintf("✅ %s:\n%s", host, output)
			}(h)
		}
		wg.Wait()
		close(results)
		for result := range results {
			fmt.Println(result)
			fmt.Println("---")
		}
	},
}

func init() {
	execCmd.Flags().StringVarP(&host, "host", "H", "", "服务器地址")
	execCmd.Flags().IntVarP(&port, "port", "P", 22, "SSH端口")
	execCmd.Flags().StringVarP(&user, "user", "u", "root", "用户名")
	execCmd.Flags().StringVarP(&password, "password", "p", "", "密码")
	execCmd.Flags().StringVarP(&keyPath, "key", "i", "", "私钥路径")
	execCmd.Flags().StringVar(&keyPass, "key-pass", "", "私钥密码")
	execCmd.MarkFlagRequired("host")
	execCmd.MarkFlagRequired("password")
	
	batchExecCmd.Flags().StringVarP(&configFile, "config", "c", "", "配置文件路径")
	batchExecCmd.Flags().StringSliceVarP(&hosts, "hosts", "L", []string{}, "服务器列表")
	batchExecCmd.Flags().IntVarP(&batchPort, "batch-port", "T", 22, "SSH端口")
	batchExecCmd.Flags().StringVarP(&batchUser, "batch-user", "U", "root", "用户名")
	batchExecCmd.Flags().StringVarP(&batchPass, "batch-password", "P", "", "密码")
	batchExecCmd.Flags().StringVarP(&batchKeyPath, "batch-key", "K", "", "私钥路径")
	batchExecCmd.Flags().StringVar(&batchKeyPass, "batch-key-pass", "", "私钥密码")
	batchExecCmd.Flags().IntVar(&parallel, "parallel", 10, "并发数")
	// 配置文件和命令行参数二选一，不强制要求
	
	rootCmd.AddCommand(execCmd)
	rootCmd.AddCommand(batchExecCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

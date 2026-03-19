package cmd

import (
	"fmt"
	"os"
	"sync"

	"github.com/spf13/cobra"
	"ops-cli/pkg/config"
	"ops-cli/pkg/ssh"
)

var (
	// exec 命令参数
	execHost     string
	execPort     int
	execUser     string
	execPassword string
	execKeyPath  string
	execKeyPass  string

	// batch 命令参数
	batchConfigFile string
	batchHosts      []string
	batchPort       int
	batchUser       string
	batchPassword   string
	batchKeyPath    string
	batchKeyPass    string
	batchParallel   int
)

var execCmd = &cobra.Command{
	Use:   "exec [command]",
	Short: "SSH执行命令",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := ssh.NewClient(execHost, execPort, execUser, execPassword, execKeyPath, execKeyPass)
		fmt.Printf("连接到 %s@%s:%d...\n", execUser, execHost, execPort)
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
		if batchConfigFile == "" && len(batchHosts) == 0 {
			fmt.Println("❌ 错误: 必须提供配置文件(-c)或服务器列表(-L)")
			os.Exit(1)
		}

		// 如果指定了配置文件，从配置文件读取服务器列表
		if batchConfigFile != "" {
			cfg, err := config.LoadConfig(batchConfigFile)
			if err != nil {
				fmt.Printf("❌ 读取配置文件失败: %v\n", err)
				os.Exit(1)
			}
			batchHosts = []string{}
			for _, srv := range cfg.Servers {
				batchHosts = append(batchHosts, srv.Host)
			}
			if len(cfg.Servers) > 0 {
				batchPort = cfg.Servers[0].Port
				batchUser = cfg.Servers[0].User
				batchPassword = cfg.Servers[0].Password
			}
		}

		// 创建连接池
		pool := ssh.NewPool(batchParallel)
		defer pool.CloseAll()

		var wg sync.WaitGroup
		results := make(chan string, len(batchHosts))

		// 使用信号量控制并发
		sem := make(chan struct{}, batchParallel)

		for _, h := range batchHosts {
			wg.Add(1)
			go func(host string) {
				defer wg.Done()
				sem <- struct{}{} // 获取信号量
				defer func() { <-sem }()

				// 从连接池获取客户端
				client, err := pool.Get(host, batchPort, batchUser, batchPassword, batchKeyPath, batchKeyPass)
				if err != nil {
					results <- fmt.Sprintf("❌ %s: 连接失败 - %v", host, err)
					return
				}

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
	// exec 命令参数
	execCmd.Flags().StringVarP(&execHost, "host", "H", "", "服务器地址")
	execCmd.Flags().IntVarP(&execPort, "port", "P", 22, "SSH端口")
	execCmd.Flags().StringVarP(&execUser, "user", "u", "root", "用户名")
	execCmd.Flags().StringVarP(&execPassword, "password", "p", "", "密码")
	execCmd.Flags().StringVarP(&execKeyPath, "key", "i", "", "私钥路径")
	execCmd.Flags().StringVar(&execKeyPass, "key-pass", "", "私钥密码")
	execCmd.MarkFlagRequired("host")
	execCmd.MarkFlagRequired("password")

	// batch 命令参数
	batchExecCmd.Flags().StringVarP(&batchConfigFile, "config", "c", "", "配置文件路径")
	batchExecCmd.Flags().StringSliceVarP(&batchHosts, "hosts", "L", []string{}, "服务器列表")
	batchExecCmd.Flags().IntVarP(&batchPort, "port", "T", 22, "SSH端口")
	batchExecCmd.Flags().StringVarP(&batchUser, "user", "U", "root", "用户名")
	batchExecCmd.Flags().StringVarP(&batchPassword, "password", "P", "", "密码")
	batchExecCmd.Flags().StringVarP(&batchKeyPath, "key", "K", "", "私钥路径")
	batchExecCmd.Flags().StringVar(&batchKeyPass, "key-pass", "", "私钥密码")
	batchExecCmd.Flags().IntVar(&batchParallel, "parallel", 10, "并发数")

	rootCmd.AddCommand(execCmd)
	rootCmd.AddCommand(batchExecCmd)
}

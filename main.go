package main

import (
	"fmt"
	"os"
	"sync"
	"time"
	"github.com/spf13/cobra"
	"ops-cli/pkg/ssh"
	"ops-cli/pkg/config"
	"ops-cli/pkg/password"
)

var (
	configFile string
	keyPath    string
	keyPass    string
	host     string
	port     int
	user     string
	passwd string
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
		client := ssh.NewClient(host, port, user, passwd, keyPath, keyPass)
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
	execCmd.Flags().StringVarP(&passwd, "password", "p", "", "密码")
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
	
	saveCmd.Flags().StringVarP(&host, "host", "H", "", "服务器地址")
	saveCmd.Flags().StringVarP(&user, "user", "u", "root", "用户名")
	saveCmd.Flags().StringVar(&dbPath, "db", "passwords.db", "数据库路径")
	saveCmd.Flags().StringVar(&masterKey, "key", "", "主密钥")
	saveCmd.MarkFlagRequired("host")
	saveCmd.MarkFlagRequired("key")
	
	passwdCmd.AddCommand(generateCmd)
	showCmd.Flags().StringVar(&dbPath, "db", "passwords.db", "数据库路径")
	showCmd.Flags().StringVar(&masterKey, "key", "", "主密钥")
	showCmd.MarkFlagRequired("key")
	
	passwdCmd.AddCommand(saveCmd)
	listCmd.Flags().StringVar(&dbPath, "db", "passwords.db", "数据库路径")
	listCmd.Flags().StringVar(&masterKey, "key", "", "主密钥")
	listCmd.MarkFlagRequired("key")
	
	passwdCmd.AddCommand(showCmd)
	resetBatchCmd.Flags().StringVar(&dbPath, "db", "passwords.db", "数据库路径")
	resetBatchCmd.Flags().StringVar(&masterKey, "key", "", "主密钥")
	resetBatchCmd.MarkFlagRequired("key")
	
	passwdCmd.AddCommand(listCmd)
	exportCmd.Flags().StringVar(&dbPath, "db", "passwords.db", "数据库路径")
	exportCmd.Flags().StringVar(&masterKey, "key", "", "主密钥")
	exportCmd.Flags().String("output", "passwords.kdbx", "输出文件")
	exportCmd.Flags().String("kdbx-password", "", "KeePass数据库密码")
	exportCmd.MarkFlagRequired("key")
	exportCmd.MarkFlagRequired("kdbx-password")
	
	passwdCmd.AddCommand(resetBatchCmd)
	resetCmd.Flags().StringVar(&dbPath, "db", "passwords.db", "数据库路径")
	resetCmd.Flags().StringVar(&masterKey, "key", "", "主密钥")
	resetCmd.MarkFlagRequired("key")
	
	passwdCmd.AddCommand(exportCmd)
	passwdCmd.AddCommand(resetCmd)
	rootCmd.AddCommand(passwdCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var passwdCmd = &cobra.Command{
	Use:   "passwd",
	Short: "密码管理",
	Long:  `密码生成、加密存储、自动轮换`,
}

var (
	dbPath     string
	masterKey  string
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "生成强密码",
	Run: func(cmd *cobra.Command, args []string) {
		pwd, err := password.Generate(24)
		if err != nil {
			fmt.Printf("❌ 生成失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ 生成密码: %s\n", pwd)
	},
}

var saveCmd = &cobra.Command{
	Use:   "save [server-id]",
	Short: "保存密码",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serverID := args[0]
		
		// 生成新密码
		newPassword, err := password.Generate(24)
		if err != nil {
			fmt.Printf("❌ 生成密码失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ 生成新密码\n")
		
		// 打开数据库
		store, err := password.NewStore(dbPath, masterKey)
		if err != nil {
			fmt.Printf("❌ 打开数据库失败: %v\n", err)
			os.Exit(1)
		}
		defer store.Close()
		
		// 保存密码
		srv := password.Server{
			ID:   serverID,
			Name: serverID,
			Host: host,
			User: user,
		}
		
		if err := store.Save(srv, newPassword); err != nil {
			fmt.Printf("❌ 保存失败: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Printf("✅ 密码已保存到数据库\n")
		fmt.Printf("服务器: %s\n", serverID)
		fmt.Printf("密码: %s\n", newPassword)
	},
}

var showCmd = &cobra.Command{
	Use:   "show [server-id]",
	Short: "查看密码",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serverID := args[0]
		
		store, err := password.NewStore(dbPath, masterKey)
		if err != nil {
			fmt.Printf("❌ 打开数据库失败: %v\n", err)
			os.Exit(1)
		}
		defer store.Close()
		
		srv, pwd, err := store.Get(serverID)
		if err != nil || srv == nil {
			fmt.Printf("❌ 未找到服务器: %s\n", serverID)
			os.Exit(1)
		}
		
		fmt.Printf("服务器ID: %s\n", srv.ID)
		fmt.Printf("名称: %s\n", srv.Name)
		fmt.Printf("地址: %s\n", srv.Host)
		fmt.Printf("用户: %s\n", srv.User)
		fmt.Printf("密码: %s\n", pwd)
		fmt.Printf("创建时间: %s\n", srv.CreatedAt.Format("2006-01-02 15:04:05"))
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有服务器",
	Run: func(cmd *cobra.Command, args []string) {
		store, err := password.NewStore(dbPath, masterKey)
		if err != nil {
			fmt.Printf("❌ 打开数据库失败: %v\n", err)
			os.Exit(1)
		}
		defer store.Close()
		
		servers, err := store.List()
		if err != nil {
			fmt.Printf("❌ 查询失败: %v\n", err)
			os.Exit(1)
		}
		
		if len(servers) == 0 {
			fmt.Println("暂无服务器")
			return
		}
		
		fmt.Printf("总共 %d 台服务器:\n\n", len(servers))
		for _, srv := range servers {
			fmt.Printf("ID: %s\n", srv.ID)
			fmt.Printf("  名称: %s\n", srv.Name)
			fmt.Printf("  地址: %s\n", srv.Host)
			fmt.Printf("  用户: %s\n", srv.User)
			fmt.Printf("  更新: %s\n\n", srv.UpdatedAt.Format("2006-01-02 15:04:05"))
		}
	},
}

var resetBatchCmd = &cobra.Command{
	Use:   "reset-batch",
	Short: "批量改密",
	Run: func(cmd *cobra.Command, args []string) {
		store, err := password.NewStore(dbPath, masterKey)
		if err != nil {
			fmt.Printf("❌ 打开数据库失败: %v\n", err)
			os.Exit(1)
		}
		defer store.Close()
		
		servers, err := store.List()
		if err != nil {
			fmt.Printf("❌ 查询失败: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Printf("开始批量改密，共 %d 台服务器\n\n", len(servers))
		
		for _, srv := range servers {
			fmt.Printf("处理: %s (%s)\n", srv.ID, srv.Host)
			
			// 获取当前密码
			_, oldPwd, _ := store.Get(srv.ID)
			
			// 生成新密码
			newPwd, _ := password.Generate(24)
			
			// 改密
			err := password.ResetPassword(srv.Host, 22, srv.User, oldPwd, newPwd)
			if err != nil {
				fmt.Printf("  ❌ 改密失败: %v\n\n", err)
				continue
			}
			
			// 更新数据库
			srv.UpdatedAt = time.Now()
			store.Save(srv, newPwd)
			
			fmt.Printf("  ✅ 改密成功\n\n")
		}
	},
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "导出到KeePassXC",
	Run: func(cmd *cobra.Command, args []string) {
		store, err := password.NewStore(dbPath, masterKey)
		if err != nil {
			fmt.Printf("❌ 打开数据库失败: %v\n", err)
			os.Exit(1)
		}
		defer store.Close()
		
		servers, _ := store.List()
		
		outputPath, _ := cmd.Flags().GetString("output")
		kdbxPass, _ := cmd.Flags().GetString("kdbx-password")
		
		err = password.ExportToKeePass(servers, masterKey, outputPath, kdbxPass)
		if err != nil {
			fmt.Printf("❌ 导出失败: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Printf("✅ 已导出 %d 台服务器到 %s\n", len(servers), outputPath)
	},
}

var resetCmd = &cobra.Command{
	Use:   "reset [server-id]",
	Short: "修改密码",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serverID := args[0]
		
		store, err := password.NewStore(dbPath, masterKey)
		if err != nil {
			fmt.Printf("❌ 打开数据库失败: %v\n", err)
			os.Exit(1)
		}
		defer store.Close()
		
		srv, oldPwd, err := store.Get(serverID)
		if err != nil || srv == nil {
			fmt.Printf("❌ 未找到服务器: %s\n", serverID)
			os.Exit(1)
		}
		
		newPwd, _ := password.Generate(24)
		fmt.Printf("生成新密码: %s\n", newPwd)
		
		if srv.ResetMethod == "virsh" {
			fmt.Printf("使用virsh方式改密...\n")
			err = password.ResetPasswordVirsh(
				srv.HypervisorHost,
				srv.HypervisorPort,
				srv.HypervisorUser,
				srv.HypervisorPass,
				srv.InstanceID,
				srv.User,
				newPwd,
			)
		} else {
			fmt.Printf("使用SSH方式改密...\n")
			err = password.ResetPassword(srv.Host, 22, srv.User, oldPwd, newPwd)
		}
		
		if err != nil {
			fmt.Printf("❌ 改密失败: %v\n", err)
			os.Exit(1)
		}
		
		srv.UpdatedAt = time.Now()
		store.Save(*srv, newPwd)
		
		fmt.Printf("✅ 改密成功\n")
	},
}

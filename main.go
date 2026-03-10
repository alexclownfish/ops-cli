package main

import (
	"fmt"
	"os"
	"sync"
	"time"
	"ops-cli/pkg/openstack"
	"github.com/spf13/cobra"
	"ops-cli/pkg/ssh"
	"ops-cli/pkg/config"
	"ops-cli/pkg/password"
)

var (
	configFile string
	
	// virsh相关参数
	resetMethod      string
	instanceID       string
	hypervisorHost   string
	hypervisorPort   int
	hypervisorUser   string
	hypervisorPass   string
	hypervisorKey    string
	hypervisorKeyPass string
	vmUser           string
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
	saveCmd.Flags().StringVar(&resetMethod, "method", "ssh", "改密方式: ssh|virsh")
	saveCmd.Flags().StringVar(&instanceID, "instance-id", "", "虚拟机实例ID(virsh)")
	saveCmd.Flags().StringVar(&hypervisorHost, "hypervisor-host", "", "物理机地址(virsh)")
	saveCmd.Flags().IntVar(&hypervisorPort, "hypervisor-port", 22, "物理机SSH端口(virsh)")
	saveCmd.Flags().StringVar(&hypervisorUser, "hypervisor-user", "root", "物理机用户(virsh)")
	saveCmd.Flags().StringVar(&hypervisorPass, "hypervisor-pass", "", "物理机密码(virsh)")
	saveCmd.Flags().StringVar(&hypervisorKey, "hypervisor-key", "", "物理机密钥路径(virsh)")
	saveCmd.Flags().StringVar(&hypervisorKeyPass, "hypervisor-key-pass", "", "物理机密码(virsh)")
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
	importCmd.Flags().StringVar(&dbPath, "db", "passwords.db", "数据库路径")
	importCmd.Flags().StringVar(&masterKey, "key", "", "主密钥")
	importCmd.Flags().StringVar(&hypervisorHost, "hypervisor-host", "", "物理机地址")
	importCmd.Flags().IntVar(&hypervisorPort, "hypervisor-port", 22, "物理机SSH端口")
	importCmd.Flags().StringVar(&hypervisorUser, "hypervisor-user", "root", "物理机用户")
	importCmd.Flags().StringVar(&hypervisorPass, "hypervisor-pass", "", "物理机密码")
	importCmd.Flags().StringVar(&hypervisorKey, "hypervisor-key", "", "物理机密钥路径")
	importCmd.Flags().StringVar(&hypervisorKeyPass, "hypervisor-key-pass", "", "物理机密钥密码")
	importCmd.Flags().StringVar(&vmUser, "vm-user", "", "物理机密钥密码")
	importCmd.MarkFlagRequired("key")
	// hypervisor-host不再必填，自动从OpenStack获取
	
	passwdCmd.AddCommand(resetCmd)
	deleteCmd.Flags().StringVar(&dbPath, "db", "passwords.db", "数据库路径")
	deleteCmd.Flags().StringVar(&masterKey, "key", "", "主密钥")
	deleteCmd.MarkFlagRequired("key")
	
	passwdCmd.AddCommand(importCmd)
	passwdCmd.AddCommand(deleteCmd)
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
			ID:                serverID,
			Name:              serverID,
			Host:              host,
			User:              user,
			PasswordEncrypted: "",
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
			ResetMethod:       resetMethod,
			InstanceID:        instanceID,
			HypervisorHost:    hypervisorHost,
			HypervisorPort:    hypervisorPort,
			HypervisorUser:    hypervisorUser,
			HypervisorPass:    hypervisorPass,
			HypervisorKey:     hypervisorKey,
			HypervisorKeyPass: hypervisorKeyPass,
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
			
			// 智能改密
			var resetErr error
			
			// 优先SSH
			if oldPwd != "" {
				resetErr = password.ResetPassword(srv.Host, 22, srv.User, oldPwd, newPwd)
				if resetErr == nil {
					fmt.Printf("  ✅ SSH改密成功\n\n")
					srv.ResetMethod = "ssh"
					srv.UpdatedAt = time.Now()
					store.Save(srv, newPwd)
					continue
				}
			}
			
			// 回退virsh
			if srv.HypervisorHost != "" {
				resetErr = password.ResetPasswordVirsh(
					srv.HypervisorHost, srv.HypervisorPort,
					srv.HypervisorUser, srv.HypervisorPass,
					srv.HypervisorKey, srv.HypervisorKeyPass,
					srv.InstanceID, srv.User, newPwd,
				)
				if resetErr == nil {
					fmt.Printf("  ✅ virsh改密成功\n\n")
					srv.ResetMethod = "virsh"
					srv.UpdatedAt = time.Now()
					store.Save(srv, newPwd)
					continue
				}
			}
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
		
		// 智能选择改密方式
		var resetErr error
		
		// 优先SSH
		if oldPwd != "" {
			fmt.Printf("尝试SSH方式改密...\n")
			resetErr = password.ResetPassword(srv.Host, 22, srv.User, oldPwd, newPwd)
			if resetErr == nil {
				fmt.Printf("✅ SSH改密成功\n")
				srv.ResetMethod = "ssh"
				srv.UpdatedAt = time.Now()
				store.Save(*srv, newPwd)
				return
			}
			fmt.Printf("⚠️  SSH改密失败: %v\n", resetErr)
		}
		
		// 回退virsh
		if srv.HypervisorHost != "" && srv.InstanceID != "" {
			fmt.Printf("尝试virsh方式改密...\n")
			resetErr = password.ResetPasswordVirsh(
				srv.HypervisorHost, srv.HypervisorPort,
				srv.HypervisorUser, srv.HypervisorPass,
				srv.HypervisorKey, srv.HypervisorKeyPass,
				srv.InstanceID, srv.User, newPwd,
			)
			if resetErr == nil {
				fmt.Printf("✅ virsh改密成功\n")
				srv.ResetMethod = "virsh"
				srv.UpdatedAt = time.Now()
				store.Save(*srv, newPwd)
				return
			}
			fmt.Printf("⚠️  virsh改密失败: %v\n", resetErr)
		}
		
		fmt.Printf("❌ 改密失败: 所有方式都已尝试\n")
		os.Exit(1)
	},
}

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "从OpenStack导入虚拟机",
	Run: func(cmd *cobra.Command, args []string) {
		// 创建OpenStack客户端
		client := openstack.NewClient()
		
		fmt.Println("正在连接OpenStack...")
		if err := client.Authenticate(); err != nil {
			fmt.Printf("❌ 认证失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ 认证成功")
		
		// 获取物理机映射
		fmt.Println("正在获取物理机列表...")
		if err := client.GetHypervisorMap(); err != nil {
			fmt.Printf("❌ 获取物理机失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ 找到 %d 台物理机\n", len(client.HypervisorMap))
		
		// 获取虚拟机列表
		fmt.Println("正在获取虚拟机列表...")
		vms, err := client.ListVMs()
		if err != nil {
			fmt.Printf("❌ 获取失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ 找到 %d 台虚拟机\n\n", len(vms))
		
		// 打开数据库
		store, err := password.NewStore(dbPath, masterKey)
		if err != nil {
			fmt.Printf("❌ 打开数据库失败: %v\n", err)
			os.Exit(1)
		}
		defer store.Close()
		
		// 批量导入
		successCount := 0
		for _, vm := range vms {
			fmt.Printf("导入: %s (%s)\n", vm.Name, vm.IP)
			
			// 生成密码
			newPassword, _ := password.Generate(24)
			
			// 创建服务器记录
			srv := password.Server{
				ID:                vm.ID,
				Name:              vm.Name,
				Host:              vm.IP,
				User:              vmUser,
				PasswordEncrypted: "",
				CreatedAt:         time.Now(),
				UpdatedAt:         time.Now(),
				ResetMethod:       "virsh",
				InstanceID:        vm.InstanceID,
				HypervisorHost:    vm.HypervisorHost,
				HypervisorPort:    hypervisorPort,
				HypervisorUser:    hypervisorUser,
				HypervisorPass:    hypervisorPass,
				HypervisorKey:     hypervisorKey,
				HypervisorKeyPass: hypervisorKeyPass,
			}
			
			if err := store.Save(srv, newPassword); err != nil {
				fmt.Printf("  ❌ 保存失败: %v\n", err)
				continue
			}
			
			fmt.Printf("  ✅ 已导入\n")
			successCount++
		}
		
		fmt.Printf("\n导入完成！成功 %d/%d 台\n", successCount, len(vms))
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete [server-id]",
	Short: "删除服务器",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serverID := args[0]
		
		store, err := password.NewStore(dbPath, masterKey)
		if err != nil {
			fmt.Printf("❌ 打开数据库失败: %v\n", err)
			os.Exit(1)
		}
		defer store.Close()
		
		if err := store.Delete(serverID); err != nil {
			fmt.Printf("❌ 删除失败: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Printf("✅ 已删除服务器: %s\n", serverID)
	},
}

package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"go.etcd.io/bbolt"
	"ops-cli/pkg/openstack"
	"ops-cli/pkg/password"
)

var (
	// 通用参数
	dbPath    string
	masterKey string

	// save 命令参数
	saveHost           string
	saveUser           string
	saveResetMethod    string
	saveInstanceID     string
	saveHypervisorHost string
	saveHypervisorPort int
	saveHypervisorUser string
	saveHypervisorPass string
	saveHypervisorKey  string
	saveHypervisorKeyPass string

	// import 命令参数
	importHypervisorPort    int
	importHypervisorUser    string
	importHypervisorPass    string
	importHypervisorKey     string
	importHypervisorKeyPass string
	importVmUser            string
)

var passwdCmd = &cobra.Command{
	Use:   "passwd",
	Short: "密码管理",
	Long:  `密码生成、加密存储、自动轮换`,
}

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
			Host:              saveHost,
			User:              saveUser,
			PasswordEncrypted: "",
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
			ResetMethod:       saveResetMethod,
			InstanceID:        saveInstanceID,
			HypervisorHost:    saveHypervisorHost,
			HypervisorPort:    saveHypervisorPort,
			HypervisorUser:    saveHypervisorUser,
			HypervisorPass:    saveHypervisorPass,
			HypervisorKey:     saveHypervisorKey,
			HypervisorKeyPass: saveHypervisorKeyPass,
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
		if srv.InstanceID != "" {
			fmt.Printf("虚拟机ID: %s\n", srv.InstanceID)
		}
		if srv.HypervisorHost != "" {
			fmt.Printf("物理机IP: %s\n", srv.HypervisorHost)
		}
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
			if srv.InstanceID != "" {
				fmt.Printf("  虚拟机ID: %s\n", srv.InstanceID)
			}
			if srv.HypervisorHost != "" {
				fmt.Printf("  物理机IP: %s\n", srv.HypervisorHost)
			}
			fmt.Printf("  更新: %s\n\n", srv.UpdatedAt.Format("2006-01-02 15:04:05"))
		}
	},
}

var updateCmd = &cobra.Command{
	Use:   "update [server-id]",
	Short: "更新服务器信息",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serverID := args[0]

		store, err := password.NewStore(dbPath, masterKey)
		if err != nil {
			fmt.Printf("❌ 打开数据库失败: %v\n", err)
			os.Exit(1)
		}
		defer store.Close()

		_, _, err = store.Get(serverID)
		if err != nil {
			fmt.Printf("❌ 服务器不存在: %s\n", serverID)
			os.Exit(1)
		}

		updates := make(map[string]interface{})

		if cmd.Flags().Changed("name") {
			name, _ := cmd.Flags().GetString("name")
			updates["name"] = name
		}
		if cmd.Flags().Changed("host") {
			host, _ := cmd.Flags().GetString("host")
			updates["host"] = host
		}
		if cmd.Flags().Changed("user") {
			user, _ := cmd.Flags().GetString("user")
			updates["user"] = user
		}
		if cmd.Flags().Changed("password") {
			pwd, _ := cmd.Flags().GetString("password")
			if pwd == "" {
				pwd, err = password.Generate(24)
				if err != nil {
					fmt.Printf("❌ 生成密码失败: %v\n", err)
					os.Exit(1)
				}
				fmt.Printf("✅ 生成新密码: %s\n", pwd)
			}
			updates["password"] = pwd
		}
		if cmd.Flags().Changed("instance-id") {
			instanceID, _ := cmd.Flags().GetString("instance-id")
			updates["instance_id"] = instanceID
		}
		if cmd.Flags().Changed("hypervisor-host") {
			hypervisorHost, _ := cmd.Flags().GetString("hypervisor-host")
			updates["hypervisor_host"] = hypervisorHost
		}

		if len(updates) == 0 {
			fmt.Println("❌ 没有指定要更新的字段")
			os.Exit(1)
		}

		if err := store.Update(serverID, updates); err != nil {
			fmt.Printf("❌ 更新失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ 已更新服务器: %s\n", serverID)
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
				User:              importVmUser,
				PasswordEncrypted: "",
				CreatedAt:         time.Now(),
				UpdatedAt:         time.Now(),
				ResetMethod:       "virsh",
				InstanceID:        vm.InstanceID,
				HypervisorHost:    vm.HypervisorHost,
				HypervisorPort:    importHypervisorPort,
				HypervisorUser:    importHypervisorUser,
				HypervisorPass:    importHypervisorPass,
				HypervisorKey:     importHypervisorKey,
				HypervisorKeyPass: importHypervisorKeyPass,
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

var resetKeyCmd = &cobra.Command{
	Use:   "reset-key",
	Short: "更换主密钥（需要旧密钥验证）",
	Run: func(cmd *cobra.Command, args []string) {
		oldKey, _ := cmd.Flags().GetString("old-key")
		newKey, _ := cmd.Flags().GetString("new-key")

		if oldKey == "" || newKey == "" {
			fmt.Printf("❌ 必须提供 --old-key 和 --new-key\n")
			os.Exit(1)
		}

		// 1. 验证旧key
		oldStore, err := password.NewStore(dbPath, oldKey)
		if err != nil {
			fmt.Printf("❌ 旧密钥验证失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ 旧密钥验证成功\n")

		// 2. 获取所有服务器
		servers, err := oldStore.List()
		if err != nil {
			fmt.Printf("❌ 获取服务器列表失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("正在重新加密 %d 台服务器的密码...\n", len(servers))

		// 3. 解密所有密码
		type ServerWithPassword struct {
			Server   password.Server
			Password string
		}
		var serversWithPasswords []ServerWithPassword

		for _, srv := range servers {
			_, pwd, err := oldStore.Get(srv.ID)
			if err != nil {
				fmt.Printf("⚠️  跳过 %s: %v\n", srv.ID, err)
				continue
			}
			serversWithPasswords = append(serversWithPasswords, ServerWithPassword{
				Server:   srv,
				Password: pwd,
			})
		}

		oldStore.Close()

		// 4. 删除key哈希
		db, err := bbolt.Open(dbPath, 0600, nil)
		if err != nil {
			fmt.Printf("❌ 打开数据库失败: %v\n", err)
			os.Exit(1)
		}

		err = db.Update(func(tx *bbolt.Tx) error {
			return tx.DeleteBucket([]byte("key_hash"))
		})
		db.Close()

		if err != nil {
			fmt.Printf("❌ 删除旧哈希失败: %v\n", err)
			os.Exit(1)
		}

		// 5. 用新key重新加密
		newStore, err := password.NewStore(dbPath, newKey)
		if err != nil {
			fmt.Printf("❌ 创建新密钥失败: %v\n", err)
			os.Exit(1)
		}
		defer newStore.Close()

		for _, item := range serversWithPasswords {
			err := newStore.Save(item.Server, item.Password)
			if err != nil {
				fmt.Printf("⚠️  保存 %s 失败: %v\n", item.Server.ID, err)
			} else {
				fmt.Printf("  ✅ %s\n", item.Server.ID)
			}
		}

		fmt.Printf("\n✅ 主密钥更换完成！\n")
		fmt.Printf("⚠️  请妥善保管新密钥，旧密钥已失效\n")
	},
}

var checkAgeCmd = &cobra.Command{
	Use:   "check-age",
	Short: "检查密码年龄",
	Run: func(cmd *cobra.Command, args []string) {
		days, _ := cmd.Flags().GetInt("days")

		store, err := password.NewStore(dbPath, masterKey)
		if err != nil {
			fmt.Printf("❌ 打开数据库失败: %v\n", err)
			os.Exit(1)
		}
		defer store.Close()

		servers, err := store.List()
		if err != nil {
			fmt.Printf("❌ 获取服务器列表失败: %v\n", err)
			os.Exit(1)
		}

		var needRotate []password.Server
		now := time.Now()

		for _, srv := range servers {
			age := int(now.Sub(srv.UpdatedAt).Hours() / 24)
			if age >= days {
				needRotate = append(needRotate, srv)
			}
		}

		if len(needRotate) == 0 {
			fmt.Printf("✅ 所有服务器密码都在有效期内\n")
			return
		}

		fmt.Printf("需要改密的服务器（%d台）:\n", len(needRotate))
		for _, srv := range needRotate {
			age := int(now.Sub(srv.UpdatedAt).Hours() / 24)
			fmt.Printf("- %s (已使用%d天)\n", srv.ID, age)
		}
	},
}

var autoRotateCmd = &cobra.Command{
	Use:   "auto-rotate",
	Short: "自动改密（密码到期）",
	Run: func(cmd *cobra.Command, args []string) {
		days, _ := cmd.Flags().GetInt("days")

		store, err := password.NewStore(dbPath, masterKey)
		if err != nil {
			fmt.Printf("❌ 打开数据库失败: %v\n", err)
			os.Exit(1)
		}
		defer store.Close()

		servers, err := store.List()
		if err != nil {
			fmt.Printf("❌ 获取服务器列表失败: %v\n", err)
			os.Exit(1)
		}

		var needRotate []password.Server
		now := time.Now()

		for _, srv := range servers {
			age := int(now.Sub(srv.UpdatedAt).Hours() / 24)
			if age >= days {
				needRotate = append(needRotate, srv)
			}
		}

		if len(needRotate) == 0 {
			fmt.Printf("✅ 所有服务器密码都在有效期内\n")
			return
		}

		fmt.Printf("发现 %d 台服务器需要改密\n\n", len(needRotate))

		successCount := 0
		for _, srv := range needRotate {
			fmt.Printf("改密: %s\n", srv.ID)

			_, oldPwd, err := store.Get(srv.ID)
			if err != nil {
				fmt.Printf("  ❌ 获取密码失败: %v\n\n", err)
				continue
			}

			newPwd, _ := password.Generate(24)

			// 智能改密
			var resetErr error

			// 优先SSH
			if oldPwd != "" {
				fmt.Printf("  尝试SSH方式改密...\n")
				resetErr = password.ResetPassword(srv.Host, 22, srv.User, oldPwd, newPwd)
				if resetErr == nil {
					fmt.Printf("  ✅ SSH改密成功\n\n")
					srv.ResetMethod = "ssh"
					srv.UpdatedAt = time.Now()
					store.Save(srv, newPwd)
					successCount++
					continue
				}
				fmt.Printf("  ⚠️  SSH改密失败: %v\n", resetErr)
			}

			// 回退virsh
			if srv.HypervisorHost != "" {
				fmt.Printf("  尝试virsh方式改密...\n")
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
					successCount++
					continue
				}
				fmt.Printf("  ⚠️  virsh改密失败: %v\n", resetErr)
			}

			fmt.Printf("  ❌ 改密失败\n\n")
		}

		fmt.Printf("✅ 改密完成: 成功 %d/%d 台\n", successCount, len(needRotate))
	},
}

func init() {
	// passwd 命令
	rootCmd.AddCommand(passwdCmd)

	// generate 子命令
	passwdCmd.AddCommand(generateCmd)

	// save 子命令
	saveCmd.Flags().StringVarP(&saveHost, "host", "H", "", "服务器地址")
	saveCmd.Flags().StringVarP(&saveUser, "user", "u", "root", "用户名")
	saveCmd.Flags().StringVar(&dbPath, "db", "passwords.db", "数据库路径")
	saveCmd.Flags().StringVar(&masterKey, "key", "", "主密钥")
	saveCmd.Flags().StringVar(&saveResetMethod, "method", "ssh", "改密方式: ssh|virsh")
	saveCmd.Flags().StringVar(&saveInstanceID, "instance-id", "", "虚拟机实例ID(virsh)")
	saveCmd.Flags().StringVar(&saveHypervisorHost, "hypervisor-host", "", "物理机地址(virsh)")
	saveCmd.Flags().IntVar(&saveHypervisorPort, "hypervisor-port", 22, "物理机SSH端口(virsh)")
	saveCmd.Flags().StringVar(&saveHypervisorUser, "hypervisor-user", "root", "物理机用户(virsh)")
	saveCmd.Flags().StringVar(&saveHypervisorPass, "hypervisor-pass", "", "物理机密码(virsh)")
	saveCmd.Flags().StringVar(&saveHypervisorKey, "hypervisor-key", "", "物理机密钥路径(virsh)")
	saveCmd.Flags().StringVar(&saveHypervisorKeyPass, "hypervisor-key-pass", "", "物理机密钥密码(virsh)")
	saveCmd.MarkFlagRequired("host")
	saveCmd.MarkFlagRequired("key")
	passwdCmd.AddCommand(saveCmd)

	// show 子命令
	showCmd.Flags().StringVar(&dbPath, "db", "passwords.db", "数据库路径")
	showCmd.Flags().StringVar(&masterKey, "key", "", "主密钥")
	showCmd.MarkFlagRequired("key")
	passwdCmd.AddCommand(showCmd)

	// list 子命令
	listCmd.Flags().StringVar(&dbPath, "db", "passwords.db", "数据库路径")
	listCmd.Flags().StringVar(&masterKey, "key", "", "主密钥")
	listCmd.MarkFlagRequired("key")
	passwdCmd.AddCommand(listCmd)

	// update 子命令
	updateCmd.Flags().String("name", "", "服务器名称")
	updateCmd.Flags().String("host", "", "服务器地址")
	updateCmd.Flags().String("user", "", "用户名")
	updateCmd.Flags().String("password", "", "密码（留空自动生成）")
	updateCmd.Flags().String("instance-id", "", "虚拟机ID")
	updateCmd.Flags().String("hypervisor-host", "", "物理机IP")
	updateCmd.Flags().StringVar(&masterKey, "key", "", "主密钥")
	updateCmd.MarkFlagRequired("key")
	passwdCmd.AddCommand(updateCmd)

	// delete 子命令
	deleteCmd.Flags().StringVar(&dbPath, "db", "passwords.db", "数据库路径")
	deleteCmd.Flags().StringVar(&masterKey, "key", "", "主密钥")
	deleteCmd.MarkFlagRequired("key")
	passwdCmd.AddCommand(deleteCmd)

	// reset 子命令
	resetCmd.Flags().StringVar(&dbPath, "db", "passwords.db", "数据库路径")
	resetCmd.Flags().StringVar(&masterKey, "key", "", "主密钥")
	resetCmd.MarkFlagRequired("key")
	passwdCmd.AddCommand(resetCmd)

	// reset-batch 子命令
	resetBatchCmd.Flags().StringVar(&dbPath, "db", "passwords.db", "数据库路径")
	resetBatchCmd.Flags().StringVar(&masterKey, "key", "", "主密钥")
	resetBatchCmd.MarkFlagRequired("key")
	passwdCmd.AddCommand(resetBatchCmd)

	// import 子命令
	importCmd.Flags().StringVar(&dbPath, "db", "passwords.db", "数据库路径")
	importCmd.Flags().StringVar(&masterKey, "key", "", "主密钥")
	importCmd.Flags().IntVar(&importHypervisorPort, "hypervisor-port", 22, "物理机SSH端口")
	importCmd.Flags().StringVar(&importHypervisorUser, "hypervisor-user", "root", "物理机用户")
	importCmd.Flags().StringVar(&importHypervisorPass, "hypervisor-pass", "", "物理机密码")
	importCmd.Flags().StringVar(&importHypervisorKey, "hypervisor-key", "", "物理机密钥路径")
	importCmd.Flags().StringVar(&importHypervisorKeyPass, "hypervisor-key-pass", "", "物理机密钥密码")
	importCmd.Flags().StringVar(&importVmUser, "vm-user", "", "虚拟机用户")
	importCmd.MarkFlagRequired("key")
	passwdCmd.AddCommand(importCmd)

	// export 子命令
	exportCmd.Flags().StringVar(&dbPath, "db", "passwords.db", "数据库路径")
	exportCmd.Flags().StringVar(&masterKey, "key", "", "主密钥")
	exportCmd.Flags().String("output", "passwords.kdbx", "输出文件")
	exportCmd.Flags().String("kdbx-password", "", "KeePass数据库密码")
	exportCmd.MarkFlagRequired("key")
	exportCmd.MarkFlagRequired("kdbx-password")
	passwdCmd.AddCommand(exportCmd)

	// reset-key 子命令
	resetKeyCmd.Flags().StringVar(&dbPath, "db", "passwords.db", "数据库路径")
	resetKeyCmd.Flags().String("old-key", "", "旧主密钥")
	resetKeyCmd.Flags().String("new-key", "", "新主密钥")
	resetKeyCmd.MarkFlagRequired("old-key")
	resetKeyCmd.MarkFlagRequired("new-key")
	passwdCmd.AddCommand(resetKeyCmd)

	// check-age 子命令
	checkAgeCmd.Flags().StringVar(&dbPath, "db", "passwords.db", "数据库路径")
	checkAgeCmd.Flags().StringVar(&masterKey, "key", "", "主密钥")
	checkAgeCmd.Flags().Int("days", 85, "密码有效期（天）")
	checkAgeCmd.MarkFlagRequired("key")
	passwdCmd.AddCommand(checkAgeCmd)

	// auto-rotate 子命令
	autoRotateCmd.Flags().StringVar(&dbPath, "db", "passwords.db", "数据库路径")
	autoRotateCmd.Flags().StringVar(&masterKey, "key", "", "主密钥")
	autoRotateCmd.Flags().Int("days", 85, "密码有效期（天）")
	autoRotateCmd.MarkFlagRequired("key")
	passwdCmd.AddCommand(autoRotateCmd)
}

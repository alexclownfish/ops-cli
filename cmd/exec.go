package cmd

import (
	"fmt"
	"os"
	"sync"

	"github.com/spf13/cobra"
	"ops-cli/pkg/config"
	"ops-cli/pkg/password"
	"ops-cli/pkg/ssh"
)

var (
	// exec 命令参数
	execHost      string
	execPort      int
	execUser      string
	execPassword  string
	execKeyPath   string
	execKeyPass   string
	execFromDB    string
	execDBPath    string
	execKey string

	// batch 命令参数
	batchConfigFile string
	batchHosts      []string
	batchPort       int
	batchUser       string
	batchPassword   string
	batchKeyPath    string
	batchKeyPass    string
	batchParallel   int
	batchFromDB     bool // 批量从数据库读取账密

	// scp 命令参数
	scpHost      string
	scpPort      int
	scpUser      string
	scpPassword  string
	scpKeyPath   string
	scpKeyPass   string
	scpFromDB    string
	scpRemote    string
	scpHosts     []string
	scpBatchPort int
	scpBatchUser string
	scpBatchPass string
	scpParallel  int
	scpBatchDB   bool
	scpDBPath    string
	scpKey string

	// batch 独立 db/key
	batchFromDBPath string
	batchKey  string
)

// loadServerFromDB 从数据库加载服务器信息
func loadServerFromDB(dbPath, masterKey, serverID string) (*password.Server, string, error) {
	store, err := password.NewStore(dbPath, masterKey)
	if err != nil {
		return nil, "", fmt.Errorf("打开数据库失败: %v", err)
	}
	defer store.Close()

	srv, pwd, err := store.Get(serverID)
	if err != nil || srv == nil {
		return nil, "", fmt.Errorf("未找到服务器: %s", serverID)
	}
	return srv, pwd, nil
}

var execCmd = &cobra.Command{
	Use:   "exec [command]",
	Short: "SSH执行命令",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		host, port, user, passwd := execHost, execPort, execUser, execPassword

		// 从数据库读取账密
		if execFromDB != "" {
			srv, pwd, err := loadServerFromDB(execDBPath, execKey, execFromDB)
			if err != nil {
				fmt.Printf("❌ %v\n", err)
				os.Exit(1)
			}
			host = srv.Host
			user = srv.User
			passwd = pwd
			if port == 22 {
				port = 22
			}
			fmt.Printf("📦 从数据库加载: %s (%s@%s)\n", execFromDB, user, host)
		}

		client := ssh.NewClient(host, port, user, passwd, execKeyPath, execKeyPass)
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

		type hostInfo struct {
			host     string
			port     int
			user     string
			password string
			keyPath  string
			keyPass  string
		}

		var targets []hostInfo

		// 从数据库批量读取
		if batchFromDB {
			if batchKey == "" {
				fmt.Println("❌ 使用 --from-db 需要提供 --key 主密钥")
				os.Exit(1)
			}
			store, err := password.NewStore(batchFromDBPath, batchKey)
			if err != nil {
				fmt.Printf("❌ 打开数据库失败: %v\n", err)
				os.Exit(1)
			}
			servers, err := store.List()
			store.Close()
			if err != nil {
				fmt.Printf("❌ 读取服务器列表失败: %v\n", err)
				os.Exit(1)
			}

			store2, _ := password.NewStore(batchFromDBPath, batchKey)
			defer store2.Close()
			for _, srv := range servers {
				_, pwd, err := store2.Get(srv.ID)
				if err != nil {
					continue
				}
				targets = append(targets, hostInfo{
					host:     srv.Host,
					port:     22,
					user:     srv.User,
					password: pwd,
				})
			}
			fmt.Printf("📦 从数据库加载 %d 台服务器\n\n", len(targets))
		} else {
			// 验证：配置文件或命令行参数至少提供一个
			if batchConfigFile == "" && len(batchHosts) == 0 {
				fmt.Println("❌ 错误: 必须提供配置文件(-c)、服务器列表(-L)或 --from-db")
				os.Exit(1)
			}

			// 从配置文件读取
			if batchConfigFile != "" {
				cfg, err := config.LoadConfig(batchConfigFile)
				if err != nil {
					fmt.Printf("❌ 读取配置文件失败: %v\n", err)
					os.Exit(1)
				}
				for _, srv := range cfg.Servers {
					targets = append(targets, hostInfo{
						host:     srv.Host,
						port:     srv.Port,
						user:     srv.User,
						password: srv.Password,
					})
				}
			} else {
				for _, h := range batchHosts {
					targets = append(targets, hostInfo{
						host:     h,
						port:     batchPort,
						user:     batchUser,
						password: batchPassword,
						keyPath:  batchKeyPath,
						keyPass:  batchKeyPass,
					})
				}
			}
		}

		// 创建连接池
		pool := ssh.NewPool(batchParallel)
		defer pool.CloseAll()

		var wg sync.WaitGroup
		results := make(chan string, len(targets))
		sem := make(chan struct{}, batchParallel)

		for _, t := range targets {
			wg.Add(1)
			go func(t hostInfo) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				client, err := pool.Get(t.host, t.port, t.user, t.password, t.keyPath, t.keyPass)
				if err != nil {
					results <- fmt.Sprintf("❌ %s: 连接失败 - %v", t.host, err)
					return
				}

				output, err := client.Execute(command)
				if err != nil {
					results <- fmt.Sprintf("❌ %s: 执行失败 - %v", t.host, err)
					return
				}
				results <- fmt.Sprintf("✅ %s:\n%s", t.host, output)
			}(t)
		}

		wg.Wait()
		close(results)

		for result := range results {
			fmt.Println(result)
			fmt.Println("---")
		}
	},
}

var scpCmd = &cobra.Command{
	Use:   "scp [local-path]",
	Short: "文件分发（单机或批量）",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		localPath := args[0]

		if scpRemote == "" {
			fmt.Println("❌ 必须指定远端路径 --remote")
			os.Exit(1)
		}

		type hostInfo struct {
			id       string
			host     string
			port     int
			user     string
			password string
			keyPath  string
			keyPass  string
		}

		var targets []hostInfo

		// 批量从数据库读取
		if scpBatchDB {
			if scpKey == "" {
				fmt.Println("❌ 使用 --all-from-db 需要提供 --key 主密钥")
				os.Exit(1)
			}
			store, err := password.NewStore(scpDBPath, scpKey)
			if err != nil {
				fmt.Printf("❌ 打开数据库失败: %v\n", err)
				os.Exit(1)
			}
			servers, err := store.List()
			if err != nil {
				store.Close()
				fmt.Printf("❌ 读取服务器列表失败: %v\n", err)
				os.Exit(1)
			}
			for _, srv := range servers {
				_, pwd, err := store.Get(srv.ID)
				if err != nil {
					continue
				}
				targets = append(targets, hostInfo{
					id:       srv.ID,
					host:     srv.Host,
					port:     22,
					user:     srv.User,
					password: pwd,
				})
			}
			store.Close()
			fmt.Printf("📦 从数据库加载 %d 台服务器\n\n", len(targets))
		} else if scpFromDB != "" {
			// 单机从数据库读取
			store, err := password.NewStore(scpDBPath, scpKey)
			if err != nil {
				fmt.Printf("❌ 打开数据库失败: %v\n", err)
				os.Exit(1)
			}
			srv, pwd, err := store.Get(scpFromDB)
			store.Close()
			if err != nil || srv == nil {
				fmt.Printf("❌ 未找到服务器: %s\n", scpFromDB)
				os.Exit(1)
			}
			targets = append(targets, hostInfo{
				id:       scpFromDB,
				host:     srv.Host,
				port:     22,
				user:     srv.User,
				password: pwd,
			})
			fmt.Printf("📦 从数据库加载: %s (%s@%s)\n\n", scpFromDB, srv.User, srv.Host)
		} else if len(scpHosts) > 0 {
			// 批量命令行指定
			for _, h := range scpHosts {
				targets = append(targets, hostInfo{
					host:     h,
					port:     scpBatchPort,
					user:     scpBatchUser,
					password: scpBatchPass,
				})
			}
		} else {
			// 单机命令行指定
			targets = append(targets, hostInfo{
				host:     scpHost,
				port:     scpPort,
				user:     scpUser,
				password: scpPassword,
				keyPath:  scpKeyPath,
				keyPass:  scpKeyPass,
			})
		}

		if len(targets) == 0 {
			fmt.Println("❌ 没有目标服务器，请指定 -H、-L 或 --from-db")
			os.Exit(1)
		}

		// 检查本地文件是否存在
		if _, err := os.Stat(localPath); os.IsNotExist(err) {
			fmt.Printf("❌ 本地文件不存在: %s\n", localPath)
			os.Exit(1)
		}

		fmt.Printf("📤 分发文件: %s → %s\n\n", localPath, scpRemote)

		var wg sync.WaitGroup
		results := make(chan string, len(targets))
		sem := make(chan struct{}, scpParallel)

		for _, t := range targets {
			wg.Add(1)
			go func(t hostInfo) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				label := t.host
				if t.id != "" {
					label = fmt.Sprintf("%s (%s)", t.id, t.host)
				}

				client := ssh.NewClient(t.host, t.port, t.user, t.password, t.keyPath, t.keyPass)
				if err := client.Connect(); err != nil {
					results <- fmt.Sprintf("❌ %s: 连接失败 - %v", label, err)
					return
				}
				defer client.Close()

				// 判断是文件还是目录
				info, _ := os.Stat(localPath)
				var err error
				if info.IsDir() {
					err = client.UploadDir(localPath, scpRemote)
				} else {
					err = client.UploadFile(localPath, scpRemote)
				}

				if err != nil {
					results <- fmt.Sprintf("❌ %s: 上传失败 - %v", label, err)
					return
				}
				results <- fmt.Sprintf("✅ %s: 上传成功", label)
			}(t)
		}

		wg.Wait()
		close(results)

		successCount := 0
		for result := range results {
			fmt.Println(result)
			if len(result) >= 3 && result[:3] == "✅" {
				successCount++
			}
		}
		fmt.Printf("\n完成: 成功 %d/%d 台\n", successCount, len(targets))
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
	execCmd.Flags().StringVar(&execFromDB, "from-db", "", "从数据库读取账密（指定server-id）")
	execCmd.Flags().StringVar(&execDBPath, "db", "passwords.db", "数据库路径")
	execCmd.Flags().StringVar(&execKey, "key", "", "主密钥（配合--from-db使用）")

	// batch 命令参数
	batchExecCmd.Flags().StringVarP(&batchConfigFile, "config", "c", "", "配置文件路径")
	batchExecCmd.Flags().StringSliceVarP(&batchHosts, "hosts", "L", []string{}, "服务器列表")
	batchExecCmd.Flags().IntVarP(&batchPort, "port", "T", 22, "SSH端口")
	batchExecCmd.Flags().StringVarP(&batchUser, "user", "U", "root", "用户名")
	batchExecCmd.Flags().StringVarP(&batchPassword, "password", "P", "", "密码")
	batchExecCmd.Flags().StringVarP(&batchKeyPath, "key", "K", "", "私钥路径")
	batchExecCmd.Flags().StringVar(&batchKeyPass, "key-pass", "", "私钥密码")
	batchExecCmd.Flags().IntVar(&batchParallel, "parallel", 10, "并发数")
	batchExecCmd.Flags().BoolVar(&batchFromDB, "from-db", false, "从数据库批量读取所有服务器账密")
	batchExecCmd.Flags().StringVar(&batchFromDBPath, "db", "passwords.db", "数据库路径")
	batchExecCmd.Flags().StringVar(&batchKey, "key", "", "主密钥（配合--from-db使用）")

	// scp 命令参数
	scpCmd.Flags().StringVarP(&scpHost, "host", "H", "", "服务器地址（单机）")
	scpCmd.Flags().IntVarP(&scpPort, "port", "P", 22, "SSH端口")
	scpCmd.Flags().StringVarP(&scpUser, "user", "u", "root", "用户名")
	scpCmd.Flags().StringVarP(&scpPassword, "password", "p", "", "密码")
	scpCmd.Flags().StringVarP(&scpKeyPath, "key", "i", "", "私钥路径")
	scpCmd.Flags().StringVar(&scpKeyPass, "key-pass", "", "私钥密码")
	scpCmd.Flags().StringVar(&scpFromDB, "from-db", "", "从数据库读取账密（指定server-id）")
	scpCmd.Flags().StringSliceVarP(&scpHosts, "hosts", "L", []string{}, "批量服务器列表")
	scpCmd.Flags().IntVar(&scpBatchPort, "batch-port", 22, "批量SSH端口")
	scpCmd.Flags().StringVar(&scpBatchUser, "batch-user", "root", "批量用户名")
	scpCmd.Flags().StringVar(&scpBatchPass, "batch-pass", "", "批量密码")
	scpCmd.Flags().BoolVar(&scpBatchDB, "all-from-db", false, "从数据库批量读取所有服务器")
	scpCmd.Flags().StringVarP(&scpRemote, "remote", "r", "", "远端目标路径")
	scpCmd.Flags().IntVar(&scpParallel, "parallel", 10, "并发数")
	scpCmd.Flags().StringVar(&scpDBPath, "db", "passwords.db", "数据库路径")
	scpCmd.Flags().StringVar(&scpKey, "key", "", "主密钥（配合--from-db使用）")
	scpCmd.MarkFlagRequired("remote")

	rootCmd.AddCommand(execCmd)
	rootCmd.AddCommand(batchExecCmd)
	rootCmd.AddCommand(scpCmd)
}

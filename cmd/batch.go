package main

import (
	"fmt"
	"sync"
	"ops-cli/pkg/ssh"
	"github.com/spf13/cobra"
)

var (
	hosts    []string
	parallel int
)

var batchExecCmd = &cobra.Command{
	Use:   "batch [command]",
	Short: "批量执行命令",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		command := args[0]
		
		var wg sync.WaitGroup
		results := make(chan string, len(hosts))
		
		for _, h := range hosts {
			wg.Add(1)
			go func(host string) {
				defer wg.Done()
				
				client := ssh.NewClient(host, port, user, password)
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

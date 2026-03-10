package password

import (
	"fmt"
	"ops-cli/pkg/ssh"
)

func ResetPassword(host string, port int, user, oldPassword, newPassword string) error {
	client := ssh.NewClient(host, port, user, oldPassword, "", "")
	
	if err := client.Connect(); err != nil {
		return fmt.Errorf("连接失败: %v", err)
	}
	defer client.Close()
	
	cmd := fmt.Sprintf("printf '%%s\\n' \"%s:%s\" | sudo /usr/sbin/chpasswd", user, newPassword)
	// fmt.Printf("  [DEBUG] 执行命令: %s\\n", cmd)
	
	output, err := client.Execute(cmd)
	// fmt.Printf("  [DEBUG] 命令输出: %s\\n", output)
	
	if err != nil {
		return fmt.Errorf("改密失败: %v (output: %s)", err, output)
	}
	
	return nil
}

func VerifyPassword(host string, port int, user, password string) error {
	client := ssh.NewClient(host, port, user, password, "", "")
	if err := client.Connect(); err != nil {
		return fmt.Errorf("验证失败: %v", err)
	}
	client.Close()
	return nil
}

func ResetPasswordVirsh(hypervisorHost string, hypervisorPort int, hypervisorUser, hypervisorPassword, hypervisorKey, hypervisorKeyPass string, instanceID, vmUser, newPassword string) error {
	client := ssh.NewClient(hypervisorHost, hypervisorPort, hypervisorUser, hypervisorPassword, hypervisorKey, hypervisorKeyPass)
	
	if err := client.Connect(); err != nil {
		return fmt.Errorf("连接物理机失败: %v", err)
	}
	defer client.Close()
	
	cmd := fmt.Sprintf("sudo virsh set-user-password %s %s %s", instanceID, vmUser, newPassword)
	_, err := client.Execute(cmd)
	if err != nil {
		return fmt.Errorf("virsh改密失败: %v", err)
	}
	
	return nil
}
package openstack

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type Client struct {
	AuthURL  string
	Username string
	Password string
	Project  string
	Token    string
	NovaURL  string
}

type VM struct {
	ID             string
	Name           string
	IP             string
	InstanceID     string
	HypervisorHost string
	User           string
}

func NewClient() *Client {
	return &Client{
		AuthURL:  os.Getenv("OS_AUTH_URL"),
		Username: os.Getenv("OS_USERNAME"),
		Password: os.Getenv("OS_PASSWORD"),
		Project:  os.Getenv("OS_PROJECT_NAME"),
	}
}
func (c *Client) Authenticate() error {
	authData := map[string]interface{}{
		"auth": map[string]interface{}{
			"identity": map[string]interface{}{
				"methods": []string{"password"},
				"password": map[string]interface{}{
					"user": map[string]interface{}{
						"name":     c.Username,
						"password": c.Password,
						"domain":   map[string]string{"name": "default"},
					},
				},
			},
			"scope": map[string]interface{}{
				"project": map[string]interface{}{
					"name":   c.Project,
					"domain": map[string]string{"name": "default"},
				},
			},
		},
	}

	body, _ := json.Marshal(authData)
	req, _ := http.NewRequest("POST", c.AuthURL+"/auth/tokens", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("认证失败 %d: %s", resp.StatusCode, string(bodyBytes))
	}

	c.Token = resp.Header.Get("X-Subject-Token")
	
	// 获取Nova URL
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	catalog := result["token"].(map[string]interface{})["catalog"].([]interface{})
	for _, svc := range catalog {
		service := svc.(map[string]interface{})
		if service["type"].(string) == "compute" {
			endpoints := service["endpoints"].([]interface{})
			for _, ep := range endpoints {
				endpoint := ep.(map[string]interface{})
				if endpoint["interface"].(string) == "public" {
					c.NovaURL = endpoint["url"].(string)
					break
				}
			}
		}
	}
	
	return nil
}

func (c *Client) ListVMs() ([]VM, error) {
	req, _ := http.NewRequest("GET", c.NovaURL+"/servers/detail", nil)
	req.Header.Set("X-Auth-Token", c.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("获取虚拟机列表失败 %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	servers := result["servers"].([]interface{})
	vms := []VM{}

	for _, srv := range servers {
		server := srv.(map[string]interface{})
		
		// 获取IP地址
		ip := ""
		addresses := server["addresses"].(map[string]interface{})
		for _, addrs := range addresses {
			addrList := addrs.([]interface{})
			if len(addrList) > 0 {
				ip = addrList[0].(map[string]interface{})["addr"].(string)
				break
			}
		}
		
		// 获取物理机
		hypervisor := ""
		if h, ok := server["OS-EXT-SRV-ATTR:hypervisor_hostname"]; ok {
			hypervisor = h.(string)
		}
		
		vm := VM{
			ID:             server["name"].(string),
			Name:           server["name"].(string),
			IP:             ip,
			InstanceID:     server["id"].(string),
			HypervisorHost: hypervisor,
			User:           "root",
		}
		vms = append(vms, vm)
	}

	return vms, nil
}

package password

import (
	"os"
	"testing"
	"time"
)

func TestNewStore(t *testing.T) {
	// 创建临时数据库
	tmpFile := "/tmp/test-passwords-" + time.Now().Format("20060102150405") + ".db"
	defer os.Remove(tmpFile)

	masterKey := "dGhpcyBpcyBhIDMyIGJ5dGUga2V5IGZvciB0ZXN0aW4="

	store, err := NewStore(tmpFile, masterKey)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	defer store.Close()

	// 验证数据库文件创建
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Errorf("数据库文件未创建")
	}
}

func TestStoreSaveAndGet(t *testing.T) {
	tmpFile := "/tmp/test-passwords-save-" + time.Now().Format("20060102150405") + ".db"
	defer os.Remove(tmpFile)

	masterKey := "dGhpcyBpcyBhIDMyIGJ5dGUga2V5IGZvciB0ZXN0aW4="

	store, err := NewStore(tmpFile, masterKey)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	defer store.Close()

	// 保存服务器
	srv := Server{
		ID:     "test-server-1",
		Name:   "测试服务器",
		Host:   "192.168.1.100",
		User:   "root",
	}
	password := "Abc123#xyz~test"

	err = store.Save(srv, password)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// 获取服务器
	gotSrv, gotPwd, err := store.Get("test-server-1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if gotSrv.ID != srv.ID {
		t.Errorf("Get() ID = %v, want %v", gotSrv.ID, srv.ID)
	}

	if gotPwd != password {
		t.Errorf("Get() password = %v, want %v", gotPwd, password)
	}
}

func TestStoreList(t *testing.T) {
	tmpFile := "/tmp/test-passwords-list-" + time.Now().Format("20060102150405") + ".db"
	defer os.Remove(tmpFile)

	masterKey := "dGhpcyBpcyBhIDMyIGJ5dGUga2V5IGZvciB0ZXN0aW4="

	store, err := NewStore(tmpFile, masterKey)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	defer store.Close()

	// 保存多个服务器
	servers := []struct {
		id   string
		host string
		pwd  string
	}{
		{"server-1", "192.168.1.1", "password1"},
		{"server-2", "192.168.1.2", "password2"},
		{"server-3", "192.168.1.3", "password3"},
	}

	for _, s := range servers {
		err := store.Save(Server{ID: s.id, Host: s.host}, s.pwd)
		if err != nil {
			t.Fatalf("Save() error = %v", err)
		}
	}

	// 获取列表
	list, err := store.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(list) != len(servers) {
		t.Errorf("List() length = %v, want %v", len(list), len(servers))
	}
}

func TestStoreDelete(t *testing.T) {
	tmpFile := "/tmp/test-passwords-delete-" + time.Now().Format("20060102150405") + ".db"
	defer os.Remove(tmpFile)

	masterKey := "dGhpcyBpcyBhIDMyIGJ5dGUga2V5IGZvciB0ZXN0aW4="

	store, err := NewStore(tmpFile, masterKey)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	defer store.Close()

	// 保存并删除
	srv := Server{ID: "to-delete", Host: "192.168.1.100"}
	err = store.Save(srv, "password")
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	err = store.Delete("to-delete")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// 验证删除
	_, _, err = store.Get("to-delete")
	if err == nil {
		t.Errorf("删除后仍能获取服务器")
	}
}

func TestStoreUpdate(t *testing.T) {
	tmpFile := "/tmp/test-passwords-update-" + time.Now().Format("20060102150405") + ".db"
	defer os.Remove(tmpFile)

	masterKey := "dGhpcyBpcyBhIDMyIGJ5dGUga2V5IGZvciB0ZXN0aW4="

	store, err := NewStore(tmpFile, masterKey)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	defer store.Close()

	// 保存服务器
	srv := Server{ID: "to-update", Host: "192.168.1.100", User: "root"}
	err = store.Save(srv, "old-password")
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// 更新
	updates := map[string]interface{}{
		"name": "新名称",
		"host": "192.168.1.200",
		"user": "admin",
	}

	err = store.Update("to-update", updates)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// 验证更新
	gotSrv, _, err := store.Get("to-update")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if gotSrv.Name != "新名称" {
		t.Errorf("Update() name = %v, want 新名称", gotSrv.Name)
	}

	if gotSrv.Host != "192.168.1.200" {
		t.Errorf("Update() host = %v, want 192.168.1.200", gotSrv.Host)
	}

	if gotSrv.User != "admin" {
		t.Errorf("Update() user = %v, want admin", gotSrv.User)
	}
}

func TestStoreWithWrongMasterKey(t *testing.T) {
	tmpFile := "/tmp/test-passwords-wrongkey-" + time.Now().Format("20060102150405") + ".db"
	defer os.Remove(tmpFile)

	correctKey := "dGhpcyBpcyBhIDMyIGJ5dGUga2V5IGZvciB0ZXN0aW4="
	wrongKey := "YW5vdGhlciAzMiBieXRlIGtleSBmb3IgdGVzdGluZyE="

	// 使用正确密钥创建并保存
	store, err := NewStore(tmpFile, correctKey)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	srv := Server{ID: "test", Host: "192.168.1.100"}
	err = store.Save(srv, "password")
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	store.Close()

	// 使用错误密钥打开
	_, err = NewStore(tmpFile, wrongKey)
	if err == nil {
		t.Errorf("使用错误密钥成功打开数据库，应该失败")
	}
}

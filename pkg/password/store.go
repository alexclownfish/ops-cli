package password

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
	"go.etcd.io/bbolt"
)

type Server struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Host              string    `json:"host"`
	User              string    `json:"user"`
	PasswordEncrypted string    `json:"password_encrypted"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	
	// virsh相关
	InstanceID       string `json:"instance_id,omitempty"`
	HypervisorHost   string `json:"hypervisor_host,omitempty"`
	HypervisorPort   int    `json:"hypervisor_port,omitempty"`
	HypervisorUser   string `json:"hypervisor_user,omitempty"`
	HypervisorPass    string `json:"hypervisor_pass,omitempty"`
	HypervisorKey     string `json:"hypervisor_key,omitempty"`
	HypervisorKeyPass string `json:"hypervisor_key_pass,omitempty"`
	ResetMethod      string `json:"reset_method,omitempty"` // ssh | virsh
}

type Store struct {
	db        *bbolt.DB
	masterKey string
}

func NewStore(dbPath, masterKey string) (*Store, error) {
	db, err := bbolt.Open(dbPath, 0600, nil)
	if err != nil {
		return nil, err
	}
	
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("servers"))
		return err
	})
	
	if err != nil {
		db.Close()
		return nil, err
	}
	
	store := &Store{db: db, masterKey: masterKey}
	
	// 验证主密钥
	if err := store.InitOrVerifyKey(); err != nil {
		db.Close()
		return nil, err
	}
	
	return store, nil
}
func (s *Store) Save(srv Server, plainPassword string) error {
	encrypted, err := Encrypt(plainPassword, s.masterKey)
	if err != nil {
		return err
	}
	
	srv.PasswordEncrypted = encrypted
	srv.UpdatedAt = time.Now()
	if srv.CreatedAt.IsZero() {
		srv.CreatedAt = time.Now()
	}
	
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("servers"))
		data, err := json.Marshal(srv)
		if err != nil {
			return err
		}
		return b.Put([]byte(srv.ID), data)
	})
}

func (s *Store) Get(id string) (*Server, string, error) {
	var srv Server
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("servers"))
		data := b.Get([]byte(id))
		if data == nil {
			return nil
		}
		return json.Unmarshal(data, &srv)
	})
	
	if err != nil {
		return nil, "", err
	}
	
	plainPassword, err := Decrypt(srv.PasswordEncrypted, s.masterKey)
	if err != nil {
		return nil, "", err
	}
	
	return &srv, plainPassword, nil
}

func (s *Store) List() ([]Server, error) {
	var servers []Server
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("servers"))
		return b.ForEach(func(k, v []byte) error {
			var srv Server
			if err := json.Unmarshal(v, &srv); err != nil {
				return err
			}
			servers = append(servers, srv)
			return nil
		})
	})
	return servers, err
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Delete(id string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("servers"))
		return b.Delete([]byte(id))
	})
}

// 初始化或验证主密钥
func (s *Store) InitOrVerifyKey() error {
	keyHash := sha256.Sum256([]byte(s.masterKey))
	keyHashStr := hex.EncodeToString(keyHash[:])
	
	return s.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("key_hash"))
		if err != nil {
			return err
		}
		
		stored := b.Get([]byte("hash"))
		if stored == nil {
			// 首次使用，存储哈希
			return b.Put([]byte("hash"), []byte(keyHashStr))
		}
		
		// 验证哈希
		if string(stored) != keyHashStr {
			return fmt.Errorf("主密钥错误")
		}
		
		return nil
	})
}
// Update 更新服务器信息
func (s *Store) Update(id string, updates map[string]interface{}) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("servers"))
		data := b.Get([]byte(id))
		if data == nil {
			return fmt.Errorf("服务器不存在: %s", id)
		}
		
		var srv Server
		if err := json.Unmarshal(data, &srv); err != nil {
			return err
		}
		
		// 更新字段
		for key, value := range updates {
			switch key {
			case "name":
				srv.Name = value.(string)
			case "host":
				srv.Host = value.(string)
			case "user":
				srv.User = value.(string)
			case "password":
				encrypted, err := Encrypt(value.(string), s.masterKey)
				if err != nil {
					return err
				}
				srv.PasswordEncrypted = encrypted
			case "instance_id":
				srv.InstanceID = value.(string)
			case "hypervisor_host":
				srv.HypervisorHost = value.(string)
			case "hypervisor_port":
				srv.HypervisorPort = value.(int)
			case "hypervisor_user":
				srv.HypervisorUser = value.(string)
			}
		}
		
		srv.UpdatedAt = time.Now()
		
		newData, err := json.Marshal(srv)
		if err != nil {
			return err
		}
		return b.Put([]byte(srv.ID), newData)
	})
}

package password

import (
	"encoding/json"
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
	HypervisorPass   string `json:"hypervisor_pass,omitempty"`
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
	
	return &Store{db: db, masterKey: masterKey}, nil
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

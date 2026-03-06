package password

import (
	"os"
	"time"
	"github.com/tobischo/gokeepasslib/v3"
	"github.com/tobischo/gokeepasslib/v3/wrappers"
)

func ExportToKeePass(servers []Server, masterKey, outputPath, kdbxPassword string) error {
	// 创建新数据库
	db := gokeepasslib.NewDatabase()
	
	// 创建根组
	rootGroup := gokeepasslib.NewGroup()
	rootGroup.Name = "Root"
	rootGroup.UUID = gokeepasslib.NewUUID()
	
	// 创建密码组
	passwordGroup := gokeepasslib.NewGroup()
	passwordGroup.Name = "ops-cli-passwords"
	passwordGroup.UUID = gokeepasslib.NewUUID()
	
	// 添加条目
	for _, srv := range servers {
		plainPwd, err := Decrypt(srv.PasswordEncrypted, masterKey)
		if err != nil {
			continue
		}
		
		entry := gokeepasslib.NewEntry()
		entry.UUID = gokeepasslib.NewUUID()
		entry.Values = []gokeepasslib.ValueData{
			{Key: "Title", Value: gokeepasslib.V{Content: srv.Name}},
			{Key: "UserName", Value: gokeepasslib.V{Content: srv.User}},
			{Key: "Password", Value: gokeepasslib.V{Content: plainPwd}},
			{Key: "URL", Value: gokeepasslib.V{Content: "ssh://" + srv.Host}},
			{Key: "Notes", Value: gokeepasslib.V{Content: srv.ID}},
		}
		passwordGroup.Entries = append(passwordGroup.Entries, entry)
	}
	
	// 将密码组添加到根组
	rootGroup.Groups = append(rootGroup.Groups, passwordGroup)
	
	// 将根组添加到数据库
	db.Content.Root.Groups = []gokeepasslib.Group{rootGroup}
	
	// 设置元数据
	db.Content.Meta.DatabaseName = "ops-cli passwords"
	db.Content.Meta.DatabaseNameChanged = &wrappers.TimeWrapper{Time: time.Now()}
	
	// 设置凭据
	db.Credentials = gokeepasslib.NewPasswordCredentials(kdbxPassword)
	
	// 写入文件
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	encoder := gokeepasslib.NewEncoder(file)
	return encoder.Encode(db)
}
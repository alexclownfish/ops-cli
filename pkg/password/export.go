package password

import (
	"os"
	"github.com/tobischo/gokeepasslib/v3"
	"github.com/tobischo/gokeepasslib/v3/wrappers"
)

func ExportToKeePass(servers []Server, masterKey, outputPath, kdbxPassword string) error {
	// 创建新数据库
	db := gokeepasslib.NewDatabase()
	
	// 创建密码组
	rootGroup := gokeepasslib.NewGroup()
	rootGroup.Name = "ops-cli-passwords"
	
	// 添加条目
	for _, srv := range servers {
		plainPwd, err := Decrypt(srv.PasswordEncrypted, masterKey)
		if err != nil {
			continue
		}
		
		entry := gokeepasslib.NewEntry()
		entry.Values = []gokeepasslib.ValueData{
			{Key: "Title", Value: gokeepasslib.V{Content: srv.Name}},
			{Key: "UserName", Value: gokeepasslib.V{Content: srv.User}},
			{Key: "Password", Value: gokeepasslib.V{Content: plainPwd, Protected: wrappers.NewBoolWrapper(true)}},
			{Key: "URL", Value: gokeepasslib.V{Content: "ssh://" + srv.Host}},
			{Key: "Notes", Value: gokeepasslib.V{Content: srv.ID}},
		}
		rootGroup.Entries = append(rootGroup.Entries, entry)
	}
	
	// 添加到数据库
	db.Content.Root.Groups = append(db.Content.Root.Groups, rootGroup)
	
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
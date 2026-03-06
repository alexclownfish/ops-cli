package password

import (
	"os"
	"github.com/tobischo/gokeepasslib/v3"
	"github.com/tobischo/gokeepasslib/v3/wrappers"
)

func ExportToKeePass(servers []Server, masterKey, outputPath, kdbxPassword string) error {
	db := gokeepasslib.NewDatabase()
	
	rootGroup := gokeepasslib.NewGroup()
	rootGroup.Name = "ops-cli-passwords"
	
	for _, srv := range servers {
		plainPwd, err := Decrypt(srv.PasswordEncrypted, masterKey)
		if err != nil {
			continue
		}
		
		entry := gokeepasslib.NewEntry()
		entry.Values = append(entry.Values,
			gokeepasslib.ValueData{Key: "Title", Value: gokeepasslib.V{Content: srv.Name}},
			gokeepasslib.ValueData{Key: "UserName", Value: gokeepasslib.V{Content: srv.User}},
			gokeepasslib.ValueData{Key: "Password", Value: gokeepasslib.V{Content: plainPwd, Protected: wrappers.NewBoolWrapper(true)}},
			gokeepasslib.ValueData{Key: "URL", Value: gokeepasslib.V{Content: "ssh://" + srv.Host}},
			gokeepasslib.ValueData{Key: "Notes", Value: gokeepasslib.V{Content: srv.ID}},
		)
		rootGroup.Entries = append(rootGroup.Entries, entry)
	}
	
	db.Content.Root.Groups = append(db.Content.Root.Groups, rootGroup)
	db.Credentials = gokeepasslib.NewPasswordCredentials(kdbxPassword)
	
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	encoder := gokeepasslib.NewEncoder(file)
	return encoder.Encode(db)
}